package vault

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"vault0/internal/config"
	"vault0/internal/core/contract"
	coreWallet "vault0/internal/core/wallet"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/services/wallet"
	"vault0/internal/types"
)

// Default polling intervals
const (
	defaultDeploymentInterval = 15 * time.Second
	defaultRecoveryInterval   = 60 * time.Second
	// Default recovery timelock (matches contract)
	defaultRecoveryDelay = 72 * time.Hour
)

// Service defines the interface for vault-related business logic operations.
type Service interface {
	MonitorService

	// CreateVault initializes a new vault, including deploying its associated smart contract.
	// It takes the owner wallet ID, vault name, recovery address, initial signers,
	// the required signature quorum, and optionally whitelisted tokens.
	// Parameters:
	//   - ctx: The context for the request.
	//   - walletID: The ID of the user's wallet that will own the vault.
	//   - name: The user-defined name for the vault.
	//   - recoveryAddress: The address designated for vault recovery.
	//   - signers: A list of addresses that are authorized signers for the vault.
	//   - quorum: The minimum number of signatures required to approve transactions.
	//   - whitelistedTokens: Optional list of token addresses initially allowed for transactions.
	// Returns:
	//   - *Vault: The newly created Vault details (initially in 'Deploying' status).
	//   - error: An error if validation fails, contract deployment fails, or DB save fails.
	CreateVault(ctx context.Context, walletID int64, name string, recoveryAddress string, signers []string, quorum int, whitelistedTokens []string) (*Vault, error)
	// GetVaultByID retrieves a vault's details by its unique ID.
	// Parameters:
	//   - ctx: The context for the request.
	//   - vaultID: The unique identifier of the vault to retrieve.
	// Returns:
	//   - *Vault: The details of the requested Vault.
	//   - error: An error if the vault is not found (ErrCodeNotFound) or another DB error occurs.
	GetVaultByID(ctx context.Context, vaultID int64) (*Vault, error)
	// ListVaults retrieves a paginated list of vaults, optionally filtered by criteria.
	// Parameters:
	//   - ctx: The context for the request.
	//   - filter: A VaultFilter struct containing criteria to filter the results.
	//   - limit: The maximum number of vaults to return per page.
	//   - nextToken: A pagination token from a previous response to fetch the next page.
	// Returns:
	//   - *types.Page[*Vault]: A paginated response containing a list of Vaults and a next token.
	//   - error: An error if the database query fails.
	ListVaults(ctx context.Context, filter VaultFilter, limit int, nextToken string) (*types.Page[*Vault], error)
	// UpdateVaultName modifies the name of an existing vault.
	// Parameters:
	//   - ctx: The context for the request.
	//   - vaultID: The ID of the vault to update.
	//   - newName: The new desired name for the vault.
	// Returns:
	//   - *Vault: The updated Vault details.
	//   - error: An error if the vault is not found, the name is invalid, or the DB update fails.
	UpdateVaultName(ctx context.Context, vaultID int64, newName string) (*Vault, error)
	// AddSupportedToken adds a token address to the vault's whitelist on the smart contract.
	// Parameters:
	//   - ctx: The context for the request.
	//   - vaultID: The ID of the vault to modify.
	//   - tokenAddress: The address of the token to add to the whitelist.
	// Returns:
	//   - txHash: The transaction hash of the blockchain operation.
	//   - err: An error if the vault is not found, not active, the address is invalid, or contract execution fails.
	AddSupportedToken(ctx context.Context, vaultID int64, tokenAddress string) (txHash string, err error)
	// RemoveSupportedToken removes a token address from the vault's whitelist on the smart contract.
	// Parameters:
	//   - ctx: The context for the request.
	//   - vaultID: The ID of the vault to modify.
	//   - tokenAddress: The address of the token to remove from the whitelist.
	// Returns:
	//   - txHash: The transaction hash of the blockchain operation.
	//   - err: An error if the vault is not found, not active, the address is invalid, or contract execution fails.
	RemoveSupportedToken(ctx context.Context, vaultID int64, tokenAddress string) (txHash string, err error)
	// StartRecovery initiates the recovery process for a vault on the smart contract.
	// Parameters:
	//   - ctx: The context for the request.
	//   - vaultID: The ID of the vault to start recovery for.
	// Returns:
	//   - txHash: The transaction hash of the blockchain operation.
	//   - err: An error if the vault is not found, the state transition is invalid, or contract execution fails.
	StartRecovery(ctx context.Context, vaultID int64) (txHash string, err error)
	// CancelRecovery cancels an ongoing recovery process for a vault on the smart contract.
	// This can only be done before the recovery timelock expires.
	// Parameters:
	//   - ctx: The context for the request.
	//   - vaultID: The ID of the vault whose recovery process should be cancelled.
	// Returns:
	//   - txHash: The transaction hash of the blockchain operation.
	//   - err: An error if the vault is not found, not in recovery, the timelock expired, or contract execution fails.
	CancelRecovery(ctx context.Context, vaultID int64) (txHash string, err error)
	// ExecuteRecovery finalizes the recovery process after the timelock has expired,
	// transferring control to the recovery address.
	// Parameters:
	//   - ctx: The context for the request.
	//   - vaultID: The ID of the vault to finalize recovery for.
	// Returns:
	//   - txHash: The transaction hash of the blockchain operation.
	//   - err: An error if the vault is not found, not in recovery, timelock not expired, or contract execution fails.
	ExecuteRecovery(ctx context.Context, vaultID int64) (txHash string, err error)
}

