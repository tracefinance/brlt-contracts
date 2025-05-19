package mocks

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/stretchr/testify/mock"
)

// MockHTTPClient implements http.Client for testing purposes
type MockHTTPClient struct {
	mock.Mock

	// Mutex to protect responses map
	mutex sync.RWMutex

	// Map of URL pattern -> response
	responses map[string]*http.Response

	// Map of URL pattern -> error
	errors map[string]error

	// Default response for unmatched URLs
	defaultResponse *http.Response
	defaultError    error

	// Count requests to track rate limiting simulation
	requestCount int
}

// NewMockHTTPClient creates a new mock HTTP client for testing
func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		responses: make(map[string]*http.Response),
		errors:    make(map[string]error),
	}
}

// Do implements the http.Client.Do method
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)

	// Record the call for verification
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.requestCount++

	// Check for specific URL pattern matches
	for pattern, resp := range m.responses {
		if strings.Contains(req.URL.String(), pattern) {
			if err := m.errors[pattern]; err != nil {
				return nil, err
			}
			return resp, nil
		}
	}

	// Fall back to the explicit mock expectations
	if args.Get(0) != nil {
		return args.Get(0).(*http.Response), args.Error(1)
	}

	// Default response if no mock expectation matched
	if m.defaultResponse != nil {
		return m.defaultResponse, m.defaultError
	}

	// Default not found response
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(bytes.NewBufferString("not found")),
	}, nil
}

// AddResponse registers a response for a URL pattern
func (m *MockHTTPClient) AddResponse(urlPattern string, statusCode int, body string) *MockHTTPClient {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.responses[urlPattern] = &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
	return m
}

// AddResponseWithHeader registers a response with headers for a URL pattern
func (m *MockHTTPClient) AddResponseWithHeader(urlPattern string, statusCode int, body string, headers map[string]string) *MockHTTPClient {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	header := http.Header{}
	for k, v := range headers {
		header.Add(k, v)
	}

	m.responses[urlPattern] = &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     header,
	}
	return m
}

// AddError registers an error for a URL pattern
func (m *MockHTTPClient) AddError(urlPattern string, err error) *MockHTTPClient {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.errors[urlPattern] = err
	return m
}

// SetDefaultResponse sets a default response for unmatched URLs
func (m *MockHTTPClient) SetDefaultResponse(statusCode int, body string) *MockHTTPClient {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.defaultResponse = &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
	return m
}

// SetDefaultError sets a default error for unmatched URLs
func (m *MockHTTPClient) SetDefaultError(err error) *MockHTTPClient {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.defaultError = err
	return m
}

// GetRequestCount returns the number of requests made
func (m *MockHTTPClient) GetRequestCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.requestCount
}

// ResetRequestCount resets the request counter
func (m *MockHTTPClient) ResetRequestCount() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.requestCount = 0
}

// Clear clears all registered responses and errors
func (m *MockHTTPClient) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.responses = make(map[string]*http.Response)
	m.errors = make(map[string]error)
	m.defaultResponse = nil
	m.defaultError = nil
	m.requestCount = 0
}

// HTTPClient returns an http.Client that uses this mock client for its transport
func (m *MockHTTPClient) HTTPClient() *http.Client {
	return &http.Client{
		Transport: &mockTransport{m},
	}
}

// mockTransport is an http.RoundTripper that delegates to the MockHTTPClient
type mockTransport struct {
	mock *MockHTTPClient
}

// RoundTrip implements http.RoundTripper
func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.mock.Do(req)
}
