package types

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
