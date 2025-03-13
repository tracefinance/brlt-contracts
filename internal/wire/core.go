package wire

import (
	"github.com/google/wire"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/contract"
	"vault0/internal/core/db"
	"vault0/internal/core/keystore"
	"vault0/internal/core/wallet"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// CoreSet combines all core dependencies
var CoreSet = wire.NewSet(
	config.LoadConfig,
	db.NewDatabase,
	logger.NewLogger,
	keystore.NewKeyStore,
	blockchain.NewRegistry,
	wallet.NewFactory,
	contract.NewSmartContract,
	types.NewChains,
)
