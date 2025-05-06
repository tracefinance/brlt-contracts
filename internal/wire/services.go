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
	Service               transaction.Service
	TransformerService    transaction.TransformerService
	PoolingService        transaction.PoolingService
	MonitorService        transaction.MonitorService
	HistoryService        transaction.HistoryService
	BlockchainTransformer transaction.BlockchainTransformer
}

// Services holds instances of all application services.
type Services struct {
	WalletService            wallet.Service
	WalletMonitorService     wallet.WalletMonitor
	UserService              user.Service
	Transaction              Transaction
	TokenService             token.Service
	TokenMonitorService      token.TokenMonitorService
	SignerService            signer.Service
	TokenPriceService        tokenprice.Service
	TokenPricePollingService tokenprice.PricePoolingService
	KeystoreService          keystore.Service
	VaultService             vault.Service
}

// Define Wire provider sets for each service
var WalletServiceSet = wire.NewSet(
	wallet.NewRepository,
	wallet.NewService,
	wallet.NewBalanceService,
	wallet.NewWalletMonitorService,
)
var UserServiceSet = wire.NewSet(user.NewRepository, user.NewService)
var TransactionServiceSet = wire.NewSet(
	transaction.NewRepository,
	transaction.NewService,
	transaction.NewTransformerService,
	transaction.NewPoolingService,
	transaction.NewMonitorService,
	transaction.NewHistoryService,
	transaction.NewBlockchainTransformer,
)
var TokenServiceSet = wire.NewSet(token.NewService, token.NewTokenMonitorService)
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
	walletMonitorSvc wallet.WalletMonitor,
	userSvc user.Service,
	transactionSvc transaction.Service,
	transformerSvc transaction.TransformerService,
	poolingSvc transaction.PoolingService,
	monitorSvc transaction.MonitorService,
	tokenMonitorSvc token.TokenMonitorService,
	historySvc transaction.HistoryService,
	tokenSvc token.Service,
	signerSvc signer.Service,
	tokenPriceSvc tokenprice.Service,
	tokenPricePollingSvc tokenprice.PricePoolingService,
	keystoreSvc keystore.Service,
	vaultSvc vault.Service,
	blockchainTransformer transaction.BlockchainTransformer,
) *Services {
	return &Services{
		Transaction: Transaction{
			Service:               transactionSvc,
			TransformerService:    transformerSvc,
			PoolingService:        poolingSvc,
			MonitorService:        monitorSvc,
			HistoryService:        historySvc,
			BlockchainTransformer: blockchainTransformer,
		},
		WalletService:            walletSvc,
		WalletMonitorService:     walletMonitorSvc,
		UserService:              userSvc,
		TokenMonitorService:      tokenMonitorSvc,
		TokenService:             tokenSvc,
		SignerService:            signerSvc,
		TokenPriceService:        tokenPriceSvc,
		TokenPricePollingService: tokenPricePollingSvc,
		KeystoreService:          keystoreSvc,
		VaultService:             vaultSvc,
	}
}
