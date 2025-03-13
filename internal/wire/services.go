package wire

import (
	"github.com/google/wire"

	"vault0/internal/services/blockchain"
	"vault0/internal/services/user"
	"vault0/internal/services/wallet"
)

type Services struct {
	WalletService     wallet.Service
	UserService       user.Service
	BlockchainService blockchain.Service
}

func NewServices(
	walletService wallet.Service,
	userService user.Service,
	blockchainService blockchain.Service,
) *Services {
	return &Services{
		WalletService:     walletService,
		UserService:       userService,
		BlockchainService: blockchainService,
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

var BlockchainServiceSet = wire.NewSet(
	blockchain.NewRepository,
	blockchain.NewService,
)

var ServicesSet = wire.NewSet(
	WalletServiceSet,
	UserServiceSet,
	BlockchainServiceSet,
	NewServices,
)
