package types

import "math/big"

// MultiSigMethodSignature represents a method name in the MultiSigWallet contract
type MultiSigMethodSignature string

// MultiSigEventSignature represents an event signature in the MultiSigWallet contract
type MultiSigEventSignature string

// MultiSigContractName is the standard name for the vault contract
const MultiSigContractName = "MultiSigWallet"

// MultiSigWallet contract method signatures
const (
	MultiSigAddSupportedTokenMethod            MultiSigMethodSignature = "addSupportedToken(address)"
	MultiSigRemoveSupportedTokenMethod         MultiSigMethodSignature = "removeSupportedToken(address)"
	MultiSigRequestRecoveryMethod              MultiSigMethodSignature = "requestRecovery()"
	MultiSigCancelRecoveryMethod               MultiSigMethodSignature = "cancelRecovery()"
	MultiSigExecuteRecoveryMethod              MultiSigMethodSignature = "executeRecovery()"
	MultiSigRequestWithdrawalMethod            MultiSigMethodSignature = "requestWithdrawal(address,uint256,address)"
	MultiSigSignWithdrawalMethod               MultiSigMethodSignature = "signWithdrawal(bytes32)"
	MultiSigProposeRecoveryAddressChangeMethod MultiSigMethodSignature = "proposeRecoveryAddressChange(address)"
	MultiSigSignRecoveryAddressChangeMethod    MultiSigMethodSignature = "signRecoveryAddressChange(bytes32)"
	MultiSigRecoverNonSupportedTokenMethod     MultiSigMethodSignature = "recoverNonSupportedToken(address,address)"
	MultiSigGetSupportedTokensMethod           MultiSigMethodSignature = "getSupportedTokens()"
	MultiSigGetSignersMethod                   MultiSigMethodSignature = "getSigners()"
)

// MultiSigWallet contract event signatures (hashed topics are used for filtering)
const (
	MultiSigDepositedEventSig                           MultiSigEventSignature = "Deposited(address,address,uint256)"
	MultiSigWithdrawalRequestedEventSig                 MultiSigEventSignature = "WithdrawalRequested(bytes32,address,uint256,address,uint256)"
	MultiSigWithdrawalSignedEventSig                    MultiSigEventSignature = "WithdrawalSigned(bytes32,address)"
	MultiSigWithdrawalExecutedEventSig                  MultiSigEventSignature = "WithdrawalExecuted(bytes32,address,uint256,address)"
	MultiSigRecoveryRequestedEventSig                   MultiSigEventSignature = "RecoveryRequested(uint256)"
	MultiSigRecoveryCancelledEventSig                   MultiSigEventSignature = "RecoveryCancelled()"
	MultiSigRecoveryExecutedEventSig                    MultiSigEventSignature = "RecoveryExecuted(address,uint256)"
	MultiSigRecoveryCompletedEventSig                   MultiSigEventSignature = "RecoveryCompleted()"
	MultiSigTokenSupportedEventSig                      MultiSigEventSignature = "TokenSupported(address)"
	MultiSigTokenRemovedEventSig                        MultiSigEventSignature = "TokenRemoved(address)"
	MultiSigNonSupportedTokenRecoveredEventSig          MultiSigEventSignature = "NonSupportedTokenRecovered(address,uint256,address)"
	MultiSigTokenWhitelistedEventSig                    MultiSigEventSignature = "TokenWhitelisted(address)"
	MultiSigRecoveryAddressChangeProposedEventSig       MultiSigEventSignature = "RecoveryAddressChangeProposed(address,address,bytes32)"
	MultiSigRecoveryAddressChangeSignatureAddedEventSig MultiSigEventSignature = "RecoveryAddressChangeSignatureAdded(address,bytes32)"
	MultiSigRecoveryAddressChangedEventSig              MultiSigEventSignature = "RecoveryAddressChanged(address,address,bytes32)"
)

