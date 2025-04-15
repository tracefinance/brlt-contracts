package wallet

import (
	"context"
	"math/big"

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

	// UpdateTokenBalance updates the token balance for a wallet based on a transaction.
	// It identifies the involved wallet (sender or receiver) and updates the specific
	// token balance. For senders, it also deducts the native currency gas cost.
	// Parameters:
	//   - ctx: Context for the operation
	//   - tx: The transaction details (must be of type ERC20)
	// Returns:
	//   - error: ErrWalletNotFound, ErrTokenNotFound, or other processing errors
	UpdateTokenBalance(ctx context.Context, tx *types.Transaction) error

	// GetWalletBalances retrieves the native and token balances for a wallet
	GetWalletBalances(ctx context.Context, id int64) ([]*TokenBalanceData, error)

	// GetWalletBalancesByAddress retrieves the native and token balances for a wallet by its address
	GetWalletBalancesByAddress(ctx context.Context, chainType types.ChainType, address string) ([]*TokenBalanceData, error)
}

// isOutgoingTransaction returns (isOutgoingTransaction, error)
// isOutgoingTransaction is true if the wallet is the sender (tx.From), false if receiver (tx.To)
func (s *walletService) isOutgoingTransaction(ctx context.Context, tx *types.Transaction) (bool, error) {
	exists, err := s.repository.Exists(ctx, tx.Chain, tx.From)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil // Sender
	}

	exists, err = s.repository.Exists(ctx, tx.Chain, tx.To)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil // Receiver
	}

	return false, errors.NewWalletNotFoundError(tx.From + " or " + tx.To)
}

// UpdateWalletBalance updates the native balance for a wallet based on a transaction
func (s *walletService) UpdateWalletBalance(ctx context.Context, tx *types.Transaction) error {
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
		wallet, err = s.repository.GetByAddress(ctx, tx.Chain, tx.From)
	} else {
		wallet, err = s.repository.GetByAddress(ctx, tx.Chain, tx.To)
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

// UpdateTokenBalance updates the token balance for a wallet based on a transaction
func (s *walletService) UpdateTokenBalance(ctx context.Context, tx *types.Transaction) error {
	if tx == nil {
		return errors.NewInvalidInputError("Transaction is required", "tx", nil)
	}
	if tx.Type != types.TransactionTypeERC20 {
		return errors.NewInvalidInputError("Transaction must be ERC20 type", "tx.Type", tx.Type)
	}
	if tx.TokenAddress == "" {
		return errors.NewInvalidInputError("Token address is required for ERC20 transaction", "tx.TokenAddress", "")
	}

	isOutgoing, err := s.isOutgoingTransaction(ctx, tx)
	if err != nil {
		return err
	}

	var wallet *Wallet
	if isOutgoing {
		wallet, err = s.repository.GetByAddress(ctx, tx.Chain, tx.From)
	} else {
		wallet, err = s.repository.GetByAddress(ctx, tx.Chain, tx.To)
	}
	if err != nil {
		return err
	}

	tokenBalances, err := s.repository.GetTokenBalances(ctx, wallet.ID)
	if err != nil {
		return err
	}

	var currentTokenBalance *big.Int
	found := false
	for _, tb := range tokenBalances {
		if tb.TokenAddress == tx.TokenAddress {
			currentTokenBalance = tb.Balance.ToBigInt()
			found = true
			break
		}
	}
	if !found {
		currentTokenBalance = big.NewInt(0)
	}

	var newTokenBalance *big.Int
	if isOutgoing {
		newTokenBalance = new(big.Int).Sub(currentTokenBalance, tx.Value)
		if newTokenBalance.Sign() < 0 {
			newTokenBalance = big.NewInt(0)
		}
	} else {
		newTokenBalance = new(big.Int).Add(currentTokenBalance, tx.Value)
	}

	if err := s.repository.UpdateTokenBalance(ctx, wallet, tx.TokenAddress, newTokenBalance); err != nil {
		return err
	}

	if isOutgoing {
		currentNativeBalance := wallet.Balance.ToBigInt()
		gasUsed := new(big.Int).SetUint64(tx.GasUsed)
		gasCost := new(big.Int).Mul(tx.GasPrice, gasUsed)
		newNativeBalance := new(big.Int).Sub(currentNativeBalance, gasCost)
		if newNativeBalance.Sign() < 0 {
			newNativeBalance = big.NewInt(0)
		}
		if err := s.repository.UpdateBalance(ctx, wallet, newNativeBalance); err != nil {
			s.log.Error("Failed to update sender native balance for gas during token transfer",
				logger.Int64("wallet_id", wallet.ID),
				logger.String("tx_hash", tx.Hash),
				logger.Error(err))
		}
	}

	return nil
}

// GetWalletBalances retrieves the native and token balances for a wallet
func (s *walletService) GetWalletBalances(ctx context.Context, id int64) ([]*TokenBalanceData, error) {
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
func (s *walletService) GetWalletBalancesByAddress(ctx context.Context, chainType types.ChainType, address string) ([]*TokenBalanceData, error) {
	if chainType == "" {
		return nil, errors.NewInvalidInputError("Chain type is required", "chain_type", "")
	}

	if address == "" {
		return nil, errors.NewInvalidInputError("Address is required", "address", "")
	}

	chain, err := s.chains.Get(chainType)
	if err != nil {
		return nil, err
	}

	if !chain.IsValidAddress(address) {
		return nil, errors.NewInvalidAddressError(address)
	}

	wallet, err := s.repository.GetByAddress(ctx, chainType, address)
	if err != nil {
		return nil, err
	}

	return s.GetWalletBalances(ctx, wallet.ID)
}
