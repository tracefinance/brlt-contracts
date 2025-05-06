package wallet

import (
	"context"
	"fmt"
	"math/big"

	"vault0/internal/core/tokenstore"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

type BalanceService interface {
	// UpdateWalletBalance updates the native balance for a wallet based on a transaction.
	// It determines if the transaction involves the wallet as sender or receiver
	// and updates the native balance accordingly, including gas costs for senders.
	// Parameters:
	//   - ctx: Context for the operation
	//   - tx: The transaction details
	// Returns:
	//   - error: ErrWalletNotFound if wallet doesn't exist, or other processing errors
	UpdateWalletBalance(ctx context.Context, tx *types.Transaction) error

	// UpdateTokenBalance updates the token balance for a wallet based on an ERC20 transfer.
	// It identifies the involved wallet (sender or receiver) and updates the specific
	// token balance. For senders, it also deducts the native currency gas cost.
	// Parameters:
	//   - ctx: Context for the operation
	//   - transfer: The parsed ERC20 transfer details
	// Returns:
	//   - error: ErrWalletNotFound, ErrTokenNotFound, or other processing errors
	UpdateTokenBalance(ctx context.Context, transfer *types.ERC20Transfer) error

	// GetWalletBalances retrieves the native and token balances for a wallet
	GetWalletBalances(ctx context.Context, id int64) ([]*TokenBalanceData, error)

	// GetWalletBalancesByAddress retrieves the native and token balances for a wallet by its address
	GetWalletBalancesByAddress(ctx context.Context, chainType types.ChainType, address string) ([]*TokenBalanceData, error)
}

type balanceService struct {
	repository Repository
	log        logger.Logger
	tokenStore tokenstore.TokenStore
}

func NewBalanceService(
	repository Repository,
	log logger.Logger,
	tokenStore tokenstore.TokenStore,
) BalanceService {
	return &balanceService{repository, log, tokenStore}
}

// isOutgoingTransaction returns (isOutgoingTransaction, error)
// isOutgoingTransaction is true if the wallet is the sender (tx.From), false if receiver (tx.To)
func (s *balanceService) isOutgoingTransaction(ctx context.Context, tx *types.Transaction) (bool, error) {
	exists, err := s.repository.Exists(ctx, tx.ChainType, tx.From)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil // Sender
	}

	exists, err = s.repository.Exists(ctx, tx.ChainType, tx.To)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil // Receiver
	}

	return false, errors.NewWalletNotFoundError(tx.From + " or " + tx.To)
}

// UpdateWalletBalance updates the native balance for a wallet based on a transaction
func (s *balanceService) UpdateWalletBalance(ctx context.Context, tx *types.Transaction) error {
	if tx == nil {
		return errors.NewInvalidInputError("Transaction is required", "tx", nil)
	}

	if tx.Type != types.TransactionTypeNative {
		return nil
	}

	isOutgoing, err := s.isOutgoingTransaction(ctx, tx)
	if err != nil {
		return err
	}

	var wallet *Wallet
	if isOutgoing {
		wallet, err = s.repository.GetByAddress(ctx, tx.ChainType, tx.From)
	} else {
		wallet, err = s.repository.GetByAddress(ctx, tx.ChainType, tx.To)
	}
	if err != nil {
		return err
	}

	currentBalance := wallet.Balance.ToBigInt()
	var newBalance *big.Int

	if isOutgoing {
		gasUsed := new(big.Int).SetUint64(tx.GasUsed)
		totalSpent := new(big.Int).Add(tx.Value, new(big.Int).Mul(tx.GasPrice, gasUsed))
		newBalance = new(big.Int).Sub(currentBalance, totalSpent)
		if newBalance.Sign() < 0 {
			newBalance = big.NewInt(0)
		}
	} else {
		newBalance = new(big.Int).Add(currentBalance, tx.Value)
	}

	return s.repository.UpdateBalance(ctx, wallet, newBalance)
}

