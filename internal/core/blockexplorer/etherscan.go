package blockexplorer

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"

	"golang.org/x/time/rate"
)

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
			backoffDelay := baseRetryDelay * time.Duration(1<<uint(attempt+2)) // More aggressive backoff for rate limits
			if backoffDelay > 15*time.Second {
				backoffDelay = 15 * time.Second // Cap at 15 seconds for rate limits
			}
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
				backoffDelay := baseRetryDelay * time.Duration(1<<uint(attempt+2)) // More aggressive backoff for rate limits
				if backoffDelay > 15*time.Second {
					backoffDelay = 15 * time.Second // Cap at 15 seconds for rate limits
				}
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
func (e *EtherscanExplorer) getNormalTransactionHistory(ctx context.Context, address string, options TransactionHistoryOptions) ([]*types.Transaction, error) {
	params := url.Values{}
	e.setTransactionHistoryParams(params, address, options, "txlist")

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

	result := make([]*types.Transaction, len(txs))
	for i, tx := range txs {
		blockNumber := new(big.Int)
		blockNumber.SetString(tx.BlockNumber, 10)
		gasLimit, _ := strconv.ParseUint(tx.Gas, 10, 64)
		gasPrice := new(big.Int)
		gasPrice.SetString(tx.GasPrice, 10)
		nonce, _ := strconv.ParseUint(tx.Nonce, 10, 64)
		value := new(big.Int)
		value.SetString(tx.Value, 10)

		// Parse timestamp with error checking
		timestamp := time.Now().Unix()
		if tx.Timestamp != "" {
			if timestampInt, err := strconv.ParseInt(tx.Timestamp, 10, 64); err == nil {
				timestamp = timestampInt
			} else {
				e.log.Debug("Failed to parse transaction timestamp",
					logger.String("hash", tx.Hash),
					logger.String("timestamp_str", tx.Timestamp),
					logger.Error(err))
			}
		}

		result[i] = &types.Transaction{
			Chain:        e.chain.Type,
			Hash:         tx.Hash,
			From:         tx.From,
			To:           tx.To,
			Value:        value,
			Data:         nil, // Etherscan doesn't return transaction data
			Nonce:        nonce,
			GasPrice:     gasPrice,
			GasLimit:     gasLimit,
			Type:         types.TransactionTypeNative,
			TokenAddress: tx.ContractAddress,
			Status:       types.TransactionStatus(map[string]string{"0": "success", "1": "failed"}[tx.IsError]),
			Timestamp:    timestamp,
			BlockNumber:  blockNumber,
		}
	}

	return result, nil
}

// getInternalTransactionHistory fetches internal transactions for an address
func (e *EtherscanExplorer) getInternalTransactionHistory(ctx context.Context, address string, options TransactionHistoryOptions) ([]*types.Transaction, error) {
	params := url.Values{}
	e.setTransactionHistoryParams(params, address, options, "txlistinternal")

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

	result := make([]*types.Transaction, len(txs))
	for i, tx := range txs {
		blockNumber := new(big.Int)
		blockNumber.SetString(tx.BlockNumber, 10)
		value := new(big.Int)
		value.SetString(tx.Value, 10)

		// Parse timestamp with error checking
		timestamp := time.Now().Unix()
		if tx.Timestamp != "" {
			if timestampInt, err := strconv.ParseInt(tx.Timestamp, 10, 64); err == nil {
				timestamp = timestampInt
			} else {
				e.log.Debug("Failed to parse transaction timestamp",
					logger.String("hash", tx.Hash),
					logger.String("timestamp_str", tx.Timestamp),
					logger.Error(err))
			}
		}

		result[i] = &types.Transaction{
			Chain:        e.chain.Type,
			Hash:         tx.Hash,
			From:         tx.From,
			To:           tx.To,
			Value:        value,
			Data:         nil,
			Type:         types.TransactionTypeNative,
			TokenAddress: tx.ContractAddress,
			Status:       types.TransactionStatus(map[string]string{"0": "success", "1": "failed"}[tx.IsError]),
			Timestamp:    timestamp,
			BlockNumber:  blockNumber,
		}
	}

	return result, nil
}

