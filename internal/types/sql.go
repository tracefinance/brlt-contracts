package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math/big"
)

// BigInt is a wrapper around big.Int that implements database interfaces
// for seamless integration with SQL storage and retrieval.
type BigInt struct {
	*big.Int
}

// NewBigInt creates a new BigInt from a *big.Int.
func NewBigInt(i *big.Int) BigInt {
	if i == nil {
		return BigInt{nil}
	}
	return BigInt{new(big.Int).Set(i)}
}

// ZeroBigInt returns a BigInt with a value of 0.
func ZeroBigInt() BigInt {
	return BigInt{big.NewInt(0)}
}

// NewBigIntFromString creates a new BigInt from a string representation.
// Returns an error if the string cannot be parsed.
func NewBigIntFromString(s string) (BigInt, error) {
	if s == "" {
		return BigInt{nil}, nil
	}

	i := new(big.Int)
	_, ok := i.SetString(s, 10)
	if !ok {
		return BigInt{nil}, fmt.Errorf("failed to parse %q as big.Int", s)
	}

	return BigInt{i}, nil
}

// String returns the string representation of the BigInt.
// Returns an empty string if the BigInt is nil.
func (b BigInt) String() string {
	if b.Int == nil {
		return ""
	}
	return b.Int.String()
}

// Scan implements the sql.Scanner interface for database deserialization.
func (b *BigInt) Scan(value any) error {
	if value == nil {
		b.Int = nil
		return nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			b.Int = nil
			return nil
		}
		if b.Int == nil {
			b.Int = new(big.Int)
		}
		_, ok := b.Int.SetString(v, 10)
		if !ok {
			return fmt.Errorf("failed to parse %q as big.Int", v)
		}
		return nil
	case []byte:
		if len(v) == 0 {
			b.Int = nil
			return nil
		}
		if b.Int == nil {
			b.Int = new(big.Int)
		}
		_, ok := b.Int.SetString(string(v), 10)
		if !ok {
			return fmt.Errorf("failed to parse %q as big.Int", string(v))
		}
		return nil
	case int64:
		if b.Int == nil {
			b.Int = new(big.Int)
		}
		b.Int.SetInt64(v)
		return nil
	default:
		return fmt.Errorf("unsupported Scan type: %T", value)
	}
}

// Value implements the driver.Valuer interface for database serialization.
func (b BigInt) Value() (driver.Value, error) {
	if b.Int == nil {
		return nil, nil
	}
	return b.String(), nil
}

// IsZero returns true if the BigInt is zero or nil.
func (b BigInt) IsZero() bool {
	return b.Int == nil || b.Int.Sign() == 0
}

// ToBigInt converts BigInt to *big.Int.
// Returns nil if BigInt is nil, otherwise returns a copy of the wrapped *big.Int.
func (b BigInt) ToBigInt() *big.Int {
	if b.Int == nil {
		return nil
	}
	return new(big.Int).Set(b.Int)
}

// JSONMap is a map[string]string that implements sql.Scanner and driver.Valuer interfaces
// for seamless integration with SQL storage and retrieval as JSON.
type JSONMap map[string]string

// NewJSONMap creates a new JSONMap from a map[string]string.
func NewJSONMap(m map[string]string) JSONMap {
	if m == nil {
		return JSONMap{}
	}
	return JSONMap(m)
}

// Scan implements the sql.Scanner interface for database deserialization.
func (m *JSONMap) Scan(value any) error {
	if value == nil {
		*m = JSONMap{}
		return nil
	}

	var jsonStr string
	switch v := value.(type) {
	case string:
		jsonStr = v
	case []byte:
		jsonStr = string(v)
	default:
		return fmt.Errorf("unsupported Scan type for JSONMap: %T", value)
	}

	if jsonStr == "" {
		*m = JSONMap{}
		return nil
	}

	// Unmarshal JSON into the map
	var result map[string]string
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return err
	}

	*m = JSONMap(result)
	return nil
}

// Value implements the driver.Valuer interface for database serialization.
func (m JSONMap) Value() (driver.Value, error) {
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

// Map returns the underlying map[string]string.
func (m JSONMap) Map() map[string]string {
	if m == nil {
		return map[string]string{}
	}
	return map[string]string(m)
}

// JSONArray is a slice of strings that implements sql.Scanner and driver.Valuer interfaces
// for seamless integration with SQL storage and retrieval as JSON.
type JSONArray []string

// NewJSONArray creates a new JSONArray from a []string.
func NewJSONArray(s []string) JSONArray {
	if s == nil {
		// Return an empty slice instead of nil for consistency with JSON marshalling
		return JSONArray{}
	}
	return JSONArray(s)
}

// Scan implements the sql.Scanner interface for database deserialization.
func (a *JSONArray) Scan(value any) error {
	if value == nil {
		*a = JSONArray{}
		return nil
	}

	var jsonStr string
	switch v := value.(type) {
	case string:
		jsonStr = v
	case []byte:
		jsonStr = string(v)
	default:
		return fmt.Errorf("unsupported Scan type for JSONArray: %T", value)
	}

	if jsonStr == "" || jsonStr == "[]" || jsonStr == "{}" { // Consider empty JSON representations
		*a = JSONArray{}
		return nil
	}

	// Unmarshal JSON into the slice
	var result []string
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return fmt.Errorf("failed to unmarshal JSON to JSONArray: %w", err)
	}

	*a = JSONArray(result)
	return nil
}

// Value implements the driver.Valuer interface for database serialization.
func (a JSONArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		// Store empty or nil slice as JSON '[]'
		return "[]", nil
	}

	// Marshal the slice to JSON
	jsonData, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	return string(jsonData), nil
}

// Slice returns the underlying []string.
func (a JSONArray) Slice() []string {
	if a == nil {
		return []string{}
	}
	return []string(a)
}