// service implements the VaultService interface.
type service struct {
	repo            Repository
	contractFactory contract.Factory
	walletService   wallet.Service
	walletFactory   coreWallet.Factory
	log             logger.Logger
	cfg             *config.Config

	recoveryPollingCtx         context.Context
	recoveryPollingCancel      context.CancelFunc
	recoveryInterval           time.Duration
	deploymentMonitoringCtx    context.Context
	deploymentMonitoringCancel context.CancelFunc
	deploymentInterval         time.Duration
}

// NewService creates a new vault service instance.
func NewService(
	repo Repository,
	contractFactory contract.Factory,
	walletService wallet.Service,
	walletFactory coreWallet.Factory,
	log logger.Logger,
	cfg *config.Config,
) Service {
	depInterval := defaultDeploymentInterval
	if cfg != nil && cfg.Vault.DeploymentUpdateInterval > 0 {
		depInterval = time.Duration(cfg.Vault.DeploymentUpdateInterval) * time.Second
	}

	recInterval := defaultRecoveryInterval
	if cfg != nil && cfg.Vault.RecoveryUpdateInterval > 0 {
		recInterval = time.Duration(cfg.Vault.RecoveryUpdateInterval) * time.Second
	}

	return &service{
		repo:               repo,
		contractFactory:    contractFactory,
		walletService:      walletService,
		walletFactory:      walletFactory,
		log:                log,
		cfg:                cfg,
		deploymentInterval: depInterval,
		recoveryInterval:   recInterval,
	}
}

