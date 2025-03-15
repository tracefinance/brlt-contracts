package blockexplorer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"vault0/internal/types"
)

const (
	// Default maximum requests per second for API rate limiting
	defaultMaxRequestsPerSecond = 5
	// Default timeout for HTTP requests
	defaultRequestTimeout = 10 * time.Second
	// Default page size for paginated requests
	defaultPageSize = 100
)

// EVMExplorer implements BlockExplorer for Etherscan-compatible APIs
type EVMExplorer struct {
	baseURL     string
	apiKey      string
	httpClient  *http.Client
	chain       types.Chain
	rateLimiter *time.Ticker
	mu          sync.Mutex
}

// EVMTransactionResponse represents the response structure from Etherscan-like APIs
type EVMTransactionResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
}

// EVMTransaction represents a transaction returned by Etherscan-like APIs
type EVMTransaction struct {
	BlockNumber       string `json:"blockNumber"`
	TimeStamp         string `json:"timeStamp"`
	Hash              string `json:"hash"`
	Nonce             string `json:"nonce"`
	BlockHash         string `json:"blockHash"`
	From              string `json:"from"`
	ContractAddress   string `json:"contractAddress"`
	To                string `json:"to"`
	Value             string `json:"value"`
	TokenName         string `json:"tokenName,omitempty"`
	TokenSymbol       string `json:"tokenSymbol,omitempty"`
	TokenDecimal      string `json:"tokenDecimal,omitempty"`
	TransactionIndex  string `json:"transactionIndex"`
	Gas               string `json:"gas"`
	GasPrice          string `json:"gasPrice"`
	GasUsed           string `json:"gasUsed"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	Input             string `json:"input"`
	Confirmations     string `json:"confirmations"`
	IsError           string `json:"isError,omitempty"`
}

// EVMTokenBalance represents token balance information specific to EVM-based blockchains.
// This is an implementation-specific type that gets converted to the abstract TokenBalance.
type EVMTokenBalance struct {
	TokenAddress string `json:"contractAddress"`
	TokenName    string `json:"tokenName"`
	TokenSymbol  string `json:"tokenSymbol"`
	// TokenDecimal is a string in the API response, converted to uint8 when mapping to TokenBalance
	TokenDecimal string `json:"tokenDecimal"`
	// Balance is a string in the API response, converted to *big.Int when mapping to TokenBalance
	Balance string `json:"balance"`
}

// EVMContractInfo represents contract information specific to EVM-based blockchains.
// This is an implementation-specific type that gets converted to the abstract ContractInfo.
type EVMContractInfo struct {
	ABI              string `json:"ABI"`
	ContractName     string `json:"ContractName"`
	CompilerVersion  string `json:"CompilerVersion"`
	OptimizationUsed string `json:"OptimizationUsed"`
	SourceCode       string `json:"SourceCode"`
	// IsVerified is derived from the presence of source code
	IsVerified bool `json:"-"`
}

// EVMExplorerError represents an error that occurred during an EVM explorer operation
// and includes the raw response from the API for debugging purposes.
type EVMExplorerError struct {
	Err         error
	RawResponse string
}

// Error implements the error interface.
func (e *EVMExplorerError) Error() string {
	return fmt.Sprintf("%v (raw response: %s)", e.Err, e.RawResponse)
}

// Unwrap returns the underlying error.
func (e *EVMExplorerError) Unwrap() error {
	return e.Err
}

// NewEVMExplorer creates a new EVMExplorer instance
func NewEVMExplorer(chain types.Chain, baseURL, apiKey string) (*EVMExplorer, error) {
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	// Ensure the base URL does not end with a slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	// Change from UI URL to API URL if needed
	if !strings.Contains(baseURL, "api") {
		baseURL = strings.Replace(baseURL, "//", "//api.", 1)
	}

	return &EVMExplorer{
		baseURL:     baseURL,
		apiKey:      apiKey,
		httpClient:  &http.Client{Timeout: defaultRequestTimeout},
		chain:       chain,
		rateLimiter: time.NewTicker(time.Second / defaultMaxRequestsPerSecond),
	}, nil
}

// GetContract checks if a contract is verified and returns its information
func (e *EVMExplorer) GetContract(ctx context.Context, address string) (*ContractInfo, error) {
	if err := e.chain.ValidateAddress(address); err != nil {
		return nil, &EVMExplorerError{Err: ErrInvalidAddress, RawResponse: ""}
	}

	// Prepare request parameters
	params := url.Values{}
	params.Add("module", "contract")
	params.Add("action", "getsourcecode")
	params.Add("address", address)
	params.Add("apikey", e.apiKey)

	var response EVMTransactionResponse
	rawResponse, err := e.aaa(ctx, params, &response)
	if err != nil {
		return nil, &EVMExplorerError{Err: err, RawResponse: rawResponse}
	}

	var evmResult []EVMContractInfo
	if err := json.Unmarshal(response.Result, &evmResult); err != nil {
		return nil, &EVMExplorerError{Err: fmt.Errorf("%w: %v", ErrInvalidResponse, err), RawResponse: rawResponse}
	}

	if len(evmResult) == 0 {
		return nil, &EVMExplorerError{Err: fmt.Errorf("%w: no contract info returned", ErrInvalidResponse), RawResponse: rawResponse}
	}

	// Convert EVMContractInfo to ContractInfo
	evmInfo := evmResult[0]

	return &ContractInfo{
		ABI:          evmInfo.ABI,
		ContractName: evmInfo.ContractName,
		SourceCode:   evmInfo.SourceCode,
		IsVerified:   evmInfo.SourceCode != "",
	}, nil
}

// GetTransactionHistory retrieves transaction history for an address
func (e *EVMExplorer) GetTransactionHistory(ctx context.Context, address string, options TransactionHistoryOptions) ([]*types.Transaction, error) {
	if err := e.chain.ValidateAddress(address); err != nil {
		return nil, ErrInvalidAddress
	}

	// Set default values if not provided
	if options.PageSize <= 0 {
		options.PageSize = defaultPageSize
	}

	if options.Page <= 0 {
		options.Page = 1
	}

	// Use the latest blocks if not specified
	startBlock := "0"
	if options.StartBlock > 0 {
		startBlock = strconv.FormatInt(options.StartBlock, 10)
	}

	endBlock := "999999999" // A large number to represent latest block
	if options.EndBlock > 0 {
		endBlock = strconv.FormatInt(options.EndBlock, 10)
	}

	// Determine which transaction types to fetch
	var txTypes []TransactionType
	if len(options.TransactionTypes) > 0 {
		txTypes = options.TransactionTypes
	} else {
		// Default to all transaction types
		txTypes = []TransactionType{TxTypeNormal, TxTypeInternal, TxTypeERC20, TxTypeERC721}
	}

	// Fetch transactions of each type in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex
	allTxs := []*types.Transaction{}
	errChan := make(chan error, len(txTypes))

	for _, txType := range txTypes {
		wg.Add(1)
		go func(txType TransactionType) {
			defer wg.Done()

			var txs []*types.Transaction
			var err error

			switch txType {
			case TxTypeNormal:
				txs, err = e.getNormalTransactions(ctx, address, startBlock, endBlock, options.Page, options.PageSize, options.SortAscending)
			case TxTypeInternal:
				txs, err = e.getInternalTransactions(ctx, address, startBlock, endBlock, options.Page, options.PageSize, options.SortAscending)
			case TxTypeERC20:
				txs, err = e.getTokenTransactions(ctx, address, "tokentx", startBlock, endBlock, options.Page, options.PageSize, options.SortAscending)
			case TxTypeERC721:
				txs, err = e.getTokenTransactions(ctx, address, "token721tx", startBlock, endBlock, options.Page, options.PageSize, options.SortAscending)
			}

			if err != nil {
				errChan <- err
				return
			}

			mu.Lock()
			allTxs = append(allTxs, txs...)
			mu.Unlock()
		}(txType)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	return allTxs, nil
}

// GetTransactionsByHash retrieves transaction details for multiple hashes
func (e *EVMExplorer) GetTransactionsByHash(ctx context.Context, hashes []string) ([]*types.Transaction, error) {
	if len(hashes) == 0 {
		return []*types.Transaction{}, nil
	}

	var wg sync.WaitGroup
	txs := make([]*types.Transaction, len(hashes))
	errChan := make(chan error, len(hashes))

	for i, hash := range hashes {
		wg.Add(1)
		go func(i int, hash string) {
			defer wg.Done()

			// Ensure hash has 0x prefix
			if !strings.HasPrefix(hash, "0x") {
				hash = "0x" + hash
			}

			// Prepare request parameters
			params := url.Values{}
			params.Add("module", "proxy")
			params.Add("action", "eth_getTransactionByHash")
			params.Add("txhash", hash)
			params.Add("apikey", e.apiKey)

			var txResponse struct {
				Jsonrpc string `json:"jsonrpc"`
				Id      int    `json:"id"`
				Result  struct {
					Hash             string `json:"hash"`
					From             string `json:"from"`
					To               string `json:"to"`
					Value            string `json:"value"`
					Gas              string `json:"gas"`
					GasPrice         string `json:"gasPrice"`
					Nonce            string `json:"nonce"`
					Input            string `json:"input"`
					BlockHash        string `json:"blockHash"`
					BlockNumber      string `json:"blockNumber"`
					TransactionIndex string `json:"transactionIndex"`
				} `json:"result"`
			}

			// Make request
			err := e.makeRequest(ctx, params, &txResponse)
			if err != nil {
				errChan <- err
				return
			}

			// Get receipt for status and additional information
			receiptParams := url.Values{}
			receiptParams.Add("module", "proxy")
			receiptParams.Add("action", "eth_getTransactionReceipt")
			receiptParams.Add("txhash", hash)
			receiptParams.Add("apikey", e.apiKey)

			var receiptResponse struct {
				Jsonrpc string `json:"jsonrpc"`
				Id      int    `json:"id"`
				Result  struct {
					BlockNumber       string `json:"blockNumber"`
					Status            string `json:"status"`
					GasUsed           string `json:"gasUsed"`
					CumulativeGasUsed string `json:"cumulativeGasUsed"`
				} `json:"result"`
			}

			err = e.makeRequest(ctx, receiptParams, &receiptResponse)
			if err != nil {
				errChan <- err
				return
			}

			// Convert hex values to integers
			blockNumber, _ := strconv.ParseInt(txResponse.Result.BlockNumber, 0, 64)
			value := new(big.Int)
			value.SetString(txResponse.Result.Value[2:], 16) // Remove 0x prefix
			gasPrice := new(big.Int)
			gasPrice.SetString(txResponse.Result.GasPrice[2:], 16)
			gasLimit, _ := strconv.ParseUint(txResponse.Result.Gas[2:], 16, 64)
			nonce, _ := strconv.ParseUint(txResponse.Result.Nonce[2:], 16, 64)

			// Get status
			status := "pending"
			if receiptResponse.Result.Status == "0x1" {
				status = "success"
			} else if receiptResponse.Result.Status == "0x0" {
				status = "failed"
			}

			// Get timestamp from block
			var timestamp int64
			if blockNumber > 0 {
				blockParams := url.Values{}
				blockParams.Add("module", "proxy")
				blockParams.Add("action", "eth_getBlockByNumber")
				blockParams.Add("tag", txResponse.Result.BlockNumber)
				blockParams.Add("boolean", "false")
				blockParams.Add("apikey", e.apiKey)

				var blockResponse struct {
					Result struct {
						Timestamp string `json:"timestamp"`
					} `json:"result"`
				}

				if err := e.makeRequest(ctx, blockParams, &blockResponse); err == nil {
					timestampHex, _ := strconv.ParseInt(blockResponse.Result.Timestamp[2:], 16, 64)
					timestamp = timestampHex
				}
			}

			// Create Transaction object
			txs[i] = &types.Transaction{
				Chain:     e.chain.Type,
				Hash:      txResponse.Result.Hash,
				From:      txResponse.Result.From,
				To:        txResponse.Result.To,
				Value:     value,
				Data:      []byte(txResponse.Result.Input),
				Nonce:     nonce,
				GasPrice:  gasPrice,
				GasLimit:  gasLimit,
				Type:      types.TransactionTypeNative,
				Status:    status,
				Timestamp: timestamp,
			}
		}(i, hash)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	return txs, nil
}

// GetAddressBalance retrieves the balance for an address
func (e *EVMExplorer) GetAddressBalance(ctx context.Context, address string) (*big.Int, error) {
	if err := e.chain.ValidateAddress(address); err != nil {
		return nil, ErrInvalidAddress
	}

	// Prepare request parameters
	params := url.Values{}
	params.Add("module", "account")
	params.Add("action", "balance")
	params.Add("address", address)
	params.Add("tag", "latest")
	params.Add("apikey", e.apiKey)

	var response EVMTransactionResponse
	if err := e.makeRequest(ctx, params, &response); err != nil {
		return nil, err
	}

	// Extract balance as string
	balanceStr := string(response.Result)
	if balanceStr == "" {
		return big.NewInt(0), nil
	}

	// Convert string to big.Int
	balance := new(big.Int)
	balance.SetString(strings.Trim(balanceStr, "\""), 10)
	return balance, nil
}

// GetTokenBalances retrieves token balances for an address
func (e *EVMExplorer) GetTokenBalances(ctx context.Context, address string) ([]*TokenBalance, error) {
	if err := e.chain.ValidateAddress(address); err != nil {
		return nil, ErrInvalidAddress
	}

	// Prepare request parameters
	params := url.Values{}
	params.Add("module", "account")
	params.Add("action", "tokenbalance")
	params.Add("address", address)
	params.Add("apikey", e.apiKey)

	var response EVMTransactionResponse
	if err := e.makeRequest(ctx, params, &response); err != nil {
		return nil, err
	}

	var evmBalances []EVMTokenBalance
	if err := json.Unmarshal(response.Result, &evmBalances); err != nil {
		// If the result is not an array, it might be a single token balance or not supported
		return []*TokenBalance{}, nil
	}

	// Convert EVMTokenBalance to TokenBalance
	balances := make([]*TokenBalance, len(evmBalances))
	for i, evmBalance := range evmBalances {
		decimal, _ := strconv.ParseUint(evmBalance.TokenDecimal, 10, 8)
		balance := new(big.Int)
		balance.SetString(evmBalance.Balance, 10)

		balances[i] = &TokenBalance{
			TokenAddress: evmBalance.TokenAddress,
			TokenName:    evmBalance.TokenName,
			TokenSymbol:  evmBalance.TokenSymbol,
			TokenDecimal: uint8(decimal),
			Balance:      balance,
		}
	}

	return balances, nil
}

// Close implements BlockExplorer.Close
func (e *EVMExplorer) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.rateLimiter != nil {
		e.rateLimiter.Stop()
		e.rateLimiter = nil
	}
	return nil
}

// Chain implements BlockExplorer.Chain
func (e *EVMExplorer) Chain() types.Chain {
	return e.chain
}

// GetTokenURL returns the URL to view the token on the block explorer.
func (e *EVMExplorer) GetTokenURL(address string) string {
	// Convert API URL to UI URL
	uiURL := strings.Replace(e.baseURL, "//api.", "//", 1)
	return fmt.Sprintf("%s/token/%s", uiURL, address)
}

// makeRequest makes a rate-limited request to the API
func (e *EVMExplorer) makeRequest(ctx context.Context, params url.Values, result interface{}) error {
	select {
	case <-e.rateLimiter.C:
		// We can proceed with the request
	case <-ctx.Done():
		return ctx.Err()
	}

	// Construct the URL
	reqURL := fmt.Sprintf("%s/api?%s", e.baseURL, params.Encode())

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	// Make the request
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode == http.StatusTooManyRequests {
		return ErrRateLimitExceeded
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status code %d", ErrRequestFailed, resp.StatusCode)
	}

	// Parse the response
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	return nil
}

// getNormalTransactions fetches normal transactions for an address
func (e *EVMExplorer) getNormalTransactions(ctx context.Context, address, startBlock, endBlock string, page, pageSize int, sortAsc bool) ([]*types.Transaction, error) {
	// Prepare request parameters
	params := url.Values{}
	params.Add("module", "account")
	params.Add("action", "txlist")
	params.Add("address", address)
	params.Add("startblock", startBlock)
	params.Add("endblock", endBlock)
	params.Add("page", strconv.Itoa(page))
	params.Add("offset", strconv.Itoa(pageSize))

	// Set sort order
	sortOrder := "desc"
	if sortAsc {
		sortOrder = "asc"
	}
	params.Add("sort", sortOrder)

	params.Add("apikey", e.apiKey)

	var response EVMTransactionResponse
	if err := e.makeRequest(ctx, params, &response); err != nil {
		return nil, err
	}

	// Check if the result is empty
	if string(response.Result) == "[]" || string(response.Result) == "\"No transactions found\"" {
		return []*types.Transaction{}, nil
	}

	// Parse transactions
	var evmTxs []EVMTransaction
	if err := json.Unmarshal(response.Result, &evmTxs); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	// Convert to our transaction model
	return e.convertEVMTransactionsToTransactions(evmTxs, types.TransactionTypeNative), nil
}

// getInternalTransactions fetches internal transactions for an address
func (e *EVMExplorer) getInternalTransactions(ctx context.Context, address, startBlock, endBlock string, page, pageSize int, sortAsc bool) ([]*types.Transaction, error) {
	// Prepare request parameters
	params := url.Values{}
	params.Add("module", "account")
	params.Add("action", "txlistinternal")
	params.Add("address", address)
	params.Add("startblock", startBlock)
	params.Add("endblock", endBlock)
	params.Add("page", strconv.Itoa(page))
	params.Add("offset", strconv.Itoa(pageSize))

	// Set sort order
	sortOrder := "desc"
	if sortAsc {
		sortOrder = "asc"
	}
	params.Add("sort", sortOrder)

	params.Add("apikey", e.apiKey)

	var response EVMTransactionResponse
	if err := e.makeRequest(ctx, params, &response); err != nil {
		return nil, err
	}

	// Check if the result is empty
	if string(response.Result) == "[]" || string(response.Result) == "\"No transactions found\"" {
		return []*types.Transaction{}, nil
	}

	// Parse transactions
	var evmTxs []EVMTransaction
	if err := json.Unmarshal(response.Result, &evmTxs); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	// Convert to our transaction model
	return e.convertEVMTransactionsToTransactions(evmTxs, types.TransactionTypeContract), nil
}

// getTokenTransactions fetches token transactions for an address
func (e *EVMExplorer) getTokenTransactions(ctx context.Context, address, action, startBlock, endBlock string, page, pageSize int, sortAsc bool) ([]*types.Transaction, error) {
	// Prepare request parameters
	params := url.Values{}
	params.Add("module", "account")
	params.Add("action", action)
	params.Add("address", address)
	params.Add("startblock", startBlock)
	params.Add("endblock", endBlock)
	params.Add("page", strconv.Itoa(page))
	params.Add("offset", strconv.Itoa(pageSize))

	// Set sort order
	sortOrder := "desc"
	if sortAsc {
		sortOrder = "asc"
	}
	params.Add("sort", sortOrder)

	params.Add("apikey", e.apiKey)

	var response EVMTransactionResponse
	if err := e.makeRequest(ctx, params, &response); err != nil {
		return nil, err
	}

	// Check if the result is empty
	if string(response.Result) == "[]" || string(response.Result) == "\"No transactions found\"" {
		return []*types.Transaction{}, nil
	}

	// Parse transactions
	var evmTxs []EVMTransaction
	if err := json.Unmarshal(response.Result, &evmTxs); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	// Determine transaction type based on action
	txType := types.TransactionTypeERC20
	if action == "token721tx" {
		txType = types.TransactionTypeERC20 // Using ERC20 type for now, could add NFT specific type later
	}

	// Convert to our transaction model
	return e.convertEVMTransactionsToTransactions(evmTxs, txType), nil
}

// convertEVMTransactionsToTransactions converts EVMTransaction slices to types.Transaction slices
func (e *EVMExplorer) convertEVMTransactionsToTransactions(evmTxs []EVMTransaction, txType types.TransactionType) []*types.Transaction {
	txs := make([]*types.Transaction, len(evmTxs))

	for i, evmTx := range evmTxs {
		// Convert string values to appropriate types
		timestamp, _ := strconv.ParseInt(evmTx.TimeStamp, 10, 64)
		nonce, _ := strconv.ParseUint(evmTx.Nonce, 10, 64)
		gasLimit, _ := strconv.ParseUint(evmTx.Gas, 10, 64)

		// Convert value to big.Int
		value := new(big.Int)
		value.SetString(evmTx.Value, 10)

		// Convert gas price to big.Int
		gasPrice := new(big.Int)
		gasPrice.SetString(evmTx.GasPrice, 10)

		// Determine transaction status
		status := "success"
		if evmTx.IsError == "1" {
			status = "failed"
		}

		// Create the transaction object
		txs[i] = &types.Transaction{
			Chain:        e.chain.Type,
			Hash:         evmTx.Hash,
			From:         evmTx.From,
			To:           evmTx.To,
			Value:        value,
			Data:         []byte(evmTx.Input),
			Nonce:        nonce,
			GasPrice:     gasPrice,
			GasLimit:     gasLimit,
			Type:         txType,
			TokenAddress: evmTx.ContractAddress,
			Status:       status,
			Timestamp:    timestamp,
		}
	}

	return txs
}

// makeRequestWithRawResponse makes an HTTP request and returns both the parsed response and raw response body
func (e *EVMExplorer) aaa(ctx context.Context, params url.Values, response interface{}) (string, error) {
	maxRetries := 3
	baseDelay := time.Second

	var lastErr error
	var body string
	var rawResp EVMTransactionResponse
	var bodyBytes []byte

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Rate limiting
		select {
		case <-e.rateLimiter.C:
		case <-ctx.Done():
			return "", ctx.Err()
		}

		// Build request URL
		reqURL := fmt.Sprintf("%s/api?%s", e.baseURL, params.Encode())

		// Create request
		req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}

		// Make request
		resp, err := e.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to make request: %w", err)
			goto retry
		}
		defer resp.Body.Close()

		// Read response body
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			goto retry
		}
		body = string(bodyBytes)

		// Check for rate limit response
		if err := json.Unmarshal(bodyBytes, &rawResp); err != nil {
			lastErr = fmt.Errorf("%w: %v", ErrInvalidResponse, err)
			goto retry
		}

		if rawResp.Status == "0" && strings.Contains(string(rawResp.Result), "rate limit") {
			lastErr = fmt.Errorf("rate limit exceeded")
			goto retry
		}

		// Parse final response
		if err := json.Unmarshal(bodyBytes, response); err != nil {
			return body, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
		}

		return body, nil

	retry:
		if attempt < maxRetries-1 {
			// Exponential backoff
			delay := baseDelay * time.Duration(1<<uint(attempt))
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return "", ctx.Err()
			}
			continue
		}
	}

	return body, lastErr
}