// UpdateTokenBalance updates the token balance for a wallet based on an ERC20 transfer
func (s *balanceService) UpdateTokenBalance(ctx context.Context, transfer *types.ERC20Transfer) error {
	if transfer == nil {
		return errors.NewInvalidInputError("ERC20 Transfer data is required", "transfer", nil)
	}

	// Determine if the transaction involves a monitored wallet as sender or receiver
	walletIsSender, senderWallet, err := s.findWalletForAddress(ctx, transfer.ChainType, transfer.From)
	if err != nil {
		appErr, ok := err.(*errors.Vault0Error)
		if !ok || appErr.Code != errors.ErrCodeNotFound {
			return fmt.Errorf("error checking sender wallet: %w", err)
		}
	}
	walletIsReceiver, receiverWallet, err := s.findWalletForAddress(ctx, transfer.ChainType, transfer.Recipient)
	if err != nil {
		appErr, ok := err.(*errors.Vault0Error)
		if !ok || appErr.Code != errors.ErrCodeNotFound {
			return fmt.Errorf("error checking receiver wallet: %w", err)
		}
	}

	if !walletIsSender && !walletIsReceiver {
		// This transfer doesn't involve any monitored wallet we manage
		s.log.Debug("ERC20 transfer does not involve a monitored wallet", logger.String("tx_hash", transfer.Hash))
		return nil
	}

	// Prefer sender if both are monitored (e.g., internal transfer), as gas cost applies to sender.
	var involvedWallet *Wallet
	isOutgoing := false
	if walletIsSender {
		involvedWallet = senderWallet
		isOutgoing = true
	} else {
		involvedWallet = receiverWallet
		isOutgoing = false // walletIsReceiver must be true here
	}

	s.log.Info("Updating token balance from ERC20 transfer",
		logger.Int64("wallet_id", involvedWallet.ID),
		logger.String("token_address", transfer.TokenAddress),
		logger.String("tx_hash", transfer.Hash),
		logger.Bool("is_outgoing", isOutgoing))

	// Normalize the token address
	tokenAddress, err := types.NewAddress(involvedWallet.ChainType, transfer.TokenAddress)
	if err != nil {
		return err
	}
	normalizedTokenAddress := tokenAddress.ToChecksum()

	// Get current balance or default to zero using the new repository method
	tb, err := s.repository.GetTokenBalance(ctx, involvedWallet.ID, normalizedTokenAddress)
	if err != nil {
		return err
	}

	// tb.Balance will be zero if the balance didn't exist in the DB
	currentTokenBalance := tb.Balance.ToBigInt()

	var newTokenBalance *big.Int
	if isOutgoing {
		newTokenBalance = new(big.Int).Sub(currentTokenBalance, transfer.Amount) // Use Amount from transfer
		if newTokenBalance.Sign() < 0 {
			newTokenBalance = big.NewInt(0)
		}
	} else {
		newTokenBalance = new(big.Int).Add(currentTokenBalance, transfer.Amount) // Use Amount from transfer
	}

	if err := s.repository.UpdateTokenBalance(ctx, involvedWallet, normalizedTokenAddress, newTokenBalance); err != nil {
		s.log.Error("Failed to update token balance in repository",
			logger.Int64("wallet_id", involvedWallet.ID),
			logger.String("token_address", normalizedTokenAddress),
			logger.Error(err))
		return err
	}

	// Deduct native gas cost ONLY if the wallet was the sender
	if isOutgoing {
		// Use GasUsed and GasPrice from the embedded BaseTransaction
		if transfer.GasUsed > 0 && transfer.GasPrice != nil && transfer.GasPrice.Sign() > 0 {
			currentNativeBalance := involvedWallet.Balance.ToBigInt()
			gasUsed := new(big.Int).SetUint64(transfer.GasUsed)
			gasCost := new(big.Int).Mul(transfer.GasPrice, gasUsed)
			newNativeBalance := new(big.Int).Sub(currentNativeBalance, gasCost)
			if newNativeBalance.Sign() < 0 {
				newNativeBalance = big.NewInt(0)
			}
			if err := s.repository.UpdateBalance(ctx, involvedWallet, newNativeBalance); err != nil {
				s.log.Error("Failed to update sender native balance for gas during token transfer",
					logger.Int64("wallet_id", involvedWallet.ID),
					logger.String("tx_hash", transfer.Hash),
					logger.Error(err))
			}
		} else {
			s.log.Warn("Missing gas details in ERC20Transfer, cannot deduct native gas cost",
				logger.String("tx_hash", transfer.Hash),
				logger.Int64("wallet_id", involvedWallet.ID))
		}
	}

	return nil
}

