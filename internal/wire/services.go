package wire

import (
	"github.com/google/wire"

	"vault0/internal/services/keystore"
	"vault0/internal/services/signer"
	"vault0/internal/services/token"
	"vault0/internal/services/tokenprice"
	"vault0/internal/services/transaction"
	"vault0/internal/services/user"
	"vault0/internal/services/vault"
	"vault0/internal/services/wallet"
)

type Transaction struct {
	Service             transaction.Service
	TransformerService  transaction.TransformerService
	PoolingService      transaction.PoolingService
	MonitorService      transaction.MonitorService
	TokenMonitorService transaction.TokenMonitorService
}

// Services holds instances of all application services.
type Services struct {
	WalletService            wallet.Service
	UserService              user.Service
	Transaction              Transaction
	TokenService             token.Service
	SignerService            signer.Service
	TokenPriceService        tokenprice.Service
	TokenPricePollingService tokenprice.PricePoolingService
	KeystoreService          keystore.Service
	VaultService             vault.Service
}

// Define Wire provider sets for each service
var WalletServiceSet = wire.NewSet(wallet.NewRepository, wallet.NewService)
var UserServiceSet = wire.NewSet(user.NewRepository, user.NewService)
var TransactionServiceSet = wire.NewSet(
	transaction.NewRepository,
	transaction.NewService,
	transaction.NewTransformerService,
	transaction.NewPoolingService,
	transaction.NewMonitorService,
	transaction.NewTokenMonitorService,
)
var TokenServiceSet = wire.NewSet(token.NewService)
var SignerServiceSet = wire.NewSet(signer.NewRepository, signer.NewService)
var TokenPriceServiceSet = wire.NewSet(tokenprice.NewRepository, tokenprice.NewService, tokenprice.NewPollingService)
var KeystoreServiceSet = wire.NewSet(keystore.NewService)
var VaultServiceSet = wire.NewSet(vault.NewRepository, vault.NewService)

// Define the set for all services
var ServicesSet = wire.NewSet(
	WalletServiceSet,
	UserServiceSet,
	TransactionServiceSet,
	TokenServiceSet,
	SignerServiceSet,
	TokenPriceServiceSet,
	KeystoreServiceSet,
	VaultServiceSet,
	NewServices,
)

// NewServices creates the Services struct (used by Wire).
func NewServices(
	walletSvc wallet.Service,
	userSvc user.Service,
	transactionSvc transaction.Service,
	transformerSvc transaction.TransformerService,
	poolingSvc transaction.PoolingService,
	monitorSvc transaction.MonitorService,
	tokenMonitorSvc transaction.TokenMonitorService,
	tokenSvc token.Service,
	signerSvc signer.Service,
	tokenPriceSvc tokenprice.Service,
	tokenPricePollingSvc tokenprice.PricePoolingService,
	keystoreSvc keystore.Service,
	vaultSvc vault.Service,
) *Services {
	return &Services{
		Transaction: Transaction{
			Service:             transactionSvc,
			TransformerService:  transformerSvc,
			PoolingService:      poolingSvc,
			MonitorService:      monitorSvc,
			TokenMonitorService: tokenMonitorSvc,
		},
		WalletService: walletSvc,
		UserService:   userSvc,

		TokenService:             tokenSvc,
		SignerService:            signerSvc,
		TokenPriceService:        tokenPriceSvc,
		TokenPricePollingService: tokenPricePollingSvc,
		KeystoreService:          keystoreSvc,
		VaultService:             vaultSvc,
	}
}
