package types

import (
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"maps"

	"github.com/ethereum/go-ethereum/common"
)

// TxMetadata represents the metadata associated with a transaction.
// It uses string values to store various data points extracted during mapping or enrichment.
type TxMetadata map[string]string

// Scan implements the sql.Scanner interface for database deserialization.
func (m *TxMetadata) Scan(value any) error {
	if value == nil {
		*m = TxMetadata{}
		return nil
	}

	var jsonStr string
	switch v := value.(type) {
	case string:
		jsonStr = v
	case []byte:
		jsonStr = string(v)
	default:
		return fmt.Errorf("unsupported Scan type for TxMetadata: %T", value)
	}

	if jsonStr == "" {
		*m = TxMetadata{}
		return nil
	}

	// Unmarshal JSON into the map
	var result map[string]string
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return err
	}

	*m = TxMetadata(result)
	return nil
}

// Value implements the driver.Valuer interface for database serialization.
func (m TxMetadata) Value() (driver.Value, error) {
	if len(m) == 0 {
		return "", nil
	}

	// Marshal the map to JSON
	jsonData, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return string(jsonData), nil
}

// GetString safely retrieves a string value for the given key.
// Returns the string value and true if the key exists, otherwise returns an empty string and false.
func (m TxMetadata) GetString(key string) (string, bool) {
	val, ok := m[key]
	return val, ok
}

// Contains checks if the metadata map contains the given key.
// Returns true if the key exists, false otherwise.
func (m TxMetadata) Contains(key string) bool {
	_, ok := m[key]
	return ok
}

// GetAddress safely retrieves and validates an address string for the given key.
// Returns the common.Address and true if the key exists and the value is a valid hex address,
// otherwise returns a zero address and false.
func (m TxMetadata) GetAddress(key string) (common.Address, bool) {
	val, ok := m[key]
	if !ok || !common.IsHexAddress(val) {
		return common.Address{}, false
	}
	return common.HexToAddress(val), true
}

// GetBigInt safely retrieves and converts a string value to a BigInt for the given key.
// Returns the BigInt value and true if the key exists and the value is a valid integer string,
// otherwise returns nil and false.
func (m TxMetadata) GetBigInt(key string) (*BigInt, bool) {
	valStr, ok := m[key]
	if !ok {
		return nil, false
	}
	bigIntValue, ok := new(big.Int).SetString(valStr, 10)
	if !ok {
		return nil, false
	}
	// Convert standard *big.Int to custom *types.BigInt using the constructor and return a pointer
	newBigInt := NewBigInt(bigIntValue)
	return &newBigInt, true
}

// GetBytes32 safely retrieves and converts a hex string to a [32]byte array for the given key.
// Returns the [32]byte array and true if the key exists, the value is a valid hex string
// without the "0x" prefix, and has the correct length (64 hex chars).
// Otherwise returns a zero byte array and false.
func (m TxMetadata) GetBytes32(key string) ([32]byte, bool) {
	valHex, ok := m[key]
	if !ok {
		return [32]byte{}, false
	}
	// Ensure it's a hex string of the correct length (64 chars for 32 bytes)
	if len(valHex) != 64 {
		return [32]byte{}, false
	}
	decoded, err := hex.DecodeString(valHex)
	if err != nil || len(decoded) != 32 {
		return [32]byte{}, false
	}
	var result [32]byte
	copy(result[:], decoded)
	return result, true
}

// GetUint64 safely retrieves and converts a string value to a uint64 for the given key.
// Returns the uint64 value and true if the key exists and the value is a valid non-negative integer string,
// otherwise returns 0 and false.
func (m TxMetadata) GetUint64(key string) (uint64, bool) {
	valStr, ok := m[key]
	if !ok {
		return 0, false
	}
	uintVal, err := strconv.ParseUint(valStr, 10, 64)
	if err != nil {
		return 0, false
	}
	return uintVal, true
}

// GetInt64 safely retrieves and converts a string value to an int64 for the given key.
// Returns the int64 value and true if the key exists and the value is a valid integer string,
// otherwise returns 0 and false.
func (m TxMetadata) GetInt64(key string) (int64, bool) {
	valStr, ok := m[key]
	if !ok {
		return 0, false
	}
	intVal, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		return 0, false
	}
	return intVal, true
}

