package matchers

import (
	"strings"

	"github.com/stretchr/testify/mock"
)

// AddressMatcher returns a custom matcher for Ethereum addresses to handle case-insensitivity
// This is useful in tests where addresses might have different case representations
// (e.g., when using EIP-55 checksum format vs. lowercase)
func AddressMatcher(expectedAddr string) interface{} {
	return mock.MatchedBy(func(actualAddr string) bool {
		return strings.EqualFold(expectedAddr, actualAddr)
	})
}
