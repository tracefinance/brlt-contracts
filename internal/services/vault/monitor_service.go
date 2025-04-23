package vault

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"vault0/internal/core/contract"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

type MonitorService interface {
	// ProcessVaultDeploymentSuccess updates the vault status to Active upon successful contract deployment.
	// Parameters:
	//   - ctx: The context for the request.
	//   - vaultID: The ID of the vault whose deployment succeeded.
	//   - contractAddress: The address of the newly deployed smart contract.
	//   - txHash: The transaction hash of the deployment.
	// Returns:
	//   - error: An error if the vault is not found or the database update fails.
	ProcessVaultDeploymentSuccess(ctx context.Context, vaultID int64, contractAddress, txHash string) error
	// ProcessVaultDeploymentFailure updates the vault status to Failed upon contract deployment failure.
	// Parameters:
	//   - ctx: The context for the request.
	//   - vaultID: The ID of the vault whose deployment failed.
	//   - errorMsg: A description of the deployment failure reason.
	// Returns:
	//   - error: An error if the vault is not found or the database update fails.
	ProcessVaultDeploymentFailure(ctx context.Context, vaultID int64, errorMsg string) error
	// StartRecoveryPolling initiates a background goroutine that periodically checks for vaults
	// in the 'Recovering' state whose timelock has expired and attempts to execute recovery.
	// Parameters:
	//   - ctx: The parent context for the polling goroutine.
	StartRecoveryPolling(ctx context.Context)
	// StopRecoveryPolling signals the background recovery polling goroutine to stop.
	StopRecoveryPolling()
	// StartDeploymentMonitoring initiates a background goroutine that periodically checks the status
	// of vaults in the 'Deploying' state by querying their deployment transaction hash.
	// Parameters:
	//   - ctx: The parent context for the monitoring goroutine.
	StartDeploymentMonitoring(ctx context.Context)
	// StopDeploymentMonitoring signals the background deployment monitoring goroutine to stop.
	StopDeploymentMonitoring()
}

// ProcessVaultDeploymentSuccess is called when deployment is confirmed.
func (s *service) ProcessVaultDeploymentSuccess(ctx context.Context, vaultID int64, contractAddress, txHash string) error {
	vault, err := s.repo.GetByID(ctx, vaultID)
	if err != nil {
		s.log.Error("Failed to get vault for deployment success update", logger.Int64("vault_id", vaultID), logger.Error(err))
		if errors.IsError(err, errors.ErrCodeNotFound) {
			return errors.NewVaultNotFoundError(vaultID)
		}
		return err
	}

	currentStatus := VaultStatus(vault.Status)
	targetStatus := VaultStatusActive

	if !CanTransition(currentStatus, targetStatus) {
		s.log.Warn("Cannot transition vault to Active after deployment success",
			logger.Int64("vault_id", vaultID),
			logger.String("current_status", string(currentStatus)),
			logger.String("target_status", string(targetStatus)))
		if currentStatus == VaultStatusActive {
			return nil
		}
		return errors.NewInvalidStateTransitionError(string(currentStatus), string(targetStatus))
	}

	vault.Status = targetStatus
	vault.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, vault.ID, vault); err != nil {
		s.log.Error("Failed to update vault status to Active after deployment success",
			logger.Int64("vault_id", vaultID),
			logger.String("contract_address", contractAddress),
			logger.Error(err))
		return err
	}

	s.log.Info("Vault successfully activated", logger.Int64("vault_id", vaultID), logger.String("contract_address", contractAddress))
	return nil
}

// ProcessVaultDeploymentFailure is called when deployment fails.
func (s *service) ProcessVaultDeploymentFailure(ctx context.Context, vaultID int64, errorMsg string) error {
	vault, err := s.repo.GetByID(ctx, vaultID)
	if err != nil {
		s.log.Error("Failed to get vault for deployment failure update", logger.Int64("vault_id", vaultID), logger.Error(err))
		if errors.IsError(err, errors.ErrCodeNotFound) {
			return errors.NewVaultNotFoundError(vaultID)
		}
		return err
	}

	currentStatus := VaultStatus(vault.Status)
	targetStatus := VaultStatusFailed

	if !CanTransition(currentStatus, targetStatus) {
		s.log.Warn("Cannot transition vault to Failed after deployment failure",
			logger.Int64("vault_id", vaultID),
			logger.String("current_status", string(currentStatus)),
			logger.String("target_status", string(targetStatus)))
		if currentStatus == VaultStatusFailed || currentStatus == VaultStatusActive {
			return nil
		}
		return errors.NewInvalidStateTransitionError(string(currentStatus), string(targetStatus))
	}

	vault.Status = targetStatus
	vault.FailureReason = &errorMsg

	if err := s.repo.Update(ctx, vault.ID, vault); err != nil {
		s.log.Error("Failed to update vault status to Failed after deployment failure",
			logger.Int64("vault_id", vaultID),
			logger.String("error_msg", errorMsg),
			logger.Error(err))
		return err
	}

	s.log.Warn("Vault deployment failed", logger.Int64("vault_id", vaultID), logger.String("reason", errorMsg))
	return nil
}

