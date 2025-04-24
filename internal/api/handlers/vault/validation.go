package vault

import (
	"fmt"
	"regexp"

	"vault0/internal/errors"
)

const (
	// Vault name validation
	minVaultNameLength = 3
	maxVaultNameLength = 50

	// Signer validation
	minSignersPerVault = 1
	maxSignersPerVault = 10
)

// Ethereum address regex
var ethereumAddressRegex = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)

// ValidateCreateVaultRequest validates the CreateVaultRequest
func ValidateCreateVaultRequest(req *CreateVaultRequest) error {
	if len(req.Name) < minVaultNameLength || len(req.Name) > maxVaultNameLength {
		return errors.NewValidationError(map[string]any{
			"name": fmt.Sprintf("Vault name must be between %d and %d characters", minVaultNameLength, maxVaultNameLength),
		})
	}

	if len(req.SignerAddresses) < minSignersPerVault {
		return errors.NewValidationError(map[string]any{
			"signer_addresses": fmt.Sprintf("Must have at least %d signer", minSignersPerVault),
		})
	}

	if len(req.SignerAddresses) > maxSignersPerVault {
		return errors.NewValidationError(map[string]any{
			"signer_addresses": fmt.Sprintf("Must have at most %d signers", maxSignersPerVault),
		})
	}

	if req.SignatureThreshold > len(req.SignerAddresses) {
		return errors.NewValidationError(map[string]any{
			"signature_threshold": "Cannot be greater than number of signers",
		})
	}

	// Validate recovery address format
	if !IsValidEthereumAddress(req.RecoveryAddress) {
		return errors.NewValidationError(map[string]any{
			"recovery_address": "Invalid Ethereum address format",
		})
	}

	// Validate signer addresses format
	for i, addr := range req.SignerAddresses {
		if !IsValidEthereumAddress(addr) {
			return errors.NewValidationError(map[string]any{
				"signer_addresses": fmt.Sprintf("Invalid Ethereum address at index %d", i),
			})
		}
	}

	// Validate whitelisted tokens if provided
	if len(req.WhitelistedTokens) > 0 {
		for i, token := range req.WhitelistedTokens {
			if !IsValidEthereumAddress(token) {
				return errors.NewValidationError(map[string]any{
					"whitelisted_tokens": fmt.Sprintf("Invalid token address at index %d", i),
				})
			}
		}
	}

	return nil
}

// ValidateTokenAddress validates a token address
func ValidateTokenAddress(address string) error {
	if address == "" {
		return errors.NewValidationError(map[string]any{
			"address": "Token address cannot be empty",
		})
	}

	if !IsValidEthereumAddress(address) {
		return errors.NewValidationError(map[string]any{
			"address": "Invalid token address format",
		})
	}

	return nil
}

// IsValidEthereumAddress checks if the address is a valid Ethereum address
func IsValidEthereumAddress(address string) bool {
	return ethereumAddressRegex.MatchString(address)
}
