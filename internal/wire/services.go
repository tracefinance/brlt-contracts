package wire

import (
	"github.com/google/wire"

	"vault0/internal/services/signer"
	"vault0/internal/services/token"
	"vault0/internal/services/transaction"
	"vault0/internal/services/user"
	"vault0/internal/services/wallet"
)

type Services struct {
	WalletService      wallet.Service
	UserService        user.Service
	TransactionService transaction.Service
	TokenService       token.Service
	SignerService      signer.Service
}

func NewServices(
	walletService wallet.Service,
	userService user.Service,
	transactionService transaction.Service,
	tokenService token.Service,
	signerService signer.Service,
) *Services {
	return &Services{
		WalletService:      walletService,
		UserService:        userService,
		TransactionService: transactionService,
		TokenService:       tokenService,
		SignerService:      signerService,
	}
}

var WalletServiceSet = wire.NewSet(
	wallet.NewRepository,
	wallet.NewService,
)

var UserServiceSet = wire.NewSet(
	user.NewRepository,
	user.NewService,
)

var TransactionServiceSet = wire.NewSet(
	transaction.NewRepository,
	transaction.NewService,
)

var TokenServiceSet = wire.NewSet(
	token.NewService,
)

var SignerServiceSet = wire.NewSet(
	signer.NewRepository,
	signer.NewService,
)

var ServicesSet = wire.NewSet(
	WalletServiceSet,
	UserServiceSet,
	TransactionServiceSet,
	TokenServiceSet,
	SignerServiceSet,
	NewServices,
)