// findWalletForAddress is a helper to check if an address belongs to a managed wallet.
// Returns (found bool, *Wallet, error).
func (s *balanceService) findWalletForAddress(ctx context.Context, chain types.ChainType, address string) (bool, *Wallet, error) {
	if address == "" {
		return false, nil, nil // Empty address cannot be a wallet
	}
	wallet, err := s.repository.GetByAddress(ctx, chain, address)
	if err != nil {
		appErr, ok := err.(*errors.Vault0Error)
		if ok && appErr.Code == errors.ErrCodeNotFound {
			return false, nil, nil // Not found is not an error here, just means address isn't ours
		}
		return false, nil, err
	}
	return true, wallet, nil
}

// GetWalletBalances retrieves the native and token balances for a wallet
func (s *balanceService) GetWalletBalances(ctx context.Context, id int64) ([]*TokenBalanceData, error) {
	if id == 0 {
		return nil, errors.NewInvalidInputError("ID is required", "id", "0")
	}

	// Get the wallet to access its native balance and chain type
	wallet, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get token balances
	tokenBalances, err := s.repository.GetTokenBalances(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create result slice with enough capacity for native token + all token balances
	result := make([]*TokenBalanceData, 0, len(tokenBalances)+1)

	// Get the native token for this chain
	nativeToken, err := wallet.GetToken()
	if err != nil {
		return nil, err
	}

	// Add native token balance
	result = append(result, &TokenBalanceData{
		Token:     nativeToken,
		Balance:   wallet.Balance.ToBigInt(),
		UpdatedAt: wallet.UpdatedAt,
	})

	// If there are no token balances, return the native token balance
	if len(tokenBalances) == 0 {
		return result, nil
	}

	// Process each token balance
	for _, tb := range tokenBalances {
		// Get token by address
		token, err := s.tokenStore.GetToken(ctx, tb.TokenAddress)
		if err != nil {
			s.log.Warn("Could not find token",
				logger.String("token_address", tb.TokenAddress),
				logger.Error(err))
			continue
		}

		result = append(result, &TokenBalanceData{
			Token:     token,
			Balance:   tb.Balance.ToBigInt(),
			UpdatedAt: tb.UpdatedAt,
		})
	}

	return result, nil
}

// GetWalletBalancesByAddress retrieves the native and token balances for a wallet by its address
func (s *balanceService) GetWalletBalancesByAddress(ctx context.Context, chainType types.ChainType, address string) ([]*TokenBalanceData, error) {
	if chainType == "" {
		return nil, errors.NewInvalidInputError("Chain type is required", "chain_type", "")
	}

	if address == "" {
		return nil, errors.NewInvalidInputError("Address is required", "address", "")
	}

	normalizedAddr, err := types.NewAddress(chainType, address)
	if err != nil {
		return nil, err
	}
	normalizedAddressStr := normalizedAddr.ToChecksum()

	wallet, err := s.repository.GetByAddress(ctx, chainType, normalizedAddressStr)
	if err != nil {
		return nil, err
	}

	return s.GetWalletBalances(ctx, wallet.ID)
}