// CreateVault handles the creation of a new vault, including deploying the smart contract.
func (s *service) CreateVault(
	ctx context.Context,
	walletID int64,
	name string,
	recoveryAddress string,
	signers []string,
	quorum int,
	whitelistedTokens []string,
) (*Vault, error) {

	walletInfo, err := s.walletService.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	if err := s.validateCreateVaultParams(walletInfo.ChainType, name, recoveryAddress, signers, quorum); err != nil {
		return nil, err
	}

	wallet, err := s.walletFactory.NewWallet(ctx, walletInfo.ChainType, walletInfo.Address)
	if err != nil {
		return nil, err
	}

	contractCore, err := s.contractFactory.NewSmartContract(ctx, wallet)
	if err != nil {
		return nil, err
	}

	// Load contract artifact
	artifact, err := contractCore.LoadArtifact(ctx, types.MultiSigContractName)
	if err != nil {
		return nil, err
	}

	// Prepare constructor arguments
	// Make sure the order matches the contract's constructor in @MultiSigWallet.sol
	constructorArgs := []any{
		signers,
		uint64(quorum),
		recoveryAddress,
		whitelistedTokens,
	}

	deployOpts := contract.DeploymentOptions{
		ConstructorArgs: constructorArgs,
	}

	deployResult, err := contractCore.Deploy(ctx, artifact, deployOpts)
	if err != nil {
		return nil, err
	}

	txHash := deployResult.TransactionHash
	s.log.Info("MultiSigWallet contract deployment submitted", logger.String("tx_hash", txHash))

	targetStatus := VaultStatusDeploying
	now := time.Now()

	// Assuming Vault struct has these fields with correct types
	vault := &Vault{
		ID:              0,
		Name:            name,
		ContractName:    types.MultiSigContractName,
		WalletID:        walletID,
		RecoveryAddress: recoveryAddress,
		Signers:         types.JSONArray(signers),
		Quorum:          quorum,
		Status:          targetStatus,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Save vault record to database
	if err := s.repo.Create(ctx, vault); err != nil {
		s.log.Error("CRITICAL: Contract deployment submitted but database creation failed",
			logger.String("tx_hash", txHash),
			logger.Error(err))
		// This is a critical state. The transaction is out, but we failed to record it.
		// Ideally, implement a fallback mechanism (e.g., write to alert queue, retry later).
		// For now, return a specific DB error.
		return nil, err
	}

	s.log.Info("Vault record created, awaiting deployment confirmation", logger.Int64("vault_id", vault.ID))
	return vault, nil
}

// Helper validation function (adapt as needed)
func (s *service) validateCreateVaultParams(chainType types.ChainType, name string, recoveryAddress string,
	signers []string, quorum int) error {

	if name == "" {
		return errors.NewInvalidParameterError("name", "missing")
	}

	_, err := types.NewAddress(chainType, recoveryAddress)
	if err != nil {
		return errors.NewInvalidParameterError("recovery_address", err.Error())
	}

	// Validate signers based on contract constraints (assuming Min/Max constants exist or use config)
	const MinSignersPerVault = 2
	const MaxSignersPerVault = 7
	if len(signers) < MinSignersPerVault || len(signers) > MaxSignersPerVault {
		return errors.NewInvalidParameterError("signers", fmt.Sprintf("must have between %d and %d signers, got %d", MinSignersPerVault, MaxSignersPerVault, len(signers)))
	}

	for _, signer := range signers {
		_, err := types.NewAddress(chainType, signer)
		if err != nil {
			return errors.NewInvalidParameterError("signers", fmt.Sprintf("invalid signer address format: %s", signer))
		}
	}

	// Validate threshold (quorum in contract terms >= (N+1)/2 and >= 2)
	minQuorum := max((len(signers)+1)/2, 2)

	if quorum < minQuorum || quorum > len(signers) {
		return errors.NewInvalidParameterError("signature_threshold", fmt.Sprintf("must be between %d and %d (inclusive) for %d signers", minQuorum, len(signers), len(signers)))
	}

	return nil
}

// ExecuteRecovery provides an optional manual trigger for recovery.
// The main logic is handled by the polling job.
func (s *service) ExecuteRecovery(ctx context.Context, vaultID int64) (string, error) {
	vault, err := s.repo.GetByID(ctx, vaultID)
	if err != nil {
		if errors.IsError(err, errors.ErrCodeNotFound) {
			return "", errors.NewVaultNotFoundError(vaultID)
		}
		s.log.Error("Failed to get vault for execute recovery", logger.Int64("vault_id", vaultID), logger.Error(err))
		return "", err
	}

	currentStatus := VaultStatus(vault.Status)

	// Verify vault is in recovering state
	if currentStatus != VaultStatusRecovering {
		return "", errors.NewInvalidStateTransitionError(string(currentStatus), string(VaultStatusRecovered))
	}

	// Check timelock
	if vault.RecoveryRequestTimestamp == nil {
		return "", errors.NewOperationFailedError("execute_recovery", fmt.Errorf("recovery timestamp missing"))
	}

	if time.Since(*vault.RecoveryRequestTimestamp) < defaultRecoveryDelay {
		return "", errors.NewOperationFailedError("execute_recovery", fmt.Errorf("recovery timelock has not expired yet (required: %s, elapsed: %s)",
			defaultRecoveryDelay, time.Since(*vault.RecoveryRequestTimestamp)))
	}

	// Fetch wallet to get the address for signing
	walletInfo, err := s.walletService.GetByID(ctx, vault.WalletID)
	if err != nil {
		s.log.Error("Failed to get signing wallet for ExecuteRecovery", logger.Int64("vault_id", vaultID), logger.Error(err))
		return "", errors.NewOperationFailedError("get_signing_wallet", err)
	}

	walletCore, err := s.walletFactory.NewWallet(ctx, walletInfo.ChainType, walletInfo.Address)
	if err != nil {
		s.log.Error("Failed to create core wallet instance for execute recovery",
			logger.Int64("vault_id", vault.ID),
			logger.Error(err))
		return "", err
	}

	contractCore, err := s.contractFactory.NewSmartContract(ctx, walletCore)
	if err != nil {
		s.log.Error("Failed to create core contract instance for execute recovery",
			logger.Int64("vault_id", vault.ID),
			logger.Error(err))
		return "", err
	}

	artifact, err := contractCore.LoadArtifact(ctx, vault.ContractName)
	if err != nil {
		s.log.Error("Failed to load contract artifact for execute recovery",
			logger.Int64("vault_id", vault.ID),
			logger.String("contract_name", vault.ContractName),
			logger.Error(err))
		return "", err
	}

	s.log.Info("Calling executeRecovery on contract",
		logger.Int64("vault_id", vaultID),
		logger.String("contract_address", vault.Address))

	// Execute method using the contractCore instance
	opts := contract.ExecuteOptions{Value: big.NewInt(0)}
	txHash, err := contractCore.ExecuteMethod(
		ctx,
		vault.Address,
		string(types.MultiSigExecuteRecoveryMethod),
		artifact.ABI,
		opts)

	if err != nil {
		s.log.Error("Failed to execute executeRecovery on contract", logger.Error(err))
		return "", errors.NewOperationFailedError("execute_recovery", fmt.Errorf("contract execution failed: %w", err))
	}

	// Update vault status in DB
	targetStatus := VaultStatusRecovered
	now := time.Now()
	vault.Status = targetStatus
	// Keep RecoveryRequestTimestamp for record
	vault.UpdatedAt = now

	if err := s.repo.Update(ctx, vault.ID, vault); err != nil {
		s.log.Error("CRITICAL: Recovery executed on contract but database update failed",
			logger.Int64("vault_id", vaultID),
			logger.String("tx_hash", txHash),
			logger.Error(err))
		return txHash, errors.NewDatabaseError(err)
	}

	s.log.Info("Vault status updated to recovered", logger.Int64("vault_id", vaultID))
	return txHash, nil
}

func (s *service) GetVaultByID(ctx context.Context, vaultID int64) (*Vault, error) {
	vault, err := s.repo.GetByID(ctx, vaultID)
	if err != nil {
		if errors.IsError(err, errors.ErrCodeNotFound) {
			return nil, errors.NewVaultNotFoundError(vaultID)
		}
		s.log.Error("Failed to get vault by ID from repository", logger.Int64("vault_id", vaultID), logger.Error(err))
		return nil, err
	}
	return vault, nil
}

func (s *service) ListVaults(ctx context.Context, filter VaultFilter, limit int, nextToken string) (*types.Page[*Vault], error) {
	page, err := s.repo.List(ctx, filter, limit, nextToken)
	if err != nil {
		s.log.Error("Failed to list vaults from repository", logger.Error(err))
		return nil, err
	}
	return page, nil
}

func (s *service) UpdateVaultName(ctx context.Context, vaultID int64, newName string) (*Vault, error) {
	if newName == "" {
		return nil, errors.NewMissingParameterError("newName")
	}

	vault, err := s.repo.GetByID(ctx, vaultID)
	if err != nil {
		if errors.IsError(err, errors.ErrCodeNotFound) {
			return nil, errors.NewVaultNotFoundError(vaultID)
		}
		s.log.Error("Failed to get vault for name update", logger.Int64("vault_id", vaultID), logger.Error(err))
		return nil, err
	}

	if VaultStatus(vault.Status) == VaultStatusFailed {
		return nil, errors.NewOperationFailedError("update_name", fmt.Errorf("cannot update name of a failed vault"))
	}

	vault.Name = newName

	if err := s.repo.Update(ctx, vault.ID, vault); err != nil {
		s.log.Error("Failed to update vault name in repository", logger.Int64("vault_id", vaultID), logger.Error(err))
		return nil, err
	}

	s.log.Info("Vault name updated", logger.Int64("vault_id", vaultID), logger.String("new_name", newName))
	return vault, nil
}

func (s *service) AddSupportedToken(ctx context.Context, vaultID int64, tokenAddress string) (string, error) {
	vault, err := s.repo.GetByID(ctx, vaultID)
	if err != nil {
		return "", err
	}

	walletInfo, err := s.walletService.GetByID(ctx, vault.WalletID)
	if err != nil {
		return "", err
	}

	validatedAddr, err := types.NewAddress(walletInfo.ChainType, tokenAddress)
	if err != nil {
		return "", err
	}
	normalizedTokenAddr := validatedAddr.String()

	if VaultStatus(vault.Status) != VaultStatusActive {
		return "", errors.NewOperationFailedError("add_token", fmt.Errorf("vault must be active (current status: %s)", vault.Status))
	}

	contractInfo, err := s.repo.GetByID(ctx, vaultID)
	if err != nil {
		if errors.IsError(err, errors.ErrCodeNotFound) {
			return "", errors.NewVaultNotFoundError(vaultID)
		}
		s.log.Error("Failed to get contract info for token add",
			logger.Int64("vault_id", vaultID),
			logger.Error(err))
		return "", err
	}
	if contractInfo == nil || contractInfo.Address == "" {
		return "", errors.NewOperationFailedError("add_token", fmt.Errorf("contract address is missing for vault %d", vaultID))
	}
	vaultAddress := contractInfo.Address

	walletCore, err := s.walletFactory.NewWallet(ctx, walletInfo.ChainType, walletInfo.Address)
	if err != nil {
		s.log.Error("Failed to create core wallet instance for token add",
			logger.Int64("vault_id", vaultID),
			logger.Error(err))
		return "", err
	}

	contractCore, err := s.contractFactory.NewSmartContract(ctx, walletCore)
	if err != nil {
		s.log.Error("Failed to create core contract instance for token add",
			logger.Int64("vault_id", vaultID),
			logger.Error(err))
		return "", err
	}
	contractArtifact, err := contractCore.LoadArtifact(ctx, contractInfo.ContractName)
	if err != nil {
		s.log.Error("Failed to load contract artifact for token add",
			logger.Int64("vault_id", vaultID),
			logger.String("contract_name", contractInfo.ContractName),
			logger.Error(err))
		return "", err
	}
	args := []any{normalizedTokenAddr}

	s.log.Info("Calling addSupportedToken on contract",
		logger.Int64("vault_id", vaultID),
		logger.String("contract_address", vaultAddress),
		logger.String("token_address", normalizedTokenAddr))

	// Execute method using the contractCore instance
	opts := contract.ExecuteOptions{Value: big.NewInt(0)}
	txHash, err := contractCore.ExecuteMethod(
		ctx,
		contractInfo.Address,
		string(types.MultiSigAddSupportedTokenMethod),
		contractArtifact.ABI,
		opts,
		args...)

	if err != nil {
		s.log.Error("Failed to execute addSupportedToken on contract", logger.Error(err))
		return "", errors.NewOperationFailedError("execute_add_token", fmt.Errorf("contract execution failed: %w", err))
	}

	s.log.Info("addSupportedToken transaction submitted", logger.String("tx_hash", txHash))

	return txHash, nil
}

func (s *service) RemoveSupportedToken(ctx context.Context, vaultID int64, tokenAddress string) (string, error) {
	vault, err := s.repo.GetByID(ctx, vaultID)
	if err != nil {
		return "", err
	}

	walletInfo, err := s.walletService.GetByID(ctx, vault.WalletID)
	if err != nil {
		return "", err
	}

	validatedAddr, err := types.NewAddress(walletInfo.ChainType, tokenAddress)
	if err != nil {
		return "", errors.NewInvalidParameterError("tokenAddress", err.Error())
	}
	normalizedTokenAddr := validatedAddr.String()

	if VaultStatus(vault.Status) != VaultStatusActive {
		return "", errors.NewOperationFailedError("remove_token", fmt.Errorf("vault must be active to remove tokens (current status: %s)", vault.Status))
	}

	vaultContractAddress := vault.Address
	if vaultContractAddress == "" {
		return "", errors.NewOperationFailedError("remove_token", fmt.Errorf("contract address is missing for vault %d", vaultID))
	}

	walletCore, err := s.walletFactory.NewWallet(ctx, walletInfo.ChainType, walletInfo.Address)
	if err != nil {
		s.log.Error("Failed to create core wallet instance for token remove", logger.Int64("vault_id", vaultID), logger.Error(err))
		return "", err
	}

	contractCore, err := s.contractFactory.NewSmartContract(ctx, walletCore)
	if err != nil {
		s.log.Error("Failed to create core contract instance for token remove", logger.Int64("vault_id", vaultID), logger.Error(err))
		return "", err
	}

	artifact, err := contractCore.LoadArtifact(ctx, vault.ContractName)
	if err != nil {
		s.log.Error("Failed to load contract artifact for token remove",
			logger.Int64("vault_id", vaultID),
			logger.String("contract_name", vault.ContractName),
			logger.Error(err))
		return "", fmt.Errorf("failed to load artifact %s: %w", vault.ContractName, err)
	}

	args := []any{normalizedTokenAddr}

	s.log.Info("Calling removeSupportedToken on contract",
		logger.Int64("vault_id", vaultID),
		logger.String("contract_address", vaultContractAddress),
		logger.String("token_address", normalizedTokenAddr))

	opts := contract.ExecuteOptions{Value: big.NewInt(0)}

	txHash, err := contractCore.ExecuteMethod(
		ctx,
		vaultContractAddress,
		string(types.MultiSigRemoveSupportedTokenMethod),
		artifact.ABI,
		opts,
		args...)

	if err != nil {
		s.log.Error("Failed to execute removeSupportedToken on contract", logger.Error(err))
		return "", errors.NewOperationFailedError("execute_remove_token", fmt.Errorf("contract execution failed: %w", err))
	}

	s.log.Info("removeSupportedToken transaction submitted", logger.String("tx_hash", txHash))

	return txHash, nil
}

func (s *service) StartRecovery(ctx context.Context, vaultID int64) (string, error) {
	vault, err := s.repo.GetByID(ctx, vaultID)
	if err != nil {
		if errors.IsError(err, errors.ErrCodeNotFound) {
			return "", errors.NewVaultNotFoundError(vaultID)
		}
		s.log.Error("Failed to get vault for start recovery", logger.Int64("vault_id", vaultID), logger.Error(err))
		return "", err
	}

	currentStatus := VaultStatus(vault.Status)
	targetStatus := VaultStatusRecovering

	if !CanTransition(currentStatus, targetStatus) {
		return "", errors.NewInvalidStateTransitionError(string(currentStatus), string(targetStatus))
	}

	// Fetch wallet to get the address for signing
	walletInfo, err := s.walletService.GetByID(ctx, vault.WalletID)
	if err != nil {
		s.log.Error("Failed to get signing wallet for StartRecovery", logger.Int64("vault_id", vaultID), logger.Error(err))
		return "", errors.NewOperationFailedError("get_signing_wallet", err)
	}

	walletCore, err := s.walletFactory.NewWallet(ctx, walletInfo.ChainType, walletInfo.Address)
	if err != nil {
		s.log.Error("Failed to create core wallet instance for recovery",
			logger.Int64("vault_id", vaultID),
			logger.Error(err))
		return "", err
	}

	contractCore, err := s.contractFactory.NewSmartContract(ctx, walletCore)
	if err != nil {
		s.log.Error("Failed to create core contract instance for recovery",
			logger.Int64("vault_id", vaultID),
			logger.Error(err))
		return "", err
	}

	artifact, err := contractCore.LoadArtifact(ctx, vault.ContractName)
	if err != nil {
		s.log.Error("Failed to load contract artifact for recovery",
			logger.Int64("vault_id", vaultID),
			logger.String("contract_name", vault.ContractName),
			logger.Error(err))
		return "", err
	}

	s.log.Info("Calling requestRecovery on contract",
		logger.Int64("vault_id", vaultID),
		logger.String("contract_address", vault.Address))

	// Execute method using the contractCore instance
	opts := contract.ExecuteOptions{Value: big.NewInt(0)}
	txHash, err := contractCore.ExecuteMethod(
		ctx,
		vault.Address,
		string(types.MultiSigRequestRecoveryMethod),
		artifact.ABI,
		opts)

	if err != nil {
		s.log.Error("Failed to execute requestRecovery on contract", logger.Error(err))
		return "", errors.NewOperationFailedError("execute_request_recovery", fmt.Errorf("contract execution failed: %w", err))
	}

	// Update vault status and timestamp in DB
	now := time.Now()
	vault.Status = targetStatus
	vault.RecoveryRequestTimestamp = &now
	vault.UpdatedAt = now

	if err := s.repo.Update(ctx, vault.ID, vault); err != nil {
		s.log.Error("CRITICAL: Recovery requested on contract but database update failed",
			logger.Int64("vault_id", vaultID),
			logger.String("tx_hash", txHash),
			logger.Error(err))
		// Return DB error but also the txHash so it can potentially be tracked
		return txHash, errors.NewDatabaseError(err)
	}

	s.log.Info("Vault status updated to recovering", logger.Int64("vault_id", vaultID))
	return txHash, nil
}

func (s *service) CancelRecovery(ctx context.Context, vaultID int64) (string, error) {
	vault, err := s.repo.GetByID(ctx, vaultID)
	if err != nil {
		if errors.IsError(err, errors.ErrCodeNotFound) {
			return "", errors.NewVaultNotFoundError(vaultID)
		}
		s.log.Error("Failed to get vault for cancel recovery", logger.Int64("vault_id", vaultID), logger.Error(err))
		return "", err
	}

	currentStatus := VaultStatus(vault.Status)
	targetStatus := VaultStatusActive

	if !CanTransition(currentStatus, targetStatus) {
		// It's only valid to cancel if status is Recovering
		return "", errors.NewInvalidStateTransitionError(string(currentStatus), string(targetStatus))
	}

	// Check timelock (must be BEFORE delay expires)
	if vault.RecoveryRequestTimestamp == nil {
		return "", errors.NewOperationFailedError("cancel_recovery", fmt.Errorf("recovery not requested"))
	}
	if time.Since(*vault.RecoveryRequestTimestamp) >= defaultRecoveryDelay {
		return "", errors.NewOperationFailedError("cancel_recovery", fmt.Errorf("recovery delay period has expired, cannot cancel"))
	}

	// Fetch wallet to get the address for signing
	walletInfo, err := s.walletService.GetByID(ctx, vault.WalletID)
	if err != nil {
		s.log.Error("Failed to get signing wallet for CancelRecovery", logger.Int64("vault_id", vaultID), logger.Error(err))
		return "", errors.NewOperationFailedError("get_signing_wallet", err)
	}

	walletCore, err := s.walletFactory.NewWallet(ctx, walletInfo.ChainType, walletInfo.Address)
	if err != nil {
		s.log.Error("Failed to create core wallet instance for cancel recovery",
			logger.Int64("vault_id", vaultID),
			logger.Error(err))
		return "", err
	}

	contractCore, err := s.contractFactory.NewSmartContract(ctx, walletCore)
	if err != nil {
		s.log.Error("Failed to create core contract instance for cancel recovery",
			logger.Int64("vault_id", vaultID),
			logger.Error(err))
		return "", err
	}

	artifact, err := contractCore.LoadArtifact(ctx, vault.ContractName)
	if err != nil {
		s.log.Error("Failed to load contract artifact for cancel recovery",
			logger.Int64("vault_id", vaultID),
			logger.String("contract_name", vault.ContractName),
			logger.Error(err))
		return "", err
	}

	s.log.Info("Calling cancelRecovery on contract",
		logger.Int64("vault_id", vaultID),
		logger.String("contract_address", vault.Address))

	// Execute method using the contractCore instance
	opts := contract.ExecuteOptions{Value: big.NewInt(0)}
	txHash, err := contractCore.ExecuteMethod(
		ctx,
		vault.Address,
		string(types.MultiSigCancelRecoveryMethod),
		artifact.ABI,
		opts)

	if err != nil {
		s.log.Error("Failed to execute cancelRecovery on contract", logger.Error(err))
		return "", errors.NewOperationFailedError("execute_cancel_recovery", fmt.Errorf("contract execution failed: %w", err))
	}

	// Update vault status and clear timestamp in DB
	now := time.Now()
	vault.Status = targetStatus
	vault.RecoveryRequestTimestamp = nil
	vault.UpdatedAt = now

	if err := s.repo.Update(ctx, vault.ID, vault); err != nil {
		s.log.Error("CRITICAL: Recovery cancelled on contract but database update failed",
			logger.Int64("vault_id", vaultID),
			logger.String("tx_hash", txHash),
			logger.Error(err))
		return txHash, errors.NewDatabaseError(err)
	}

	s.log.Info("Vault status updated to active after recovery cancellation", logger.Int64("vault_id", vaultID))
	return txHash, nil
}
