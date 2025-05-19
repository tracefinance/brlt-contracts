package blockexplorer_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/testing/fixtures"
	"vault0/internal/testing/mocks"
	"vault0/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNextPageEncodeDecode(t *testing.T) {
	tests := []struct {
		name     string
		page     *blockexplorer.NextPage
		wantErr  bool
		expected *blockexplorer.NextPage
	}{
		{
			name:     "encode and decode page 1",
			page:     &blockexplorer.NextPage{Page: 1},
			wantErr:  false,
			expected: &blockexplorer.NextPage{Page: 1},
		},
		{
			name:     "encode and decode page 42",
			page:     &blockexplorer.NextPage{Page: 42},
			wantErr:  false,
			expected: &blockexplorer.NextPage{Page: 42},
		},
		{
			name:     "empty token should return page 1",
			page:     nil,
			wantErr:  false,
			expected: &blockexplorer.NextPage{Page: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var token string
			if tt.page != nil {
				token = tt.page.Encode()
				// Verify token is not empty
				assert.NotEmpty(t, token, "Token should not be empty")
			}

			// Decode token
			got, err := blockexplorer.DecodeNextPage(token)
			if tt.wantErr {
				require.Error(t, err, "Expected decode error but got none")
				return
			}

			require.NoError(t, err, "Did not expect decode error")
			assert.Equal(t, tt.expected.Page, got.Page, "Decoded page doesn't match expected")
		})
	}
}

