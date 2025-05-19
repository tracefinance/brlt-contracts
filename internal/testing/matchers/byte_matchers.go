package matchers

import (
	"bytes"

	"github.com/stretchr/testify/mock"
)

// EmptyBytesMatcher returns a matcher that matches empty byte arrays
func EmptyBytesMatcher() interface{} {
	return mock.MatchedBy(func(b []byte) bool {
		return len(b) == 0
	})
}

// BytesMatcher returns a matcher that matches byte arrays with the exact content
func BytesMatcher(expected []byte) interface{} {
	return mock.MatchedBy(func(actual []byte) bool {
		return bytes.Equal(expected, actual)
	})
}

// AnyBytesMatcher returns a matcher that matches any byte array
func AnyBytesMatcher() interface{} {
	return mock.MatchedBy(func(b []byte) bool {
		return true
	})
}
