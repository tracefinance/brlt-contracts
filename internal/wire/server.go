package wire

import (
	"vault0/internal/api/handlers/blockchain"
	"vault0/internal/api/handlers/transaction"
	"vault0/internal/api/handlers/user"
	"vault0/internal/api/handlers/wallet"

	"vault0/internal/api"

	"github.com/google/wire"
)

var ServerSet = wire.NewSet(
	wallet.NewHandler,
	user.NewHandler,
	blockchain.NewHandler,
	transaction.NewHandler,
	api.NewServer,
)