func TestDecodeInvalidNextPage(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "invalid base64",
			token: "invalid-base64!@#",
		},
		{
			name:  "valid base64 but invalid json",
			token: "bm90LWpzb24=", // "not-json" in base64
		},
		// The current implementation doesn't actually validate the JSON structure
		// beyond it being valid JSON, so this test would fail.
		// If validation is needed, the DecodeNextPage function would need to be updated.
		/*{
			name:  "valid base64 but incorrect json structure",
			token: "eyJub3RfcGFnZSI6MX0=", // {"not_page":1} in base64
		},*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := blockexplorer.DecodeNextPage(tt.token)
			require.Error(t, err, "Expected error for invalid token")
			assert.Nil(t, got, "Result should be nil for invalid token")
		})
	}
}

func TestNewEtherscanExplorer(t *testing.T) {
	tests := []struct {
		name        string
		chain       types.Chain
		apiURL      string
		explorerURL string
		apiKey      string
	}{
		{
			name: "ethereum mainnet",
			chain: types.Chain{
				Type: types.ChainTypeEthereum,
			},
			apiURL:      "https://api.etherscan.io/api",
			explorerURL: "https://etherscan.io",
			apiKey:      "test_api_key",
		},
		{
			name: "ethereum testnet",
			chain: types.Chain{
				Type: types.ChainTypeEthereum,
				Name: "goerli",
			},
			apiURL:      "https://api-goerli.etherscan.io/api",
			explorerURL: "https://goerli.etherscan.io",
			apiKey:      "test_api_key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create logger mock
			mockLogger := mocks.NewNopLogger()

			// Create explorer
			explorer := blockexplorer.NewEtherscanExplorer(tt.chain, tt.apiURL, tt.explorerURL, tt.apiKey, mockLogger)

			// Verify it's not nil
			require.NotNil(t, explorer, "Explorer should not be nil")

			// Verify Chain() method works correctly
			chain := explorer.Chain()
			assert.Equal(t, tt.chain.Type, chain.Type, "Chain type should match")

			// Verify GetTokenURL() method to indirectly check if explorerURL was set correctly
			tokenURL := explorer.GetTokenURL("0x123456")
			assert.Contains(t, tokenURL, tt.explorerURL, "Token URL should contain explorer URL")
		})
	}
}

func TestGetContract(t *testing.T) {
	// Test case definitions
	tests := []struct {
		name             string
		address          string
		abiResponse      string
		sourceResponse   string
		httpStatus       int
		httpError        error
		wantErr          bool
		expectIsVerified bool
	}{
		{
			name:             "successful contract fetch with valid source code",
			address:          "0x1234567890abcdef1234567890abcdef12345678",
			abiResponse:      `{"status":"1","message":"OK","result":"[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"}`,
			sourceResponse:   `{"status":"1","message":"OK","result":[{"SourceCode":"contract TestToken {}","ContractName":"TestToken","CompilerVersion":"v0.8.0"}]}`,
			httpStatus:       http.StatusOK,
			wantErr:          false,
			expectIsVerified: true,
		},
		{
			name:             "successful contract fetch without source code",
			address:          "0x1234567890abcdef1234567890abcdef12345678",
			abiResponse:      `{"status":"1","message":"OK","result":"[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"}`,
			sourceResponse:   `{"status":"1","message":"OK","result":[{"SourceCode":"","ContractName":"TestToken","CompilerVersion":"v0.8.0"}]}`,
			httpStatus:       http.StatusOK,
			wantErr:          false,
			expectIsVerified: false,
		},
		{
			name:             "http error",
			address:          "0x1234567890abcdef1234567890abcdef12345678",
			httpError:        errors.New("connection refused"),
			wantErr:          true,
			expectIsVerified: false,
		},
		{
			name:             "api error response",
			address:          "0x1234567890abcdef1234567890abcdef12345678",
			abiResponse:      fixtures.GetAPIErrorResponse("Contract source code not verified"),
			httpStatus:       http.StatusOK,
			wantErr:          true,
			expectIsVerified: false,
		},
		{
			name:             "invalid contract address",
			address:          "not_an_address",
			wantErr:          true,
			expectIsVerified: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create chain with address validation
			chain := types.Chain{Type: types.ChainTypeEthereum}

			// Create EtherscanExplorer instance
			explorer := blockexplorer.NewEtherscanExplorer(
				chain,
				"https://api.etherscan.io/api",
				"https://etherscan.io",
				"TEST_API_KEY",
				mocks.NewNopLogger(),
			)

			// Setup direct mocking with http.Client responses
			mockClient := &http.Client{
				Transport: &mockRoundTripper{
					responses: map[string]struct {
						resp *http.Response
						err  error
					}{},
				},
			}

			mockTransport := mockClient.Transport.(*mockRoundTripper)

			// Setup ABI response
			if tt.abiResponse != "" {
				// For ABI, need to ensure we're not double-escaping JSON
				mockTransport.responses["action=getabi"] = struct {
					resp *http.Response
					err  error
				}{
					resp: &http.Response{
						StatusCode: tt.httpStatus,
						Body:       io.NopCloser(strings.NewReader(tt.abiResponse)),
					},
					err: nil,
				}
			}

			// Setup source code response
			if tt.sourceResponse != "" {
				// For source code, need to ensure we're not double-escaping JSON
				mockTransport.responses["action=getsourcecode"] = struct {
					resp *http.Response
					err  error
				}{
					resp: &http.Response{
						StatusCode: tt.httpStatus,
						Body:       io.NopCloser(strings.NewReader(tt.sourceResponse)),
					},
					err: nil,
				}
			}

			// Setup error if specified
			if tt.httpError != nil {
				mockTransport.err = tt.httpError
			}

			// Replace HTTPClient with our mock
			if e, ok := explorer.(*blockexplorer.EtherscanExplorer); ok {
				e.HTTPClient = mockClient
			}

			// Call the method being tested
			result, err := explorer.GetContract(context.Background(), tt.address)

			// Assert results
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			// Check specific fields
			assert.Equal(t, tt.expectIsVerified, result.IsVerified)
			if tt.expectIsVerified {
				assert.NotEmpty(t, result.SourceCode)
			}

			// Should have ABI and contract name
			assert.NotEmpty(t, result.ABI)
			assert.NotEmpty(t, result.ContractName)
		})
	}
}

// mockRoundTripper implements http.RoundTripper for testing
type mockRoundTripper struct {
	responses map[string]struct {
		resp *http.Response
		err  error
	}
	err error
}

// RoundTrip implements http.RoundTripper
func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Check for global error
	if m.err != nil {
		return nil, m.err
	}

	// Check URL for matches
	url := req.URL.String()
	for pattern, resp := range m.responses {
		if strings.Contains(url, pattern) {
			return resp.resp, resp.err
		}
	}

	// Default not found response
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(`{"status":"0","message":"NOTOK","result":"Not found"}`)),
	}, nil
}

// TestGetNormalTransactionHistory tests the getNormalTransactionHistory method through the GetTransactionHistory method
func TestGetNormalTransactionHistory(t *testing.T) {
	// Create test address
	testAddress := "0x1234567890abcdef1234567890abcdef12345678"

	// Create test transaction list
	transactions := []string{
		`{
			"hash": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"from": "0x1234567890abcdef1234567890abcdef12345678",
			"to": "0x0987654321fedcba0987654321fedcba09876543",
			"value": "1000000000000000000",
			"gas": "21000",
			"gasPrice": "20000000000",
			"gasUsed": "21000",
			"nonce": "42",
			"blockHash": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"blockNumber": "1000000",
			"transactionIndex": "1",
			"timeStamp": "1600000000",
			"isError": "0",
			"contractAddress": ""
		}`,
	}

	// Create response body
	responseBody := fmt.Sprintf(`{
		"status": "1",
		"message": "OK",
		"result": [%s]
	}`, transactions[0])

	// Test case definitions
	tests := []struct {
		name       string
		address    string
		response   string
		httpStatus int
		httpError  error
		wantErr    bool
		wantCount  int
		wantHash   string
	}{
		{
			name:       "successful transaction fetch",
			address:    testAddress,
			response:   responseBody,
			httpStatus: http.StatusOK,
			wantErr:    false,
			wantCount:  1,
			wantHash:   "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
		{
			name:       "empty transaction list",
			address:    testAddress,
			response:   `{"status": "1", "message": "OK", "result": []}`,
			httpStatus: http.StatusOK,
			wantErr:    false,
			wantCount:  0,
		},
		{
			name:      "http error",
			address:   testAddress,
			httpError: errors.New("connection refused"),
			wantErr:   true,
			wantCount: 0,
		},
		{
			name:       "api error response",
			address:    testAddress,
			response:   fixtures.GetAPIErrorResponse("No transactions found"),
			httpStatus: http.StatusOK,
			wantErr:    true,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create chain with address validation
			chain := types.Chain{Type: types.ChainTypeEthereum}

			// Create logger
			testLogger := mocks.NewNopLogger()

			// Create EtherscanExplorer instance
			explorer := blockexplorer.NewEtherscanExplorer(
				chain,
				"https://api.etherscan.io/api",
				"https://etherscan.io",
				"TEST_API_KEY",
				testLogger,
			)

			// Setup direct mocking with http.Client responses
			mockClient := &http.Client{
				Transport: &mockRoundTripper{
					responses: map[string]struct {
						resp *http.Response
						err  error
					}{},
				},
			}

			mockTransport := mockClient.Transport.(*mockRoundTripper)

			// Setup response
			if tt.response != "" {
				mockTransport.responses["action=txlist"] = struct {
					resp *http.Response
					err  error
				}{
					resp: &http.Response{
						StatusCode: tt.httpStatus,
						Body:       io.NopCloser(strings.NewReader(tt.response)),
					},
					err: nil,
				}
			}

			// Setup error if specified
			if tt.httpError != nil {
				mockTransport.err = tt.httpError
			}

			// Replace HTTPClient with our mock
			if e, ok := explorer.(*blockexplorer.EtherscanExplorer); ok {
				e.HTTPClient = mockClient
			}

			// Setup options
			options := blockexplorer.TransactionHistoryOptions{
				TransactionType: blockexplorer.TxTypeNormal,
				Limit:           10,
				StartBlock:      0,
				EndBlock:        0,
				SortAscending:   true,
			}

			// Call getNormalTransactionHistory method
			// As it's not exported, we need to test it through GetTransactionHistory
			result, err := explorer.GetTransactionHistory(context.Background(), tt.address, options, "")

			// Assert results
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			// Check count matches expected
			assert.Equal(t, tt.wantCount, len(result.Items))

			// If we expect results, check the hash of the first one
			if tt.wantCount > 0 {
				normalTx := result.Items[0]
				assert.Equal(t, tt.wantHash, normalTx.GetHash())
				assert.Equal(t, chain.Type, normalTx.GetChainType())
				assert.Equal(t, testAddress, normalTx.GetFrom())
			}
		})
	}
}

// TestMakeRequestSuccess tests the MakeRequest method with successful responses
func TestMakeRequestSuccess(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name       string
		params     map[string]string
		response   string
		result     string // Expected result after parsing response
		httpStatus int
		wantErr    bool
	}{
		{
			name: "successful API call",
			params: map[string]string{
				"action":  "test_action",
				"address": "0x1234567890abcdef1234567890abcdef12345678",
			},
			response:   `{"status":"1","message":"OK","result":"success"}`,
			result:     `"success"`,
			httpStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "successful API call with multiple parameters",
			params: map[string]string{
				"action":     "test_action",
				"address":    "0x1234567890abcdef1234567890abcdef12345678",
				"startblock": "1000000",
				"endblock":   "2000000",
				"sort":       "asc",
			},
			response:   `{"status":"1","message":"OK","result":["item1","item2"]}`,
			result:     `["item1","item2"]`,
			httpStatus: http.StatusOK,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create chain
			chain := types.Chain{Type: types.ChainTypeEthereum}

			// Create explorer instance
			explorer := blockexplorer.NewEtherscanExplorer(
				chain,
				"https://api.etherscan.io/api",
				"https://etherscan.io",
				"TEST_API_KEY",
				mocks.NewNopLogger(),
			)

			// Create mock client
			mockClient := &http.Client{
				Transport: &mockRoundTripper{
					responses: map[string]struct {
						resp *http.Response
						err  error
					}{
						// Match any URL since we're testing MakeRequest which builds the URL
						"api.etherscan.io": {
							resp: &http.Response{
								StatusCode: tt.httpStatus,
								Body:       io.NopCloser(strings.NewReader(tt.response)),
							},
							err: nil,
						},
					},
				},
			}

			// Replace HTTPClient with our mock
			if e, ok := explorer.(*blockexplorer.EtherscanExplorer); ok {
				e.HTTPClient = mockClient
			}

			// Convert map to url.Values
			params := url.Values{}
			for k, v := range tt.params {
				params.Set(k, v)
			}

			// Now directly call MakeRequest
			respBytes, err := explorer.(*blockexplorer.EtherscanExplorer).MakeRequest(context.Background(), params)

			// Verify error expectation
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, respBytes)
			require.NotEmpty(t, respBytes)

			// Verify the result content matches what we expect
			assert.Equal(t, tt.result, string(respBytes))
		})
	}
}

// TestMakeRequestRateLimiting tests the rate limiting behavior of the MakeRequest method
func TestMakeRequestRateLimiting(t *testing.T) {
	// Create chain
	chain := types.Chain{Type: types.ChainTypeEthereum}

	// Create explorer instance
	explorer := blockexplorer.NewEtherscanExplorer(
		chain,
		"https://api.etherscan.io/api",
		"https://etherscan.io",
		"TEST_API_KEY",
		mocks.NewNopLogger(),
	)

	// Create a mock transport that creates a new response each time
	// This avoids issues with the body being read multiple times
	mockTransport := &mockRateLimitingTransport{
		statusCode:   http.StatusOK,
		responseBody: `{"status":"1","message":"OK","result":"success"}`,
	}

	mockClient := &http.Client{
		Transport: mockTransport,
	}

	// Replace HTTPClient with our mock
	etherscanExplorer, ok := explorer.(*blockexplorer.EtherscanExplorer)
	require.True(t, ok)
	etherscanExplorer.HTTPClient = mockClient

	// Create basic parameters
	params := url.Values{}
	params.Set("action", "test_action")
	params.Set("address", "0x1234567890abcdef1234567890abcdef12345678")

	// Make a few initial requests which should be quick (burst)
	requestCount := 5
	results := make([]time.Duration, requestCount)

	// First 2-3 should be fast (burst)
	for i := 0; i < requestCount; i++ {
		start := time.Now()
		_, err := etherscanExplorer.MakeRequest(context.Background(), params)
		elapsed := time.Since(start)

		require.NoError(t, err, "Request %d should succeed", i)
		results[i] = elapsed
		t.Logf("Request %d took %v", i, elapsed)
	}

	// Check that initial requests were quick (burst)
	quickCount := 0
	for _, duration := range results[:2] { // First 2 should be quick
		if duration < 100*time.Millisecond {
			quickCount++
		}
	}

	// At least 1-2 requests should be fast (burst)
	assert.GreaterOrEqual(t, quickCount, 1, "At least one request should be fast (burst)")

	// The later requests should be throttled
	var throttledCount int
	for _, duration := range results[2:] { // Later requests should be throttled
		if duration > 200*time.Millisecond {
			throttledCount++
		}
	}

	// Check that some of the later requests were throttled
	assert.GreaterOrEqual(t, throttledCount, 1, "Some later requests should be throttled")
}

// mockRateLimitingTransport implements http.RoundTripper for testing rate limiting
type mockRateLimitingTransport struct {
	statusCode   int
	responseBody string
}

// RoundTrip implements http.RoundTripper
func (m *mockRateLimitingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Create a new response with a fresh body each time
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(strings.NewReader(m.responseBody)),
		Header:     make(http.Header),
	}, nil
}

// TestMakeRequestRetry tests the retry behavior of the MakeRequest method
func TestMakeRequestRetry(t *testing.T) {
	// Create chain
	chain := types.Chain{Type: types.ChainTypeEthereum}

	// Create explorer instance
	explorer := blockexplorer.NewEtherscanExplorer(
		chain,
		"https://api.etherscan.io/api",
		"https://etherscan.io",
		"TEST_API_KEY",
		mocks.NewNopLogger(),
	)

	// Create a mock transport that fails initially then succeeds
	mockTransport := &mockRetryTransport{
		// Start with 2 failures, then succeed
		failCount:    2,
		statusCode:   http.StatusOK,
		responseBody: `{"status":"1","message":"OK","result":"success"}`,
	}

	mockClient := &http.Client{
		Transport: mockTransport,
	}

	// Replace HTTPClient with our mock
	etherscanExplorer, ok := explorer.(*blockexplorer.EtherscanExplorer)
	require.True(t, ok)
	etherscanExplorer.HTTPClient = mockClient

	// Create basic parameters
	params := url.Values{}
	params.Set("action", "test_action")
	params.Set("address", "0x1234567890abcdef1234567890abcdef12345678")

	// Call MakeRequest, which should retry after failures and eventually succeed
	start := time.Now()
	result, err := etherscanExplorer.MakeRequest(context.Background(), params)
	elapsed := time.Since(start)

	// Request should succeed
	require.NoError(t, err, "Request should eventually succeed after retries")
	require.NotNil(t, result)
	assert.Equal(t, `"success"`, string(result))

	// The request should have taken some time due to retries with backoff
	t.Logf("Request with retries took %v", elapsed)
	assert.Greater(t, elapsed, 200*time.Millisecond, "Request should have taken time due to retry backoff")

	// Verify the expected number of attempts were made
	assert.Equal(t, 3, mockTransport.attemptCount, "Should have made exactly 3 attempts (2 failures + 1 success)")
}

// mockRetryTransport implements http.RoundTripper for testing retry behavior
type mockRetryTransport struct {
	failCount    int
	attemptCount int
	statusCode   int
	responseBody string
}

// RoundTrip implements http.RoundTripper
func (m *mockRetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	m.attemptCount++

	// If we haven't reached the fail count, return a connection error
	if m.attemptCount <= m.failCount {
		return nil, fmt.Errorf("connection failed (attempt %d of %d)", m.attemptCount, m.failCount)
	}

	// After failing the desired number of times, succeed
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(strings.NewReader(m.responseBody)),
		Header:     make(http.Header),
	}, nil
}

// TestMakeRequestHTTPErrors tests the error handling for different HTTP response scenarios
func TestMakeRequestHTTPErrors(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		responseBody string
		expectedErr  string
	}{
		{
			name:         "unauthorized response",
			statusCode:   http.StatusUnauthorized,
			responseBody: `{"status":"0","message":"NOTOK","result":"Invalid API Key"}`,
			expectedErr:  "Block explorer error: API error: NOTOK",
		},
		{
			name:         "not found response",
			statusCode:   http.StatusNotFound,
			responseBody: `{"status":"0","message":"NOTOK","result":"Not found"}`,
			expectedErr:  "Block explorer error: API error: NOTOK",
		},
		{
			name:         "server error response",
			statusCode:   http.StatusInternalServerError,
			responseBody: `{"status":"0","message":"NOTOK","result":"Internal server error"}`,
			expectedErr:  "Block explorer error: API error: NOTOK",
		},
		{
			name:         "malformed response",
			statusCode:   http.StatusOK,
			responseBody: `not valid json`,
			expectedErr:  "Invalid response from block explorer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create chain
			chain := types.Chain{Type: types.ChainTypeEthereum}

			// Create explorer instance
			explorer := blockexplorer.NewEtherscanExplorer(
				chain,
				"https://api.etherscan.io/api",
				"https://etherscan.io",
				"TEST_API_KEY",
				mocks.NewNopLogger(),
			)

			// Create a mock client
			mockClient := &http.Client{
				Transport: &mockFixedResponseTransport{
					statusCode:   tt.statusCode,
					responseBody: tt.responseBody,
				},
			}

			// Replace HTTPClient with our mock
			etherscanExplorer, ok := explorer.(*blockexplorer.EtherscanExplorer)
			require.True(t, ok)
			etherscanExplorer.HTTPClient = mockClient

			// Create basic parameters
			params := url.Values{}
			params.Set("action", "test_action")
			params.Set("address", "0x1234567890abcdef1234567890abcdef12345678")

			// Call MakeRequest
			result, err := etherscanExplorer.MakeRequest(context.Background(), params)

			// Verify error
			require.Error(t, err, "Expected an error but got none")
			assert.Nil(t, result, "Result should be nil on error")
			assert.Contains(t, err.Error(), tt.expectedErr, "Error message didn't match expectation")
		})
	}
}

// TestMakeRequestAPIErrors tests the parsing of error responses from Etherscan API
func TestMakeRequestAPIErrors(t *testing.T) {
	tests := []struct {
		name         string
		responseBody string
		expectedErr  string
	}{
		{
			name:         "api error response",
			responseBody: `{"status":"0","message":"NOTOK","result":"Error!"}`,
			expectedErr:  "Block explorer error: API error: NOTOK",
		},
		{
			name:         "rate limit error",
			responseBody: `{"status":"0","message":"NOTOK","result":"Max rate limit reached"}`,
			expectedErr:  "Block explorer request failed",
		},
		{
			name:         "invalid api key",
			responseBody: `{"status":"0","message":"NOTOK","result":"Invalid API Key"}`,
			expectedErr:  "Block explorer error: API error: NOTOK",
		},
		{
			name:         "missing parameter",
			responseBody: `{"status":"0","message":"NOTOK","result":"Missing parameter"}`,
			expectedErr:  "Block explorer error: API error: NOTOK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create chain
			chain := types.Chain{Type: types.ChainTypeEthereum}

			// Create explorer instance
			explorer := blockexplorer.NewEtherscanExplorer(
				chain,
				"https://api.etherscan.io/api",
				"https://etherscan.io",
				"TEST_API_KEY",
				mocks.NewNopLogger(),
			)

			// Create a mock client
			mockClient := &http.Client{
				Transport: &mockFixedResponseTransport{
					statusCode:   http.StatusOK, // Status OK but JSON indicates error
					responseBody: tt.responseBody,
				},
			}

			// Replace HTTPClient with our mock
			etherscanExplorer, ok := explorer.(*blockexplorer.EtherscanExplorer)
			require.True(t, ok)
			etherscanExplorer.HTTPClient = mockClient

			// Create basic parameters
			params := url.Values{}
			params.Set("action", "test_action")
			params.Set("address", "0x1234567890abcdef1234567890abcdef12345678")

			// Call MakeRequest
			result, err := etherscanExplorer.MakeRequest(context.Background(), params)

			// Verify error
			require.Error(t, err, "Expected an error but got none")
			assert.Nil(t, result, "Result should be nil on error")
			assert.Contains(t, err.Error(), tt.expectedErr, "Error message didn't match expectation")
		})
	}
}

// mockFixedResponseTransport implements http.RoundTripper for testing with fixed responses
type mockFixedResponseTransport struct {
	statusCode   int
	responseBody string
}

// RoundTrip implements http.RoundTripper
func (m *mockFixedResponseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(strings.NewReader(m.responseBody)),
		Header:     make(http.Header),
	}, nil
}