// getERC20TransactionHistory fetches ERC20 token transfers for an address
func (e *EtherscanExplorer) getERC20TransactionHistory(ctx context.Context, address string, options TransactionHistoryOptions) ([]*types.Transaction, error) {
	params := url.Values{}
	e.setTransactionHistoryParams(params, address, options, "tokentx")

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
		GasPrice        string `json:"gasPrice"`
		Nonce           string `json:"nonce"`
		BlockNumber     string `json:"blockNumber"`
		Timestamp       string `json:"timeStamp"`
		ContractAddress string `json:"contractAddress"`
		TokenName       string `json:"tokenName"`
		TokenSymbol     string `json:"tokenSymbol"`
		TokenDecimal    string `json:"tokenDecimal"`
	}

	if err := json.Unmarshal(data, &txs); err != nil {
		return nil, errors.NewInvalidExplorerResponseError(err, string(data))
	}

	result := make([]*types.Transaction, len(txs))
	for i, tx := range txs {
		blockNumber := new(big.Int)
		blockNumber.SetString(tx.BlockNumber, 10)
		gasLimit, _ := strconv.ParseUint(tx.Gas, 10, 64)
		gasPrice := new(big.Int)
		gasPrice.SetString(tx.GasPrice, 10)
		nonce, _ := strconv.ParseUint(tx.Nonce, 10, 64)
		value := new(big.Int)
		value.SetString(tx.Value, 10)

		// Parse timestamp with error checking
		timestamp := time.Now().Unix()
		if tx.Timestamp != "" {
			if timestampInt, err := strconv.ParseInt(tx.Timestamp, 10, 64); err == nil {
				timestamp = timestampInt
			} else {
				e.log.Debug("Failed to parse transaction timestamp",
					logger.String("hash", tx.Hash),
					logger.String("timestamp_str", tx.Timestamp),
					logger.Error(err))
			}
		}

		result[i] = &types.Transaction{
			Chain:        e.chain.Type,
			Hash:         tx.Hash,
			From:         tx.From,
			To:           tx.To,
			Value:        value,
			Data:         nil,
			Nonce:        nonce,
			GasPrice:     gasPrice,
			GasLimit:     gasLimit,
			Type:         types.TransactionTypeERC20,
			TokenAddress: tx.ContractAddress,
			Status:       types.TransactionStatusSuccess,
			Timestamp:    timestamp,
			BlockNumber:  blockNumber,
		}
	}

	return result, nil
}

// GetTransactionHistory retrieves transaction history for an address with pagination
func (e *EtherscanExplorer) GetTransactionHistory(ctx context.Context, address string, options TransactionHistoryOptions) (*types.Page[*types.Transaction], error) {
	if !e.chain.IsValidAddress(address) {
		return nil, errors.NewInvalidAddressError(address)
	}

	// Default page size if not specified
	if options.PageSize == 0 {
		options.PageSize = 10
	}

	// Default page number if not specified
	if options.Page == 0 {
		options.Page = 1
	}

	// Initialize variables to store transactions and total count
	var allTransactions []*types.Transaction
	var total int64

	// If no specific types are requested, fetch all types
	if len(options.TransactionTypes) == 0 {
		options.TransactionTypes = []TransactionType{TxTypeNormal, TxTypeInternal, TxTypeERC20}
	}

	// Fetch transactions based on requested types
	for _, txType := range options.TransactionTypes {
		var txs []*types.Transaction
		var err error

		switch txType {
		case TxTypeNormal:
			txs, err = e.getNormalTransactionHistory(ctx, address, options)
		case TxTypeInternal:
			txs, err = e.getInternalTransactionHistory(ctx, address, options)
		case TxTypeERC20:
			txs, err = e.getERC20TransactionHistory(ctx, address, options)
		default:
			continue
		}

		if err != nil {
			return nil, err
		}

		allTransactions = append(allTransactions, txs...)
		total += int64(len(txs))
	}

	// Sort transactions by timestamp
	sort.Slice(allTransactions, func(i, j int) bool {
		if options.SortAscending {
			return allTransactions[i].Timestamp < allTransactions[j].Timestamp
		}
		return allTransactions[i].Timestamp > allTransactions[j].Timestamp
	})

	// Calculate pagination
	start := (options.Page - 1) * options.PageSize
	end := min(start+options.PageSize, len(allTransactions))

	// Check if there are more pages
	hasMore := end < len(allTransactions)

	// Return paginated results
	return &types.Page[*types.Transaction]{
		Items:   allTransactions[start:end],
		Offset:  (options.Page - 1) * options.PageSize,
		Limit:   options.PageSize,
		HasMore: hasMore,
	}, nil
}

// setTransactionHistoryParams sets common parameters for transaction history queries
func (e *EtherscanExplorer) setTransactionHistoryParams(params url.Values, address string, options TransactionHistoryOptions, action string) {
	params.Set("module", "account")
	params.Set("action", action)
	params.Set("address", address)
	params.Set("startblock", strconv.FormatInt(options.StartBlock, 10))
	if options.EndBlock != 0 {
		params.Set("endblock", strconv.FormatInt(options.EndBlock, 10))
	}
	params.Set("page", strconv.Itoa(options.Page))
	params.Set("offset", strconv.Itoa(options.PageSize))
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
	if blockNumber.Int64() > 0 && tx.BlockHash != "" {
		status = types.TransactionStatusMined

		// Try to get the receipt for detailed status
		receipt, err := e.GetTransactionReceiptByHash(ctx, hash)
		if err == nil {
			// Status 1 means success, 0 means failure
			if receipt.Status == 1 {
				status = types.TransactionStatusSuccess
			} else if receipt.Status == 0 {
				status = types.TransactionStatusFailed
			}
		}
	}

	e.log.Debug("Retrieved transaction",
		logger.String("hash", hash),
		logger.String("status", string(status)),
		logger.String("block_number", blockNumber.String()))

	return &types.Transaction{
		Chain:       e.chain.Type,
		Hash:        tx.Hash,
		From:        tx.From,
		To:          tx.To,
		Value:       value,
		Nonce:       nonce,
		GasPrice:    gasPrice,
		GasLimit:    gasLimit,
		Type:        types.TransactionTypeNative,
		Status:      status,
		BlockNumber: blockNumber,
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
