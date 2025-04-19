package keystore

import (
	"vault0/internal/types"
)

// KeyFilter defines filtering options for key listing
type KeyFilter struct {
	KeyType *types.KeyType
	Tags    map[string]string
}
