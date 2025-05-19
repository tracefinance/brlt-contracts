package fixtures

import (
	"fmt"
)

// ERC20ABI represents a standard ERC20 ABI JSON string for testing
const ERC20ABI = `[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"anonymous":false,"inputs":[{"indexed":true,"name":"owner","type":"address"},{"indexed":true,"name":"spender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`

// GetContractABIResponse returns a sample Etherscan API response for getabi action
func GetContractABIResponse(abi string) string {
	return fmt.Sprintf(`{"status":"1","message":"OK","result":"%s"}`, abi)
}

// GetContractSourceResponse returns a sample Etherscan API response for getsourcecode action
func GetContractSourceResponse(name, source, compiler string) string {
	return fmt.Sprintf(`{"status":"1","message":"OK","result":[{"SourceCode":"%s","ContractName":"%s","CompilerVersion":"%s"}]}`, source, name, compiler)
}

// GetNormalTxListResponse returns a sample Etherscan API response for normal transaction list
func GetNormalTxListResponse(transactions []string) string {
	if len(transactions) == 0 {
		return `{"status":"1","message":"OK","result":[]}`
	}

	txsStr := ""
	for i, tx := range transactions {
		if i > 0 {
			txsStr += ","
		}
		txsStr += tx
	}

	return fmt.Sprintf(`{"status":"1","message":"OK","result":[%s]}`, txsStr)
}

// GetSampleNormalTransaction returns a sample normal transaction JSON for testing
func GetSampleNormalTransaction(hash, fromAddr, toAddr string, value, gasUsed, blockNumber, timestamp string) string {
	return fmt.Sprintf(`{
		"hash": "%s",
		"from": "%s",
		"to": "%s",
		"value": "%s",
		"gas": "21000",
		"gasPrice": "20000000000",
		"gasUsed": "%s",
		"nonce": "42",
		"blockHash": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		"blockNumber": "%s",
		"transactionIndex": "1",
		"timeStamp": "%s",
		"isError": "0",
		"contractAddress": ""
	}`, hash, fromAddr, toAddr, value, gasUsed, blockNumber, timestamp)
}

// GetAPIErrorResponse returns a sample Etherscan API error response
func GetAPIErrorResponse(message string) string {
	return fmt.Sprintf(`{"status":"0","message":"NOTOK","result":"%s"}`, message)
}

// GetRateLimitErrorResponse returns a sample Etherscan API rate limit error response
func GetRateLimitErrorResponse() string {
	return `{"status":"0","message":"NOTOK","result":"Max rate limit reached"}`
}