// --- Deployment Monitoring Logic --
// StartDeploymentMonitoring starts the background job for checking pending deployments.
func (s *service) StartDeploymentMonitoring(ctx context.Context) {
	if s.deploymentMonitoringCancel != nil {
		s.log.Warn("Deployment monitoring already started")
		return
	}
	s.deploymentMonitoringCtx, s.deploymentMonitoringCancel = context.WithCancel(ctx)

	s.log.Info("Starting deployment monitoring scheduler", logger.Duration("interval", s.deploymentInterval))

	go func() {
		ticker := time.NewTicker(s.deploymentInterval)
		defer ticker.Stop()

		for {
			select {
			case <-s.deploymentMonitoringCtx.Done():
				s.log.Info("Deployment monitoring scheduler stopped")
				return
			case <-ticker.C:
				s.log.Debug("Running scheduled deployment check")
				updatedCount, err := s.checkPendingDeployments(s.deploymentMonitoringCtx)
				if err != nil {
					s.log.Error("Error during scheduled deployment check", logger.Error(err))
				}
				if updatedCount > 0 {
					s.log.Info("Scheduled deployment check completed", logger.Int("deployments_processed", updatedCount))
				} else {
					s.log.Debug("Scheduled deployment check completed, no actions taken")
				}
			}
		}
	}()
}

// StopDeploymentMonitoring stops the background deployment monitoring job.
func (s *service) StopDeploymentMonitoring() {
	if s.deploymentMonitoringCancel != nil {
		s.log.Info("Stopping deployment monitoring scheduler")
		s.deploymentMonitoringCancel()
		s.deploymentMonitoringCancel = nil // Mark as stopped
	} else {
		s.log.Warn("Deployment monitoring not running")
	}
}

// checkPendingDeployments fetches vaults in 'deploying' state and checks their status.
func (s *service) checkPendingDeployments(ctx context.Context) (int, error) {
	// Correct VaultFilter usage with pointer status
	status := VaultStatusDeploying
	filters := VaultFilter{Status: &status}
	// Fetch all potentially deployable vaults (no pagination needed for polling job?)
	page, err := s.repo.List(ctx, filters, 0, "")
	if err != nil {
		s.log.Error("polling: Failed to list vaults in deploying state", logger.Error(err))
		return 0, errors.NewDatabaseError(err)
	}

	if len(page.Items) == 0 {
		return 0, nil // Nothing to do
	}

	s.log.Info("polling: Found vaults in deploying state", logger.Int("count", len(page.Items)))
	processedCount := 0
	var firstError error

	for _, vault := range page.Items {
		if ctx.Err() != nil {
			return processedCount, ctx.Err()
		}
		err := s.checkDeploymentStatus(ctx, vault)
		if err != nil {
			s.log.Error("polling: Failed to check deployment status for vault", logger.Int64("vault_id", vault.ID), logger.Error(err))
			if firstError == nil {
				firstError = err
			}
			// Continue processing others
		} else {
			processedCount++ // Count attempts, success/failure handled internally
		}
	}

	return processedCount, firstError
}

