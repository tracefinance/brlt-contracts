package wire

import (
	"vault0/internal/api/handlers/keystore"
	"vault0/internal/api/handlers/reference"
	"vault0/internal/api/handlers/signer"
	"vault0/internal/api/handlers/token"
	"vault0/internal/api/handlers/tokenprice"
	"vault0/internal/api/handlers/transaction"
	"vault0/internal/api/handlers/user"
	"vault0/internal/api/handlers/vault"
	"vault0/internal/api/handlers/wallet"

	"vault0/internal/api"

	"github.com/google/wire"
)

var ServerSet = wire.NewSet(
	wallet.NewHandler,
	user.NewHandler,
	transaction.NewHandler,
	token.NewHandler,
	signer.NewHandler,
	tokenprice.NewHandler,
	reference.NewHandler,
	keystore.NewHandler,
	vault.NewHandler,
	api.NewServer,
)
