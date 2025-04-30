package blockexplorer

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"

	"golang.org/x/time/rate"
)

// NextPage represents the pagination state for Etherscan API queries.
// It is encoded as a base64 string in the NextToken field of the Page response.
type NextPage struct {
	Page int `json:"page"`
}

// Encode serializes the NextPage struct to a base64 string
func (np *NextPage) Encode() string {
	data, err := json.Marshal(np)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeNextPage decodes a base64 string into a NextPage struct
func DecodeNextPage(token string) (*NextPage, error) {
	if token == "" {
		return &NextPage{Page: 1}, nil
	}

	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return nil, errors.NewTokenDecodingFailedError(token, err)
	}

	var np NextPage
	if err := json.Unmarshal(data, &np); err != nil {
		return nil, errors.NewInvalidPaginationTokenError(token, err)
	}

	return &np, nil
}

const (
	// Default values
	defaultTimeout = 10 * time.Second
	maxRetries     = 3
	baseRetryDelay = 1 * time.Second

	// Rate limiting
	requestsPerSecond = 3 // Etherscan limit is 5
	burstSize         = 2
)

// Response represents the standard Etherscan API response format
type etherscanResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
}

// EtherscanExplorer implements the BlockExplorer interface for Etherscan
type EtherscanExplorer struct {
	apiURL      string
	explorerURL string
	apiKey      string
	chain       *types.Chain
	httpClient  *http.Client
	log         logger.Logger

	// Rate limiting
	limiter *rate.Limiter
}

// NewEtherscanExplorer creates a new instance of EtherscanExplorer
func NewEtherscanExplorer(chain types.Chain, apiURL, explorerURL, apiKey string, log logger.Logger) BlockExplorer {
	e := &EtherscanExplorer{
		apiURL:      apiURL,
		explorerURL: explorerURL,
		apiKey:      apiKey,
		chain:       &chain,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		log:     log,
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), burstSize),
	}

	return e
}

// makeRequest performs an HTTP GET request to the Etherscan API
func (e *EtherscanExplorer) makeRequest(ctx context.Context, params url.Values) ([]byte, error) {
	if params == nil {
		params = url.Values{}
	}

	// Wait for rate limit token
	if err := e.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	return e.doRequest(ctx, params)
}