// MultiSig transaction type constants
const (
	// TransactionTypeMultiSigWithdrawalRequest indicates a transaction is a MultiSig withdrawal request
	TransactionTypeMultiSigWithdrawalRequest TransactionType = "multisig_withdrawal_request"

	// TransactionTypeMultiSigSignWithdrawal indicates a transaction is a signature for a MultiSig withdrawal
	TransactionTypeMultiSigSignWithdrawal TransactionType = "multisig_sign_withdrawal"

	// TransactionTypeMultiSigExecuteWithdrawal indicates a transaction is execution of a MultiSig withdrawal
	TransactionTypeMultiSigExecuteWithdrawal TransactionType = "multisig_execute_withdrawal"

	// TransactionTypeMultiSigAddSupportedToken indicates a transaction is adding a supported token to the MultiSig
	TransactionTypeMultiSigAddSupportedToken TransactionType = "multisig_add_token"

	// TransactionTypeMultiSigRecoveryRequest indicates a transaction is a MultiSig recovery request
	TransactionTypeMultiSigRecoveryRequest TransactionType = "multisig_recovery_request"

	// TransactionTypeMultiSigCancelRecovery indicates a transaction is cancelling a MultiSig recovery
	TransactionTypeMultiSigCancelRecovery TransactionType = "multisig_cancel_recovery"

	// TransactionTypeMultiSigExecuteRecovery indicates a transaction is executing a MultiSig recovery
	TransactionTypeMultiSigExecuteRecovery TransactionType = "multisig_execute_recovery"

	// TransactionTypeMultiSigProposeRecoveryAddressChange indicates a transaction is proposing a recovery address change
	TransactionTypeMultiSigProposeRecoveryAddressChange TransactionType = "multisig_propose_recovery_address_change"

	// TransactionTypeMultiSigSignRecoveryAddressChange indicates a transaction is signing a recovery address change
	TransactionTypeMultiSigSignRecoveryAddressChange TransactionType = "multisig_sign_recovery_address_change"
)

// MultiSigWithdrawalRequest represents a withdrawal request transaction in the MultiSig wallet
type MultiSigWithdrawalRequest struct {
	// Embeds the core transaction details
	BaseTransaction
	// Token is the address of the token to withdraw
	Token string
	// Amount is the amount of tokens to withdraw
	Amount *big.Int
	// Recipient is the address to receive the withdrawn tokens
	Recipient string
	// WithdrawalNonce is the unique identifier for this withdrawal request
	WithdrawalNonce uint64
}

// MultiSigSignWithdrawal represents a transaction signing a withdrawal request in the MultiSig wallet
type MultiSigSignWithdrawal struct {
	// Embeds the core transaction details
	BaseTransaction
	// RequestID is the unique identifier of the withdrawal request being signed
	RequestID [32]byte
}

// MultiSigExecuteWithdrawal represents a transaction executing a withdrawal in the MultiSig wallet
type MultiSigExecuteWithdrawal struct {
	// Embeds the core transaction details
	BaseTransaction
	// RequestID is the unique identifier of the withdrawal request being executed
	RequestID [32]byte
}

// MultiSigAddSupportedToken represents a transaction adding a supported token to the MultiSig wallet
type MultiSigAddSupportedToken struct {
	// Embeds the core transaction details
	BaseTransaction
	// Token is the address of the token being added as supported
	Token string
}

// MultiSigRecoveryRequest represents a transaction requesting recovery of the MultiSig wallet
type MultiSigRecoveryRequest struct {
	// Embeds the core transaction details
	BaseTransaction
}

// MultiSigCancelRecovery represents a transaction cancelling a recovery request for the MultiSig wallet
type MultiSigCancelRecovery struct {
	// Embeds the core transaction details
	BaseTransaction
}

// MultiSigExecuteRecovery represents a transaction executing recovery for the MultiSig wallet
type MultiSigExecuteRecovery struct {
	// Embeds the core transaction details
	BaseTransaction
}

// MultiSigProposeRecoveryAddressChange represents a transaction proposing a change to the recovery address
type MultiSigProposeRecoveryAddressChange struct {
	// Embeds the core transaction details
	BaseTransaction
	// NewRecoveryAddress is the proposed new recovery address
	NewRecoveryAddress string
}

// MultiSigSignRecoveryAddressChange represents a transaction signing a recovery address change
type MultiSigSignRecoveryAddressChange struct {
	// Embeds the core transaction details
	BaseTransaction
	// ProposalID is the unique identifier of the recovery address change proposal
	ProposalID [32]byte
}