// checkDeploymentStatus checks the status of a single vault's deployment transaction.
func (s *service) checkDeploymentStatus(ctx context.Context, vault *Vault) error {
	// Check if deployment transaction hash exists
	if vault.TxHash == "" {
		s.log.Error("polling: Vault in deploying state has no deployment transaction hash",
			logger.Int64("vault_id", vault.ID))
		return s.ProcessVaultDeploymentFailure(ctx, vault.ID, "Missing deployment transaction hash")
	}

	// Fetch wallet to get the address for signing
	associatedWallet, err := s.walletService.GetByID(ctx, vault.WalletID)
	if err != nil {
		s.log.Error("polling: Failed to get wallet for deployment check", logger.Int64("vault_id", vault.ID), logger.Error(err))
		return err
	}

	walletCore, err := s.walletFactory.NewWallet(ctx, associatedWallet.ChainType, associatedWallet.Address)
	if err != nil {
		s.log.Error("polling: Failed to create core wallet for deployment check", logger.Int64("vault_id", vault.ID), logger.Error(err))
		return err
	}

	contractCore, err := s.contractFactory.NewSmartContract(ctx, walletCore)
	if err != nil {
		s.log.Error("polling: Failed to create core contract for deployment check", logger.Int64("vault_id", vault.ID), logger.Error(err))
		return err
	}

	// Use GetDeployment to check if the contract deployment is complete
	s.log.Info("polling: Checking deployment status",
		logger.Int64("vault_id", vault.ID),
		logger.String("tx_hash", vault.TxHash))

	deploymentResult, err := contractCore.GetDeployment(ctx, vault.TxHash)
	if err != nil {
		// Could be a temporary network issue, or transaction failed
		s.log.Warn("polling: Error getting deployment status, will retry",
			logger.Int64("vault_id", vault.ID),
			logger.String("tx_hash", vault.TxHash),
			logger.Error(err))

		// If it's a transaction failure that's final (not just pending), mark vault as failed
		if errors.IsError(err, errors.ErrCodeTransactionFailed) {
			errorMsg := "Contract deployment transaction failed on chain"
			s.log.Error("polling: Deployment failed permanently",
				logger.Int64("vault_id", vault.ID),
				logger.String("tx_hash", vault.TxHash),
				logger.Error(err))
			return s.ProcessVaultDeploymentFailure(ctx, vault.ID, errorMsg)
		}

		// For other errors, like network issues, just return nil to retry later
		return nil
	}

	if deploymentResult == nil {
		// No result but also no error likely means transaction is still pending
		s.log.Info("polling: Contract deployment still in progress", logger.Int64("vault_id", vault.ID))
		return nil
	}

	// Successfully got deployment result with contract address
	if deploymentResult.ContractAddress != "" {
		s.log.Info("polling: Contract deployment successful",
			logger.Int64("vault_id", vault.ID),
			logger.String("contract_address", deploymentResult.ContractAddress),
			logger.Int64("block_number", int64(deploymentResult.BlockNumber)))

		return s.ProcessVaultDeploymentSuccess(ctx, vault.ID, deploymentResult.ContractAddress, vault.TxHash)
	}

	// Got a result but no contract address, indicates failure
	s.log.Error("polling: Contract deployment failed - no contract address in result",
		logger.Int64("vault_id", vault.ID),
		logger.String("tx_hash", vault.TxHash))

	return s.ProcessVaultDeploymentFailure(ctx, vault.ID, "Deployment transaction did not produce a contract address")
}

// --- Recovery Polling Logic ---

// StartRecoveryPolling starts the background job for checking eligible recoveries.
func (s *service) StartRecoveryPolling(ctx context.Context) {
	if s.recoveryPollingCancel != nil {
		s.log.Warn("Recovery polling already started")
		return
	}
	s.recoveryPollingCtx, s.recoveryPollingCancel = context.WithCancel(ctx)

	s.log.Info("Starting recovery polling scheduler", logger.Duration("interval", s.recoveryInterval))

	go func() {
		ticker := time.NewTicker(s.recoveryInterval)
		defer ticker.Stop()

		for {
			select {
			case <-s.recoveryPollingCtx.Done():
				s.log.Info("Recovery polling scheduler stopped")
				return
			case <-ticker.C:
				s.log.Debug("Running scheduled recovery check")
				updatedCount, err := s.checkAndExecuteRecoveries(s.recoveryPollingCtx) // Use the cancellable context
				if err != nil {
					s.log.Error("Error during scheduled recovery check", logger.Error(err))
				}
				if updatedCount > 0 {
					s.log.Info("Scheduled recovery check completed", logger.Int("recoveries_executed", updatedCount))
				} else {
					s.log.Debug("Scheduled recovery check completed, no actions taken")
				}
			}
		}
	}()
}

// StopRecoveryPolling stops the background recovery polling job.
func (s *service) StopRecoveryPolling() {
	if s.recoveryPollingCancel != nil {
		s.log.Info("Stopping recovery polling scheduler")
		s.recoveryPollingCancel()
		s.recoveryPollingCancel = nil // Mark as stopped
	} else {
		s.log.Warn("Recovery polling not running")
	}
}