func (e *EtherscanExplorer) doRequest(ctx context.Context, params url.Values) ([]byte, error) {
	// Add required parameters
	params.Set("apikey", e.apiKey)

	// Construct full URL
	reqURL := fmt.Sprintf("%s?%s", e.apiURL, params.Encode())

	e.log.Debug("Making request to Etherscan API",
		logger.String("url", reqURL),
		logger.String("module", params.Get("module")),
		logger.String("action", params.Get("action")),
		logger.String("chain", string(e.chain.Type)),
	)

	var lastErr error
	for attempt := range maxRetries {
		if attempt > 0 {
			e.log.Debug("Retrying request",
				logger.Int("attempt", attempt+1),
				logger.Int("max_retries", maxRetries),
			)
			// Calculate exponential backoff delay
			backoffDelay := min(baseRetryDelay*time.Duration(1<<uint(attempt)), 10*time.Second)
			time.Sleep(backoffDelay)

			// For retries, ensure we respect rate limits
			if err := e.limiter.Wait(ctx); err != nil {
				return nil, err
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			e.log.Error("Failed to create request",
				logger.Error(err),
				logger.String("module", params.Get("module")),
				logger.String("action", params.Get("action")),
			)
			return nil, errors.NewExplorerRequestFailedError(err)
		}

		resp, err := e.httpClient.Do(req)
		if err != nil {
			lastErr = err
			e.log.Warn("Request failed, will retry with backoff",
				logger.Error(err),
				logger.Int("attempt", attempt+1),
			)
			continue
		}
		defer resp.Body.Close()

		// Check for rate limiting
		if resp.StatusCode == http.StatusTooManyRequests {
			e.log.Warn("Rate limit exceeded, waiting before retry",
				logger.String("module", params.Get("module")),
				logger.String("action", params.Get("action")),
				logger.String("chain", string(e.chain.Type)),
			)
			// Calculate exponential backoff delay
			backoffDelay := min(baseRetryDelay*time.Duration(1<<uint(attempt+2)), 15*time.Second)
			time.Sleep(backoffDelay)
			continue
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			e.log.Warn("Failed to read response body, will retry",
				logger.Error(err),
				logger.Int("attempt", attempt+1),
			)
			continue
		}

		// Parse response
		var response etherscanResponse
		if err := json.Unmarshal(body, &response); err != nil {
			e.log.Error("Failed to parse response",
				logger.Error(err),
				logger.String("body", string(body)),
			)
			return nil, errors.NewInvalidExplorerResponseError(err, string(body))
		}

		// Check for API errors
		if response.Status == "0" && response.Message == "NOTOK" {
			if strings.Contains(response.Message, "Invalid API Key") {
				e.log.Error("Invalid API key",
					logger.String("module", params.Get("module")),
					logger.String("action", params.Get("action")),
					logger.String("chain", string(e.chain.Type)),
					logger.String("result", string(response.Result)),
				)
				return nil, errors.NewInvalidAPIKeyError()
			}
			if strings.Contains(string(response.Result), "Max rate limit reached") {
				e.log.Warn("Rate limit exceeded in API response, waiting before retry",
					logger.String("module", params.Get("module")),
					logger.String("action", params.Get("action")),
					logger.String("result", string(response.Result)),
				)
				// Calculate exponential backoff delay
				backoffDelay := min(baseRetryDelay*time.Duration(1<<uint(attempt+2)), 15*time.Second)
				time.Sleep(backoffDelay)
				continue
			}
			e.log.Error("API error",
				logger.String("message", response.Message),
				logger.String("module", params.Get("module")),
				logger.String("action", params.Get("action")),
				logger.String("result", string(response.Result)),
			)
			return nil, errors.NewExplorerError(fmt.Errorf("API error: %s", response.Message))
		}

		e.log.Debug("Request successful",
			logger.String("module", params.Get("module")),
			logger.String("action", params.Get("action")),
			logger.String("chain", string(e.chain.Type)),
		)

		return response.Result, nil
	}

	e.log.Error("Max retries exceeded",
		logger.Error(lastErr),
		logger.String("module", params.Get("module")),
		logger.String("action", params.Get("action")),
	)
	return nil, errors.NewExplorerRequestFailedError(lastErr)
}

// GetContract implements BlockExplorer.GetContract
func (e *EtherscanExplorer) GetContract(ctx context.Context, address string) (*ContractInfo, error) {
	// Get contract ABI
	params := url.Values{}
	params.Set("module", "contract")
	params.Set("action", "getabi")
	params.Set("address", address)

	abiData, err := e.makeRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	// Get contract source code
	params.Set("action", "getsourcecode")
	sourceData, err := e.makeRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	var sourceInfo []struct {
		SourceCode      string `json:"SourceCode"`
		ContractName    string `json:"ContractName"`
		CompilerVersion string `json:"CompilerVersion"`
	}

	if err := json.Unmarshal(sourceData, &sourceInfo); err != nil {
		return nil, errors.NewInvalidExplorerResponseError(err, string(sourceData))
	}

	if len(sourceInfo) == 0 {
		return nil, errors.NewContractNotFoundError(address, string(e.chain.Type))
	}

	return &ContractInfo{
		ABI:          string(abiData),
		ContractName: sourceInfo[0].ContractName,
		SourceCode:   sourceInfo[0].SourceCode,
		IsVerified:   sourceInfo[0].SourceCode != "",
	}, nil
}

// getNormalTransactionHistory fetches normal transactions for an address
func (e *EtherscanExplorer) getNormalTransactionHistory(ctx context.Context, address string, options TransactionHistoryOptions, page, limit int) ([]*NormalTxHistoryEntry, error) {
	params := url.Values{}
	// Request limit+1 items to determine if there's a next page
	e.setTransactionHistoryParams(params, address, options, "txlist", page, limit+1)

	data, err := e.makeRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	var txs []struct {
		Hash             string `json:"hash"`
		From             string `json:"from"`
		To               string `json:"to"`
		Value            string `json:"value"`
		Gas              string `json:"gas"`
		GasPrice         string `json:"gasPrice"`
		GasUsed          string `json:"gasUsed"`
		Nonce            string `json:"nonce"`
		BlockHash        string `json:"blockHash"`
		BlockNumber      string `json:"blockNumber"`
		TransactionIndex string `json:"transactionIndex"`
		Timestamp        string `json:"timeStamp"`
		IsError          string `json:"isError"`
		ContractAddress  string `json:"contractAddress"`
	}

	if err := json.Unmarshal(data, &txs); err != nil {
		return nil, errors.NewInvalidExplorerResponseError(err, string(data))
	}

	result := make([]*NormalTxHistoryEntry, 0, len(txs))
	for _, tx := range txs {
		blockNumber := new(big.Int)
		blockNumber.SetString(tx.BlockNumber, 10)
		gasLimit, _ := strconv.ParseUint(tx.Gas, 10, 64)
		gasPrice := new(big.Int)
		gasPrice.SetString(tx.GasPrice, 10)
		gasUsed, _ := strconv.ParseUint(tx.GasUsed, 10, 64)
		nonce, _ := strconv.ParseUint(tx.Nonce, 10, 64)
		value := new(big.Int)
		value.SetString(tx.Value, 10)
		timestamp, _ := strconv.ParseInt(tx.Timestamp, 10, 64)

		status := types.TransactionStatusSuccess
		if tx.IsError == "1" {
			status = types.TransactionStatusFailed
		}

		txType := types.TransactionTypeNative
		if (tx.To == "" || tx.To == "0x") && tx.ContractAddress != "" {
			txType = types.TransactionTypeDeploy
		} // Cannot reliably detect ContractCall without input data

		baseTx := types.BaseTransaction{
			ChainType: e.chain.Type,
			Hash:      tx.Hash,
			From:      tx.From,
			To:        tx.To,
			Value:     value,
			Data:      nil, // Not provided
			Nonce:     nonce,
			GasPrice:  gasPrice,
			GasLimit:  gasLimit,
			Type:      txType,
		}

		txEntry := types.Transaction{
			BaseTransaction: baseTx,
			Status:          status,
			Timestamp:       timestamp,
			BlockNumber:     blockNumber,
			GasUsed:         gasUsed,
		}

		normalEntry := &NormalTxHistoryEntry{
			Transaction:     txEntry,
			ContractAddress: tx.ContractAddress, // Specific to normal tx (deploy)
		}
		result = append(result, normalEntry)
	}

	return result, nil
}

// getInternalTransactionHistory fetches internal transactions for an address
func (e *EtherscanExplorer) getInternalTransactionHistory(ctx context.Context, address string, options TransactionHistoryOptions, page, limit int) ([]*InternalTxHistoryEntry, error) {
	params := url.Values{}
	// Request limit+1 items to determine if there's a next page
	e.setTransactionHistoryParams(params, address, options, "txlistinternal", page, limit+1)

	data, err := e.makeRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	var txs []struct {
		Hash            string `json:"hash"`
		From            string `json:"from"`
		To              string `json:"to"`
		Value           string `json:"value"`
		Gas             string `json:"gas"`
		GasUsed         string `json:"gasUsed"`
		BlockNumber     string `json:"blockNumber"`
		Timestamp       string `json:"timeStamp"`
		IsError         string `json:"isError"`
		ContractAddress string `json:"contractAddress"`
	}

	if err := json.Unmarshal(data, &txs); err != nil {
		return nil, errors.NewInvalidExplorerResponseError(err, string(data))
	}

	result := make([]*InternalTxHistoryEntry, 0, len(txs))
	for _, tx := range txs {
		blockNumber := new(big.Int)
		blockNumber.SetString(tx.BlockNumber, 10)
		value := new(big.Int)
		value.SetString(tx.Value, 10)
		gasUsed, _ := strconv.ParseUint(tx.GasUsed, 10, 64)
		timestamp, _ := strconv.ParseInt(tx.Timestamp, 10, 64)

		status := types.TransactionStatusSuccess
		if tx.IsError == "1" {
			status = types.TransactionStatusFailed
		}

		// Internal transactions are native transfers triggered by contracts
		txType := types.TransactionTypeNative

		baseTx := types.BaseTransaction{
			ChainType: e.chain.Type,
			Hash:      tx.Hash, // Parent tx hash
			From:      tx.From,
			To:        tx.To,
			Value:     value,
			Type:      txType,
			// Nonce, GasPrice, GasLimit, Data are zero/nil
		}

		txEntry := types.Transaction{
			BaseTransaction: baseTx,
			Status:          status,
			Timestamp:       timestamp,
			BlockNumber:     blockNumber,
			GasUsed:         gasUsed,
		}

		internalEntry := &InternalTxHistoryEntry{
			Transaction: txEntry,
		}
		result = append(result, internalEntry)
	}

	return result, nil
}

// getERC20TransactionHistory fetches ERC20 token transfers for an address
func (e *EtherscanExplorer) getERC20TransactionHistory(ctx context.Context, address string, options TransactionHistoryOptions, page, limit int) ([]*ERC20TxHistoryEntry, error) {
	params := url.Values{}
	// Request limit+1 items to determine if there's a next page
	e.setTransactionHistoryParams(params, address, options, "tokentx", page, limit+1)

	data, err := e.makeRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	var txs []struct {
		Hash            string `json:"hash"`
		From            string `json:"from"`
		To              string `json:"to"`    // This is the recipient of the transfer
		Value           string `json:"value"` // This is the token amount
		Gas             string `json:"gas"`
		GasPrice        string `json:"gasPrice"`
		Nonce           string `json:"nonce"`
		BlockNumber     string `json:"blockNumber"`
		Timestamp       string `json:"timeStamp"`
		ContractAddress string `json:"contractAddress"` // This is the token address
		TokenName       string `json:"tokenName"`
		TokenSymbol     string `json:"tokenSymbol"`
		TokenDecimal    string `json:"tokenDecimal"`
	}

	if err := json.Unmarshal(data, &txs); err != nil {
		return nil, errors.NewInvalidExplorerResponseError(err, string(data))
	}

	result := make([]*ERC20TxHistoryEntry, 0, len(txs))
	for _, tx := range txs {
		blockNumber := new(big.Int)
		blockNumber.SetString(tx.BlockNumber, 10)
		gasLimit, _ := strconv.ParseUint(tx.Gas, 10, 64)
		gasPrice := new(big.Int)
		gasPrice.SetString(tx.GasPrice, 10)
		nonce, _ := strconv.ParseUint(tx.Nonce, 10, 64)
		tokenAmount := new(big.Int)
		tokenAmount.SetString(tx.Value, 10)
		timestamp, _ := strconv.ParseInt(tx.Timestamp, 10, 64)

		// ERC20 transfers are contract calls
		txType := types.TransactionTypeContractCall

		// Status is assumed success for token transfers listed here
		status := types.TransactionStatusSuccess

		baseTx := types.BaseTransaction{
			ChainType: e.chain.Type,
			Hash:      tx.Hash,
			From:      tx.From,
			To:        tx.ContractAddress, // Tx interacts with token contract
			Value:     big.NewInt(0),      // Native value likely 0
			Data:      nil,                // Not provided
			Nonce:     nonce,
			GasPrice:  gasPrice,
			GasLimit:  gasLimit,
			Type:      txType,
		}

		txEntry := types.Transaction{
			BaseTransaction: baseTx,
			Status:          status,
			Timestamp:       timestamp,
			BlockNumber:     blockNumber,
			// GasUsed not provided by tokentx endpoint
		}

		erc20Entry := &ERC20TxHistoryEntry{
			Transaction:    txEntry,
			TokenAddress:   tx.ContractAddress,
			TokenSymbol:    tx.TokenSymbol,
			TokenRecipient: tx.To, // Actual recipient of tokens
			TokenAmount:    tokenAmount,
		}
		result = append(result, erc20Entry)
	}

	return result, nil
}

// getERC721TransactionHistory fetches ERC721 (NFT) token transfers for an address
func (e *EtherscanExplorer) getERC721TransactionHistory(ctx context.Context, address string, options TransactionHistoryOptions, page, limit int) ([]*ERC721TxHistoryEntry, error) {
	params := url.Values{}
	// Request limit+1 items to determine if there's a next page
	// Use "tokennfttx" action for ERC721
	e.setTransactionHistoryParams(params, address, options, "tokennfttx", page, limit+1)

	data, err := e.makeRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	var txs []struct {
		Hash            string `json:"hash"`
		From            string `json:"from"`
		To              string `json:"to"` // Recipient of NFT
		Gas             string `json:"gas"`
		GasPrice        string `json:"gasPrice"`
		Nonce           string `json:"nonce"`
		BlockNumber     string `json:"blockNumber"`
		Timestamp       string `json:"timeStamp"`
		ContractAddress string `json:"contractAddress"` // NFT contract address
		TokenName       string `json:"tokenName"`
		TokenSymbol     string `json:"tokenSymbol"`
		TokenID         string `json:"tokenID"`
		// tokenDecimal is usually 0 for NFTs, not needed here
	}

	if err := json.Unmarshal(data, &txs); err != nil {
		return nil, errors.NewInvalidExplorerResponseError(err, string(data))
	}

	result := make([]*ERC721TxHistoryEntry, 0, len(txs))
	for _, tx := range txs {
		blockNumber := new(big.Int)
		blockNumber.SetString(tx.BlockNumber, 10)
		gasLimit, _ := strconv.ParseUint(tx.Gas, 10, 64)
		gasPrice := new(big.Int)
		gasPrice.SetString(tx.GasPrice, 10)
		nonce, _ := strconv.ParseUint(tx.Nonce, 10, 64)
		tokenID := new(big.Int)
		tokenID.SetString(tx.TokenID, 10)
		timestamp, _ := strconv.ParseInt(tx.Timestamp, 10, 64)

		// ERC721 transfers are contract calls
		txType := types.TransactionTypeContractCall

		// Status is assumed success for token transfers listed here
		status := types.TransactionStatusSuccess

		baseTx := types.BaseTransaction{
			ChainType: e.chain.Type,
			Hash:      tx.Hash,
			From:      tx.From,
			To:        tx.ContractAddress, // Tx interacts with NFT contract
			Value:     big.NewInt(0),      // Native value likely 0
			Data:      nil,                // Not provided
			Nonce:     nonce,
			GasPrice:  gasPrice,
			GasLimit:  gasLimit,
			Type:      txType,
		}

		txEntry := types.Transaction{
			BaseTransaction: baseTx,
			Status:          status,
			Timestamp:       timestamp,
			BlockNumber:     blockNumber,
			// GasUsed not provided by tokenfttx endpoint
		}

		erc721Entry := &ERC721TxHistoryEntry{
			Transaction:  txEntry,
			TokenAddress: tx.ContractAddress,
			TokenSymbol:  tx.TokenSymbol,
			TokenName:    tx.TokenName,
			TokenID:      tokenID,
		}
		result = append(result, erc721Entry)
	}

	return result, nil
}

// GetTransactionHistory retrieves transaction history for an address with pagination
func (e *EtherscanExplorer) GetTransactionHistory(ctx context.Context, address string, options TransactionHistoryOptions, nextToken string) (*types.Page[any], error) {
	if !e.chain.IsValidAddress(address) {
		return nil, errors.NewInvalidAddressError(address)
	}

	// Default limit if not specified
	limit := 10
	if options.Limit > 0 {
		limit = options.Limit
	}

	// Decode next page token or start at page 1
	nextPage, err := DecodeNextPage(nextToken)
	if err != nil {
		return nil, err
	}
	currentPage := nextPage.Page

	// If no transaction type is specified, default to normal transactions
	txType := options.TransactionType
	if txType == "" {
		txType = TxTypeNormal
	}

	// Fetch transactions based on the specified type
	var fetchedItems []any // Use []any to hold results from different helpers
	var itemLength int
	var fetchErr error

	switch txType {
	case TxTypeNormal:
		txs, err := e.getNormalTransactionHistory(ctx, address, options, currentPage, limit)
		if err == nil {
			fetchedItems = make([]any, len(txs))
			for i, tx := range txs {
				fetchedItems[i] = tx
			}
			itemLength = len(txs)
		} else {
			fetchErr = err
		}
	case TxTypeInternal:
		txs, err := e.getInternalTransactionHistory(ctx, address, options, currentPage, limit)
		if err == nil {
			fetchedItems = make([]any, len(txs))
			for i, tx := range txs {
				fetchedItems[i] = tx
			}
			itemLength = len(txs)
		} else {
			fetchErr = err
		}
	case TxTypeERC20:
		erc20Txs, err := e.getERC20TransactionHistory(ctx, address, options, currentPage, limit)
		if err == nil {
			fetchedItems = make([]any, len(erc20Txs))
			for i, tx := range erc20Txs {
				fetchedItems[i] = tx
			}
			itemLength = len(erc20Txs)
		} else {
			fetchErr = err
		}
	case TxTypeERC721:
		erc721Txs, err := e.getERC721TransactionHistory(ctx, address, options, currentPage, limit)
		if err == nil {
			fetchedItems = make([]any, len(erc721Txs))
			for i, tx := range erc721Txs {
				fetchedItems[i] = tx
			}
			itemLength = len(erc721Txs)
		} else {
			fetchErr = err
		}
	default:
		return nil, errors.NewExplorerError(fmt.Errorf("unsupported transaction type: %s", txType))
	}

	if fetchErr != nil {
		return nil, fetchErr
	}

	// Check for next page (remember we fetched limit + 1)
	var nextPageToken string
	hasMore := itemLength > limit
	if hasMore {
		nextPage := &NextPage{Page: currentPage + 1}
		nextPageToken = nextPage.Encode()
		// Trim to limit for the response
		fetchedItems = fetchedItems[:limit]
	}

	// Return page of type any
	return &types.Page[any]{
		Items:     fetchedItems,
		NextToken: nextPageToken,
		Limit:     limit,
	}, nil
}

// setTransactionHistoryParams sets common parameters for transaction history queries
func (e *EtherscanExplorer) setTransactionHistoryParams(params url.Values, address string, options TransactionHistoryOptions, action string, page, limit int) {
	params.Set("module", "account")
	params.Set("action", action)
	params.Set("address", address)
	params.Set("startblock", strconv.FormatInt(options.StartBlock, 10))
	if options.EndBlock != 0 {
		params.Set("endblock", strconv.FormatInt(options.EndBlock, 10))
	}
	params.Set("page", strconv.Itoa(page))
	params.Set("offset", strconv.Itoa(limit))
	params.Set("sort", map[bool]string{true: "asc", false: "desc"}[options.SortAscending])
}

// GetTransactionReceiptByHash implements BlockExplorer.GetTransactionReceiptByHash
func (e *EtherscanExplorer) GetTransactionReceiptByHash(ctx context.Context, hash string) (*types.TransactionReceipt, error) {
	params := url.Values{}
	params.Set("module", "proxy")
	params.Set("action", "eth_getTransactionReceipt")
	params.Set("txhash", hash)

	data, err := e.makeRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	// If the result is empty or null, the transaction receipt was not found
	if len(data) == 0 || string(data) == "null" {
		return nil, errors.NewTransactionNotFoundError(hash)
	}

	// Parse the receipt data
	var rawReceipt struct {
		TransactionHash   string `json:"transactionHash"`
		BlockNumber       string `json:"blockNumber"`
		Status            string `json:"status"`
		GasUsed           string `json:"gasUsed"`
		CumulativeGasUsed string `json:"cumulativeGasUsed"`
		LogsBloom         string `json:"logsBloom"`
		Logs              []struct {
			Address  string   `json:"address"`
			Topics   []string `json:"topics"`
			Data     string   `json:"data"`
			LogIndex string   `json:"logIndex"`
		} `json:"logs"`
	}

	if err := json.Unmarshal(data, &rawReceipt); err != nil {
		return nil, errors.NewInvalidExplorerResponseError(err, string(data))
	}

	// Convert hex values to decimal
	blockNumber := new(big.Int)
	blockNumber.SetString(strings.TrimPrefix(rawReceipt.BlockNumber, "0x"), 16)

	status, _ := strconv.ParseUint(strings.TrimPrefix(rawReceipt.Status, "0x"), 16, 64)
	gasUsed, _ := strconv.ParseUint(strings.TrimPrefix(rawReceipt.GasUsed, "0x"), 16, 64)
	cumulativeGasUsed, _ := strconv.ParseUint(strings.TrimPrefix(rawReceipt.CumulativeGasUsed, "0x"), 16, 64)

	// Convert logs bloom from hex to bytes
	logsBloom, _ := hex.DecodeString(strings.TrimPrefix(rawReceipt.LogsBloom, "0x"))

	// Convert logs
	logs := make([]types.Log, len(rawReceipt.Logs))
	for i, log := range rawReceipt.Logs {
		// Convert log data from hex to bytes
		logData, _ := hex.DecodeString(strings.TrimPrefix(log.Data, "0x"))
		logIndex, _ := strconv.ParseUint(strings.TrimPrefix(log.LogIndex, "0x"), 16, 32)

		logs[i] = types.Log{
			Address:         log.Address,
			Topics:          log.Topics,
			Data:            logData,
			BlockNumber:     blockNumber,
			TransactionHash: rawReceipt.TransactionHash,
			LogIndex:        uint(logIndex),
		}
	}

	e.log.Debug("Retrieved transaction receipt",
		logger.String("hash", hash),
		logger.String("status", strconv.FormatUint(status, 10)),
		logger.String("block_number", blockNumber.String()),
		logger.Int("log_count", len(logs)))

	return &types.TransactionReceipt{
		Hash:              rawReceipt.TransactionHash,
		BlockNumber:       blockNumber,
		Status:            status,
		GasUsed:           gasUsed,
		CumulativeGasUsed: cumulativeGasUsed,
		LogsBloom:         logsBloom,
		Logs:              logs,
	}, nil
}

// GetTransactionByHash implements BlockExplorer.GetTransactionByHash
func (e *EtherscanExplorer) GetTransactionByHash(ctx context.Context, hash string) (*types.Transaction, error) {
	params := url.Values{}
	params.Set("module", "proxy")
	params.Set("action", "eth_getTransactionByHash")
	params.Set("txhash", hash)

	data, err := e.makeRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	// If the result is empty or null, the transaction was not found
	if len(data) == 0 || string(data) == "null" {
		return nil, errors.NewTransactionNotFoundError(hash)
	}

	var tx struct {
		Hash        string `json:"hash"`
		From        string `json:"from"`
		To          string `json:"to"`
		Value       string `json:"value"`
		Gas         string `json:"gas"`
		GasPrice    string `json:"gasPrice"`
		Nonce       string `json:"nonce"`
		BlockHash   string `json:"blockHash"`
		BlockNumber string `json:"blockNumber"`
	}

	if err := json.Unmarshal(data, &tx); err != nil {
		return nil, errors.NewInvalidExplorerResponseError(err, string(data))
	}

	// Convert hex values to decimal
	blockNumber := new(big.Int)
	if tx.BlockNumber != "" && tx.BlockNumber != "0x" {
		blockNumber.SetString(strings.TrimPrefix(tx.BlockNumber, "0x"), 16)
	}

	gasLimit, _ := strconv.ParseUint(strings.TrimPrefix(tx.Gas, "0x"), 16, 64)
	gasPrice := new(big.Int)
	gasPrice.SetString(strings.TrimPrefix(tx.GasPrice, "0x"), 16)
	nonce, _ := strconv.ParseUint(strings.TrimPrefix(tx.Nonce, "0x"), 16, 64)
	value := new(big.Int)
	value.SetString(strings.TrimPrefix(tx.Value, "0x"), 16)

	// Default status to pending
	status := types.TransactionStatusPending

	// If transaction is in a block, check the receipt status
	var gasUsed uint64
	var receiptStatus uint64
	if blockNumber.Int64() > 0 && tx.BlockHash != "" {
		status = types.TransactionStatusMined

		// Try to get the receipt for detailed status and gas used
		receipt, err := e.GetTransactionReceiptByHash(ctx, hash)
		if err == nil {
			receiptStatus = receipt.Status
			gasUsed = receipt.GasUsed
			// Status 1 means success, 0 means failure
			if receiptStatus == 1 {
				status = types.TransactionStatusSuccess
			} else if receiptStatus == 0 {
				status = types.TransactionStatusFailed
			}
		}
	}

	// Fetch input data separately if needed (not returned by eth_getTransactionByHash)
	// This requires another API call which we avoid here for performance.
	// If input data is crucial, consider using a node provider directly.
	var inputData []byte // Placeholder

	// Determine transaction type based on presence of input data and 'to' address
	txType := types.TransactionTypeNative
	if tx.To == "" || tx.To == "0x" { // Assuming input data would be present for deployment
		txType = types.TransactionTypeDeploy // Needs input data to confirm
	} else if len(inputData) > 0 { // Placeholder check
		txType = types.TransactionTypeContractCall
	}

	e.log.Debug("Retrieved transaction",
		logger.String("hash", hash),
		logger.String("status", string(status)),
		logger.String("block_number", blockNumber.String()))

	baseTx := types.BaseTransaction{
		ChainType: e.chain.Type,
		Hash:      tx.Hash,
		From:      tx.From,
		To:        tx.To,
		Value:     value,
		Data:      inputData, // Would be nil here
		Nonce:     nonce,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Type:      txType,
	}

	return &types.Transaction{
		BaseTransaction: baseTx,
		Status:          status,
		BlockNumber:     blockNumber,
		GasUsed:         gasUsed, // Populated if receipt was fetched
		// Timestamp is not directly available from eth_getTransactionByHash or its receipt
		// Needs a separate eth_getBlockByNumber call if required.
	}, nil
}

// GetTokenURL implements BlockExplorer.GetTokenURL
func (e *EtherscanExplorer) GetTokenURL(address string) string {
	return fmt.Sprintf("%s/token/%s", e.explorerURL, address)
}

// Chain implements BlockExplorer.Chain
func (e *EtherscanExplorer) Chain() types.Chain {
	return *e.chain
}