// GetUint8 safely retrieves and converts a string value to a uint8 for the given key.
// Returns the uint8 value and true if the key exists and the value is a valid integer string
// within the uint8 range (0-255), otherwise returns 0 and false.
func (m TxMetadata) GetUint8(key string) (uint8, bool) {
	valStr, ok := m[key]
	if !ok {
		return 0, false
	}
	// Use ParseUint with bitSize 8 to enforce range
	uintVal, err := strconv.ParseUint(valStr, 10, 8)
	if err != nil {
		return 0, false
	}
	return uint8(uintVal), true
}

// Set adds or updates a key-value pair in the metadata.
// It converts various supported types to their string representation before storing.
func (m TxMetadata) Set(key string, value any) error {
	if m == nil {
		return fmt.Errorf("cannot set value on nil TxMetadata")
	}
	switch v := value.(type) {
	case string:
		m[key] = v
	case common.Address:
		m[key] = v.Hex()
	case *big.Int:
		if v == nil {
			return nil
		}
		m[key] = (*big.Int)(v).String()
	case *BigInt: // Handle our custom BigInt type
		if v == nil {
			return nil
		}
		m[key] = v.String()
	case [32]byte:
		m[key] = hex.EncodeToString(v[:])
	case uint64:
		m[key] = strconv.FormatUint(v, 10)
	case uint8:
		m[key] = strconv.FormatUint(uint64(v), 10)
	case int:
		m[key] = strconv.Itoa(v)
	case int64:
		m[key] = strconv.FormatInt(v, 10)

	case fmt.Stringer: // Support types that implement String()
		m[key] = v.String()
	default:
		return fmt.Errorf("unsupported type for TxMetadata: %T", value)
	}
	return nil
}

// SetAll adds or updates multiple key-value pairs from a map[string]any.
// It iterates through the input map and calls Set for each entry.
// Returns the first error encountered during setting, or nil if all succeed.
func (m TxMetadata) SetAll(data map[string]any) error {
	if m == nil {
		return fmt.Errorf("cannot set values on nil TxMetadata")
	}
	for key, value := range data {
		if err := m.Set(key, value); err != nil {
			return fmt.Errorf("failed to set key '%s': %w", key, err)
		}
	}
	return nil
}

// Copy returns a deep copy of the TxMetadata map.
func (m TxMetadata) Copy() TxMetadata {
	copy := make(TxMetadata)
	maps.Copy(copy, m)
	return copy
}

// Metadata keys for built-in transformers
const (
	// WalletIDMetadaKey is the key for the transformer that extracts wallet ID from metadata
	WalletIDMetadaKey = "wallet_id"

	// VaultIDMetadaKey is the key for the transformer that extracts vault ID from metadata
	VaultIDMetadaKey = "vault_id"

	// TransactionTypeMetadaKey is the key for the transformer that extracts transaction type from metadata
	TransactionTypeMetadaKey = "type"

	// ERC20 specific metadata keys
	ERC20TokenAddressMetadataKey  = "token_address"
	ERC20TokenSymbolMetadataKey   = "token_symbol"
	ERC20TokenDecimalsMetadataKey = "token_decimals"
	ERC20RecipientMetadataKey     = "recipient"
	ERC20AmountMetadataKey        = "amount"

	// ERC721 Specific Metadata Keys
	ERC721TokenAddressMetadataKey = "token_address" // Can be the same as ERC20 as it's generic
	ERC721TokenSymbolMetadataKey  = "token_symbol"
	ERC721TokenNameMetadataKey    = "token_name"
	ERC721TokenIDMetadataKey      = "token_id"
	ERC721RecipientMetadataKey    = "recipient" // Can be the same as ERC20
	ERC721TokenURIMetadataKey     = "token_uri"

	// General EVM Metadata (can be shared or have specific prefixes if needed)
	EVMBlockNumberMetadataKey = "block_number" // Example, if needed more broadly

	// MultiSig specific metadata keys
	MultiSigTokenMetadataKey              = "token"
	MultiSigAmountMetadataKey             = "amount"
	MultiSigRecipientMetadataKey          = "recipient"
	MultiSigWithdrawalNonceMetadataKey    = "withdrawal_nonce"
	MultiSigRequestIDMetadataKey          = "request_id"
	MultiSigNewRecoveryAddressMetadataKey = "new_recovery_address"
	MultiSigProposalIDMetadataKey         = "proposal_id"
)