// checkAndExecuteRecoveries fetches vaults in 'recovering' state and attempts execution.
func (s *service) checkAndExecuteRecoveries(ctx context.Context) (int, error) {
	// Correct VaultFilter usage with pointer status
	status := VaultStatusRecovering
	filters := VaultFilter{Status: &status}
	// Fetch all potentially recoverable vaults (no pagination needed for polling job?)
	page, err := s.repo.List(ctx, filters, 0, "")
	if err != nil {
		s.log.Error("Polling: Failed to list vaults in recovering state", logger.Error(err))
		return 0, errors.NewDatabaseError(err)
	}

	if len(page.Items) == 0 {
		return 0, nil // Nothing to do
	}

	s.log.Info("Polling: Found vaults in recovering state", logger.Int("count", len(page.Items)))
	updatedCount := 0
	var firstError error

	for _, vault := range page.Items {
		if ctx.Err() != nil {
			return updatedCount, ctx.Err() // Context cancelled during processing
		}
		_, err := s.checkAndExecuteRecoveryForVault(ctx, vault)
		if err != nil {
			s.log.Error("Polling: Failed to check/execute recovery for vault", logger.Int64("vault_id", vault.ID), logger.Error(err))
			if firstError == nil {
				firstError = err // Report the first error encountered
			}
			// Continue processing other vaults even if one fails
		} else {
			updatedCount++ // Count successful executions initiated
		}
	}

	return updatedCount, firstError
}

// checkAndExecuteRecoveryForVault checks timelock and executes recovery for a single vault.
func (s *service) checkAndExecuteRecoveryForVault(ctx context.Context, vault *Vault) (string, error) {
	// Double check status
	if VaultStatus(vault.Status) != VaultStatusRecovering {
		// Should not happen if query is correct, but good practice
		return "", nil // Not in recovery state, nothing to do
	}

	// Check timelock
	if vault.RecoveryRequestTimestamp == nil {
		s.log.Warn("Polling: Vault in recovering state but has no recovery timestamp", logger.Int64("vault_id", vault.ID))
		return "", nil // Invalid state, skip
	}
	if time.Since(*vault.RecoveryRequestTimestamp) < defaultRecoveryDelay {
		// Timelock not yet passed
		return "", nil
	}

	s.log.Info("Polling: Timelock passed, attempting to execute recovery", logger.Int64("vault_id", vault.ID))

	// Fetch wallet to get the address for signing
	walletInfo, err := s.walletService.GetByID(ctx, vault.WalletID)
	if err != nil {
		// Use fmt.Errorf with %w for wrapping only if needed internally, otherwise return specific error
		err = fmt.Errorf("failed to get wallet address for signing recovery: %w", err)
		s.log.Error("Polling: "+err.Error(), logger.Int64("vault_id", vault.ID))
		return "", errors.NewOperationFailedError("get_signing_wallet_for_recovery", err)
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

	opts := contract.ExecuteOptions{Value: big.NewInt(0)}
	txHash, err := contractCore.ExecuteMethod(
		ctx,
		vault.Address,
		string(types.MultiSigExecuteRecoveryMethod),
		artifact.ABI,
		opts)

	if err != nil {
		s.log.Error("Polling: Failed to execute recovery on contract", logger.Int64("vault_id", vault.ID), logger.Error(err))
		return "", errors.NewOperationFailedError(string(types.MultiSigExecuteRecoveryMethod), err)
	}

	s.log.Info("Polling: executeRecovery transaction submitted", logger.Int64("vault_id", vault.ID), logger.String("tx_hash", txHash))

	// --- Update DB AFTER successful transaction submission ---
	currentStatus := VaultStatusRecovering
	targetStatus := VaultStatusRecovered // Assuming recovered is the target state

	if !CanTransition(currentStatus, targetStatus) {
		// This should ideally not happen if logic is correct
		err := errors.NewInvalidStateTransitionError(string(currentStatus), string(targetStatus))
		s.log.Error("CRITICAL: Invalid state transition after submitting recovery execution", logger.Int64("vault_id", vault.ID), logger.Error(err))
		return txHash, err
	}

	now := time.Now()
	vault.Status = targetStatus
	vault.UpdatedAt = now

	if err := s.repo.Update(ctx, vault.ID, vault); err != nil {
		s.log.Error("CRITICAL: Recovery executed on contract but database update failed",
			logger.Int64("vault_id", vault.ID),
			logger.String("tx_hash", txHash),
			logger.Error(err))
		return txHash, errors.NewDatabaseError(err)
	}

	s.log.Info("Polling: Vault status updated to recovered", logger.Int64("vault_id", vault.ID))
	return txHash, nil
}
