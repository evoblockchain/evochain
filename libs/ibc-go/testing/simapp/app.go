package simapp

import (
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/evoblockchain/evochain/app/ante"
	evoblockchainchaincodec "github.com/evoblockchain/evochain/app/codec"
	appconfig "github.com/evoblockchain/evochain/app/config"
	"github.com/evoblockchain/evochain/app/refund"
	evoblockchain "github.com/evoblockchain/evochain/app/types"
	"github.com/evoblockchain/evochain/app/utils/sanity"
	bam "github.com/evoblockchain/evochain/libs/cosmos-sdk/baseapp"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/client"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/codec"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/server"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/simapp"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/store/mpt"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/store/types"
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/types/module"
	upgradetypes "github.com/evoblockchain/evochain/libs/cosmos-sdk/types/upgrade"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/version"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/auth"
	authante "github.com/evoblockchain/evochain/libs/cosmos-sdk/x/auth/ante"
	authtypes "github.com/evoblockchain/evochain/libs/cosmos-sdk/x/auth/types"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/bank"
	capabilityModule "github.com/evoblockchain/evochain/libs/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/evoblockchain/evochain/libs/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/evoblockchain/evochain/libs/cosmos-sdk/x/capability/types"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/crisis"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/mint"
	govclient "github.com/evoblockchain/evochain/libs/cosmos-sdk/x/mint/client"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/params/subspace"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/supply"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/upgrade"
	ibctransferkeeper "github.com/evoblockchain/evochain/libs/ibc-go/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/evoblockchain/evochain/libs/ibc-go/modules/apps/transfer/types"
	ibc "github.com/evoblockchain/evochain/libs/ibc-go/modules/core"
	ibcclient "github.com/evoblockchain/evochain/libs/ibc-go/modules/core/02-client"
	ibcporttypes "github.com/evoblockchain/evochain/libs/ibc-go/modules/core/05-port/types"
	ibchost "github.com/evoblockchain/evochain/libs/ibc-go/modules/core/24-host"
	"github.com/evoblockchain/evochain/libs/ibc-go/testing/mock"
	"github.com/evoblockchain/evochain/libs/ibc-go/testing/simapp/adapter/capability"
	"github.com/evoblockchain/evochain/libs/ibc-go/testing/simapp/adapter/core"
	staking2 "github.com/evoblockchain/evochain/libs/ibc-go/testing/simapp/adapter/staking"
	"github.com/evoblockchain/evochain/libs/ibc-go/testing/simapp/adapter/transfer"
	"github.com/evoblockchain/evochain/libs/system/trace"
	abci "github.com/evoblockchain/evochain/libs/tendermint/abci/types"
	"github.com/evoblockchain/evochain/libs/tendermint/libs/cli"
	"github.com/evoblockchain/evochain/libs/tendermint/libs/log"
	tmos "github.com/evoblockchain/evochain/libs/tendermint/libs/os"
	tmtypes "github.com/evoblockchain/evochain/libs/tendermint/types"
	dbm "github.com/evoblockchain/evochain/libs/tm-db"
	"github.com/evoblockchain/evochain/x/ammswap"
	"github.com/evoblockchain/evochain/x/common/monitor"
	commonversion "github.com/evoblockchain/evochain/x/common/version"
	"github.com/evoblockchain/evochain/x/dex"
	dexclient "github.com/evoblockchain/evochain/x/dex/client"
	distr "github.com/evoblockchain/evochain/x/distribution"
	"github.com/evoblockchain/evochain/x/erc20"
	erc20client "github.com/evoblockchain/evochain/x/erc20/client"
	"github.com/evoblockchain/evochain/x/evidence"
	"github.com/evoblockchain/evochain/x/evm"
	evmclient "github.com/evoblockchain/evochain/x/evm/client"
	evmtypes "github.com/evoblockchain/evochain/x/evm/types"
	"github.com/evoblockchain/evochain/x/farm"
	farmclient "github.com/evoblockchain/evochain/x/farm/client"
	"github.com/evoblockchain/evochain/x/genutil"
	"github.com/evoblockchain/evochain/x/gov"
	"github.com/evoblockchain/evochain/x/gov/keeper"
	"github.com/evoblockchain/evochain/x/order"
	"github.com/evoblockchain/evochain/x/params"
	paramsclient "github.com/evoblockchain/evochain/x/params/client"
	"github.com/evoblockchain/evochain/x/slashing"
	"github.com/evoblockchain/evochain/x/staking"
	"github.com/evoblockchain/evochain/x/token"
	"github.com/evoblockchain/evochain/x/wasm"
	wasmclient "github.com/evoblockchain/evochain/x/wasm/client"
	wasmkeeper "github.com/evoblockchain/evochain/x/wasm/keeper"
	"github.com/spf13/viper"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"
)

func init() {
	// set the address prefixes
	config := sdk.GetConfig()
	config.SetCoinType(60)
	evoblockchain.SetBech32Prefixes(config)
	evoblockchain.SetBip44CoinType(config)
}

const (
	appName = "EVOChain"
)

var (
	// DefaultCLIHome sets the default home directories for the application CLI
	DefaultCLIHome = os.ExpandEnv("$HOME/.evochaincli")

	// DefaultNodeHome sets the folder where the applcation data and configuration will be stored
	DefaultNodeHome = os.ExpandEnv("$HOME/.evochaind")

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		supply.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(
			paramsclient.ProposalHandler,
			distr.CommunityPoolSpendProposalHandler,
			distr.ChangeDistributionTypeProposalHandler,
			distr.WithdrawRewardEnabledProposalHandler,
			distr.RewardTruncatePrecisionProposalHandler,
			dexclient.DelistProposalHandler, farmclient.ManageWhiteListProposalHandler,
			evmclient.ManageContractDeploymentWhitelistProposalHandler,
			evmclient.ManageContractBlockedListProposalHandler,
			evmclient.ManageContractMethodBlockedListProposalHandler,
			evmclient.ManageSysContractAddressProposalHandler,
			govclient.ManageTreasuresProposalHandler,
			erc20client.TokenMappingProposalHandler,
			erc20client.ProxyContractRedirectHandler,
			wasmclient.MigrateContractProposalHandler,
			wasmclient.UpdateContractAdminProposalHandler,
			wasmclient.ClearContractAdminProposalHandler,
			wasmclient.PinCodesProposalHandler,
			wasmclient.UnpinCodesProposalHandler,
			wasmclient.UpdateDeploymentWhitelistProposalHandler,
			wasmclient.UpdateWASMContractMethodBlockedListProposalHandler,
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		evidence.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evm.AppModuleBasic{},
		token.AppModuleBasic{},
		dex.AppModuleBasic{},
		order.AppModuleBasic{},
		ammswap.AppModuleBasic{},
		farm.AppModuleBasic{},
		capabilityModule.AppModuleBasic{},
		core.CoreModule{},
		capability.CapabilityModuleAdapter{},
		transfer.TransferModule{},
		erc20.AppModuleBasic{},
		mock.AppModuleBasic{},
		wasm.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		auth.FeeCollectorName:       nil,
		distr.ModuleName:            nil,
		mint.ModuleName:             {supply.Minter},
		staking.BondedPoolName:      {supply.Burner, supply.Staking},
		staking.NotBondedPoolName:   {supply.Burner, supply.Staking},
		gov.ModuleName:              nil,
		token.ModuleName:            {supply.Minter, supply.Burner},
		dex.ModuleName:              nil,
		order.ModuleName:            nil,
		ammswap.ModuleName:          {supply.Minter, supply.Burner},
		farm.ModuleName:             nil,
		farm.YieldFarmingAccount:    nil,
		farm.MintFarmingAccount:     {supply.Burner},
		ibctransfertypes.ModuleName: {authtypes.Minter, authtypes.Burner},
		erc20.ModuleName:            {authtypes.Minter, authtypes.Burner},
	}

	GlobalGpIndex = GasPriceIndex{}

	onceLog sync.Once
)

type GasPriceIndex struct {
	RecommendGp *big.Int `json:"recommend-gp"`
}

var _ simapp.App = (*SimApp)(nil)

// SimApp implements an extended ABCI application. It is an application
// that may process transactions through Ethereum's EVM running atop of
// Tendermint consensus.
type SimApp struct {
	*bam.BaseApp

	txconfig client.TxConfig

	invCheckPeriod uint

	// keys to access the substores
	keys  map[string]*sdk.KVStoreKey
	tkeys map[string]*sdk.TransientStoreKey

	// subspaces
	subspaces map[string]params.Subspace

	// keepers
	AccountKeeper  auth.AccountKeeper
	BankKeeper     bank.Keeper
	SupplyKeeper   supply.Keeper
	StakingKeeper  staking.Keeper
	SlashingKeeper slashing.Keeper
	MintKeeper     mint.Keeper
	DistrKeeper    distr.Keeper
	GovKeeper      gov.Keeper
	CrisisKeeper   crisis.Keeper
	UpgradeKeeper  upgrade.Keeper
	ParamsKeeper   params.Keeper
	EvidenceKeeper evidence.Keeper
	EvmKeeper      *evm.Keeper
	TokenKeeper    token.Keeper
	DexKeeper      dex.Keeper
	OrderKeeper    order.Keeper
	SwapKeeper     ammswap.Keeper
	FarmKeeper     farm.Keeper
	wasmKeeper     wasm.Keeper

	// the module manager
	mm *module.Manager

	// simulation manager
	sm *module.SimulationManager

	blockGasPrice []*big.Int

	configurator module.Configurator
	// ibc
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper
	ScopedIBCMockKeeper  capabilitykeeper.ScopedKeeper
	TransferKeeper       ibctransferkeeper.Keeper
	CapabilityKeeper     *capabilitykeeper.Keeper
	IBCKeeper            *ibc.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	marshal              *codec.CodecProxy
	heightTasks          map[int64]*upgradetypes.HeightTasks
	Erc20Keeper          erc20.Keeper

	ibcScopeKeep capabilitykeeper.ScopedKeeper
	WasmHandler  wasmkeeper.HandlerOption
}

// NewSimApp returns a reference to a new initialized EVOChain application.
func NewSimApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	invCheckPeriod uint,
	baseAppOptions ...func(*bam.BaseApp),
) *SimApp {
	logger.Info("Starting OEC",
		"GenesisHeight", tmtypes.GetStartBlockHeight(),
		"MercuryHeight", tmtypes.GetMercuryHeight(),
		"VenusHeight", tmtypes.GetVenusHeight(),
	)
	//onceLog.Do(func() {
	//	iavl.SetLogger(logger.With("module", "iavl"))
	//	logStartingFlags(logger)
	//})

	codecProxy, interfaceReg := evoblockchainchaincodec.MakeCodecSuit(ModuleBasics)

	// NOTE we use custom EVOChain transaction decoder that supports the sdk.Tx interface instead of sdk.StdTx
	bApp := bam.NewBaseApp(appName, logger, db, evm.TxDecoder(codecProxy), baseAppOptions...)

	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetAppVersion(version.Version)
	bApp.SetStartLogHandler(trace.StartTxLog)
	bApp.SetEndLogHandler(trace.StopTxLog)

	bApp.SetInterfaceRegistry(interfaceReg)

	keys := sdk.NewKVStoreKeys(
		bam.MainStoreKey, auth.StoreKey, staking.StoreKey,
		supply.StoreKey, mint.StoreKey, distr.StoreKey, slashing.StoreKey,
		gov.StoreKey, params.StoreKey, upgrade.StoreKey, evidence.StoreKey,
		evm.StoreKey, token.StoreKey, token.KeyLock, dex.StoreKey, dex.TokenPairStoreKey,
		order.OrderStoreKey, ammswap.StoreKey, farm.StoreKey, ibctransfertypes.StoreKey, capabilitytypes.StoreKey,
		ibchost.StoreKey,
		erc20.StoreKey,
		mpt.StoreKey, wasm.StoreKey,
	)

	tkeys := sdk.NewTransientStoreKeys(params.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	app := &SimApp{
		BaseApp:        bApp,
		invCheckPeriod: invCheckPeriod,
		keys:           keys,
		tkeys:          tkeys,
		subspaces:      make(map[string]params.Subspace),
		heightTasks:    make(map[int64]*upgradetypes.HeightTasks),
	}
	bApp.SetInterceptors(makeInterceptors())

	// init params keeper and subspaces
	app.ParamsKeeper = params.NewKeeper(codecProxy.GetCdc(), keys[params.StoreKey], tkeys[params.TStoreKey])
	app.subspaces[auth.ModuleName] = app.ParamsKeeper.Subspace(auth.DefaultParamspace)
	app.subspaces[bank.ModuleName] = app.ParamsKeeper.Subspace(bank.DefaultParamspace)
	app.subspaces[staking.ModuleName] = app.ParamsKeeper.Subspace(staking.DefaultParamspace)
	app.subspaces[mint.ModuleName] = app.ParamsKeeper.Subspace(mint.DefaultParamspace)
	app.subspaces[distr.ModuleName] = app.ParamsKeeper.Subspace(distr.DefaultParamspace)
	app.subspaces[slashing.ModuleName] = app.ParamsKeeper.Subspace(slashing.DefaultParamspace)
	app.subspaces[gov.ModuleName] = app.ParamsKeeper.Subspace(gov.DefaultParamspace)
	app.subspaces[crisis.ModuleName] = app.ParamsKeeper.Subspace(crisis.DefaultParamspace)
	app.subspaces[evidence.ModuleName] = app.ParamsKeeper.Subspace(evidence.DefaultParamspace)
	app.subspaces[evm.ModuleName] = app.ParamsKeeper.Subspace(evm.DefaultParamspace)
	app.subspaces[token.ModuleName] = app.ParamsKeeper.Subspace(token.DefaultParamspace)
	app.subspaces[dex.ModuleName] = app.ParamsKeeper.Subspace(dex.DefaultParamspace)
	app.subspaces[order.ModuleName] = app.ParamsKeeper.Subspace(order.DefaultParamspace)
	app.subspaces[ammswap.ModuleName] = app.ParamsKeeper.Subspace(ammswap.DefaultParamspace)
	app.subspaces[farm.ModuleName] = app.ParamsKeeper.Subspace(farm.DefaultParamspace)
	app.subspaces[ibchost.ModuleName] = app.ParamsKeeper.Subspace(ibchost.ModuleName)
	app.subspaces[ibctransfertypes.ModuleName] = app.ParamsKeeper.Subspace(ibctransfertypes.ModuleName)
	app.subspaces[erc20.ModuleName] = app.ParamsKeeper.Subspace(erc20.DefaultParamspace)
	app.subspaces[wasm.ModuleName] = app.ParamsKeeper.Subspace(wasm.ModuleName)

	//proxy := codec.NewMarshalProxy(cc, cdc)
	app.marshal = codecProxy
	// use custom EVOChain account for contracts
	app.AccountKeeper = auth.NewAccountKeeper(
		codecProxy.GetCdc(), keys[auth.StoreKey], keys[mpt.StoreKey], app.subspaces[auth.ModuleName], evoblockchain.ProtoAccount,
	)

	bankKeeper := bank.NewBaseKeeperWithMarshal(
		&app.AccountKeeper, codecProxy, app.subspaces[bank.ModuleName], app.ModuleAccountAddrs(),
	)
	app.BankKeeper = &bankKeeper
	app.ParamsKeeper.SetBankKeeper(app.BankKeeper)
	app.SupplyKeeper = supply.NewKeeper(
		codecProxy.GetCdc(), keys[supply.StoreKey], &app.AccountKeeper, app.BankKeeper, maccPerms,
	)

	stakingKeeper := staking2.NewStakingKeeper(
		codecProxy, keys[staking.StoreKey], app.SupplyKeeper, app.subspaces[staking.ModuleName],
	).Keeper
	app.ParamsKeeper.SetStakingKeeper(stakingKeeper)
	app.MintKeeper = mint.NewKeeper(
		codecProxy.GetCdc(), keys[mint.StoreKey], app.subspaces[mint.ModuleName], stakingKeeper,
		app.SupplyKeeper, auth.FeeCollectorName, farm.MintFarmingAccount,
	)
	app.DistrKeeper = distr.NewKeeper(
		codecProxy.GetCdc(), keys[distr.StoreKey], app.subspaces[distr.ModuleName], stakingKeeper,
		app.SupplyKeeper, auth.FeeCollectorName, app.ModuleAccountAddrs(),
	)
	app.SlashingKeeper = slashing.NewKeeper(
		codecProxy.GetCdc(), keys[slashing.StoreKey], stakingKeeper, app.subspaces[slashing.ModuleName],
	)
	app.CrisisKeeper = crisis.NewKeeper(
		app.subspaces[crisis.ModuleName], invCheckPeriod, app.SupplyKeeper, auth.FeeCollectorName,
	)
	app.UpgradeKeeper = upgrade.NewKeeper(skipUpgradeHeights, keys[upgrade.StoreKey], app.marshal.GetCdc())
	app.ParamsKeeper.RegisterSignal(evmtypes.SetEvmParamsNeedUpdate)
	app.EvmKeeper = evm.NewKeeper(
		app.marshal.GetCdc(), keys[evm.StoreKey], app.subspaces[evm.ModuleName], &app.AccountKeeper, app.SupplyKeeper, app.BankKeeper, stakingKeeper, logger)
	(&bankKeeper).SetInnerTxKeeper(app.EvmKeeper)

	app.TokenKeeper = token.NewKeeper(app.BankKeeper, app.subspaces[token.ModuleName], auth.FeeCollectorName, app.SupplyKeeper,
		keys[token.StoreKey], keys[token.KeyLock], app.marshal.GetCdc(), false, &app.AccountKeeper)

	app.DexKeeper = dex.NewKeeper(auth.FeeCollectorName, app.SupplyKeeper, app.subspaces[dex.ModuleName], app.TokenKeeper, stakingKeeper,
		app.BankKeeper, app.keys[dex.StoreKey], app.keys[dex.TokenPairStoreKey], app.marshal.GetCdc())

	app.OrderKeeper = order.NewKeeper(
		app.TokenKeeper, app.SupplyKeeper, app.DexKeeper, app.subspaces[order.ModuleName], auth.FeeCollectorName,
		app.keys[order.OrderStoreKey], app.marshal.GetCdc(), false, monitor.NopOrderMetrics())

	app.SwapKeeper = ammswap.NewKeeper(app.SupplyKeeper, app.TokenKeeper, app.marshal.GetCdc(), app.keys[ammswap.StoreKey], app.subspaces[ammswap.ModuleName])

	app.FarmKeeper = farm.NewKeeper(auth.FeeCollectorName, app.SupplyKeeper, app.TokenKeeper, app.SwapKeeper, *app.EvmKeeper, app.subspaces[farm.StoreKey],
		app.keys[farm.StoreKey], app.marshal.GetCdc())

	// create evidence keeper with router
	evidenceKeeper := evidence.NewKeeper(
		codecProxy.GetCdc(), keys[evidence.StoreKey], app.subspaces[evidence.ModuleName], &app.StakingKeeper, app.SlashingKeeper,
	)
	evidenceRouter := evidence.NewRouter()
	evidenceKeeper.SetRouter(evidenceRouter)
	app.EvidenceKeeper = *evidenceKeeper

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(codecProxy, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])
	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	app.ibcScopeKeep = scopedIBCKeeper
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	// NOTE: the IBC mock keeper and application module is used only for testing core IBC. Do
	// note replicate if you do not need to test core IBC or light clients.
	scopedIBCMockKeeper := app.CapabilityKeeper.ScopeToModule(mock.ModuleName)

	app.IBCKeeper = ibc.NewKeeper(
		codecProxy, keys[ibchost.StoreKey], app.GetSubspace(ibchost.ModuleName), stakingKeeper, app.UpgradeKeeper, &scopedIBCKeeper, interfaceReg,
	)

	// Create Transfer Keepers
	app.TransferKeeper = ibctransferkeeper.NewKeeper(
		codecProxy, keys[ibctransfertypes.StoreKey], app.GetSubspace(ibctransfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper, &app.IBCKeeper.PortKeeper,
		app.SupplyKeeper, app.SupplyKeeper, scopedTransferKeeper, interfaceReg,
	)
	ibctransfertypes.SetMarshal(codecProxy)

	app.Erc20Keeper = erc20.NewKeeper(app.marshal.GetCdc(), app.keys[erc20.ModuleName], app.subspaces[erc20.ModuleName],
		app.AccountKeeper, app.SupplyKeeper, app.BankKeeper, app.EvmKeeper, app.TransferKeeper)

	// register the proposal types
	// 3.register the proposal types
	govRouter := gov.NewRouter()
	govRouter.AddRoute(gov.RouterKey, gov.ProposalHandler).
		AddRoute(params.RouterKey, params.NewParamChangeProposalHandler(&app.ParamsKeeper)).
		AddRoute(distr.RouterKey, distr.NewDistributionProposalHandler(app.DistrKeeper)).
		AddRoute(dex.RouterKey, dex.NewProposalHandler(&app.DexKeeper)).
		AddRoute(farm.RouterKey, farm.NewManageWhiteListProposalHandler(&app.FarmKeeper)).
		AddRoute(evm.RouterKey, evm.NewManageContractDeploymentWhitelistProposalHandler(app.EvmKeeper)).
		AddRoute(mint.RouterKey, mint.NewManageTreasuresProposalHandler(&app.MintKeeper)).
		AddRoute(ibchost.RouterKey, ibcclient.NewClientUpdateProposalHandler(app.IBCKeeper.ClientKeeper)).
		AddRoute(erc20.RouterKey, erc20.NewProposalHandler(&app.Erc20Keeper))
	govProposalHandlerRouter := keeper.NewProposalHandlerRouter()
	govProposalHandlerRouter.AddRoute(params.RouterKey, &app.ParamsKeeper).
		AddRoute(dex.RouterKey, &app.DexKeeper).
		AddRoute(farm.RouterKey, &app.FarmKeeper).
		AddRoute(evm.RouterKey, app.EvmKeeper).
		AddRoute(mint.RouterKey, &app.MintKeeper).
		AddRoute(erc20.RouterKey, &app.Erc20Keeper)
	app.GovKeeper = gov.NewKeeper(
		app.marshal.GetCdc(), app.keys[gov.StoreKey], app.ParamsKeeper, app.subspaces[gov.DefaultParamspace],
		app.SupplyKeeper, stakingKeeper, gov.DefaultParamspace, govRouter,
		app.BankKeeper, govProposalHandlerRouter, auth.FeeCollectorName,
	)
	app.ParamsKeeper.SetGovKeeper(app.GovKeeper)
	app.DexKeeper.SetGovKeeper(app.GovKeeper)
	app.FarmKeeper.SetGovKeeper(app.GovKeeper)
	app.EvmKeeper.SetGovKeeper(app.GovKeeper)
	app.MintKeeper.SetGovKeeper(app.GovKeeper)
	app.Erc20Keeper.SetGovKeeper(app.GovKeeper)

	// Set EVM hooks
	//app.EvmKeeper.SetHooks(evm.NewLogProcessEvmHook(erc20.NewSendToIbcEventHandler(app.Erc20Keeper)))
	// Set IBC hooks
	//app.TransferKeeper = *app.TransferKeeper.SetHooks(erc20.NewIBCTransferHooks(app.Erc20Keeper))
	//transferModule := ibctransfer.NewAppModule(app.TransferKeeper, codecProxy)
	transferModule := transfer.TNewTransferModule(app.TransferKeeper, codecProxy)

	mockModule := mock.NewAppModule(scopedIBCMockKeeper, &app.IBCKeeper.PortKeeper)
	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := ibcporttypes.NewRouter()
	ibcRouter.AddRoute(ibctransfertypes.ModuleName, transferModule)
	ibcRouter.AddRoute(mock.ModuleName, mockModule)
	//ibcRouter.AddRoute(ibcmock.ModuleName, mockModule)
	app.IBCKeeper.SetRouter(ibcRouter)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper = *stakingKeeper.SetHooks(
		staking.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks()),
	)

	homeDir := viper.GetString(cli.HomeFlag)
	wasmDir := filepath.Join(homeDir, "wasm")
	wasmConfig, err := wasm.ReadWasmConfig()
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}

	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	supportedFeatures := wasm.SupportedFeatures
	app.wasmKeeper = wasm.NewKeeper(
		app.marshal,
		keys[wasm.StoreKey],
		app.subspaces[wasm.ModuleName],
		&app.AccountKeeper,
		bank.NewBankKeeperAdapter(app.BankKeeper),
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		nil,
		app.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		supportedFeatures,
	)

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(
		genutil.NewAppModule(app.AccountKeeper, app.StakingKeeper, app.BaseApp.DeliverTx),
		auth.NewAppModule(app.AccountKeeper),
		bank.NewAppModule(app.BankKeeper, app.AccountKeeper, app.SupplyKeeper),
		crisis.NewAppModule(&app.CrisisKeeper),
		supply.NewAppModule(app.SupplyKeeper, app.AccountKeeper),
		gov.NewAppModule(app.GovKeeper, app.SupplyKeeper),
		mint.NewAppModule(app.MintKeeper),
		slashing.NewAppModule(app.SlashingKeeper, app.AccountKeeper, app.StakingKeeper),
		distr.NewAppModule(app.DistrKeeper, app.SupplyKeeper),
		staking2.TNewStakingModule(app.StakingKeeper, app.AccountKeeper, app.SupplyKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		evm.NewAppModule(app.EvmKeeper, &app.AccountKeeper),
		token.NewAppModule(commonversion.ProtocolVersionV0, app.TokenKeeper, app.SupplyKeeper),
		dex.NewAppModule(commonversion.ProtocolVersionV0, app.DexKeeper, app.SupplyKeeper),
		order.NewAppModule(commonversion.ProtocolVersionV0, app.OrderKeeper, app.SupplyKeeper),
		ammswap.NewAppModule(app.SwapKeeper),
		farm.NewAppModule(app.FarmKeeper),
		params.NewAppModule(app.ParamsKeeper),
		// ibc
		//ibc.NewAppModule(app.IBCKeeper),
		core.NewIBCCOreAppModule(app.IBCKeeper),
		//capabilityModule.NewAppModule(codecProxy, *app.CapabilityKeeper),
		capability.TNewCapabilityModuleAdapter(codecProxy, *app.CapabilityKeeper),
		transferModule,
		erc20.NewAppModule(app.Erc20Keeper),
		mockModule,
		wasm.NewAppModule(*app.marshal, &app.wasmKeeper),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	app.mm.SetOrderBeginBlockers(
		bank.ModuleName,
		capabilitytypes.ModuleName,
		order.ModuleName,
		token.ModuleName,
		dex.ModuleName,
		mint.ModuleName,
		distr.ModuleName,
		slashing.ModuleName,
		staking.ModuleName,
		farm.ModuleName,
		evidence.ModuleName,
		evm.ModuleName,
		ibchost.ModuleName,
		ibctransfertypes.ModuleName,
		mock.ModuleName,
		wasm.ModuleName,
	)
	app.mm.SetOrderEndBlockers(
		crisis.ModuleName,
		gov.ModuleName,
		dex.ModuleName,
		order.ModuleName,
		staking.ModuleName,
		evm.ModuleName,
		mock.ModuleName,
		wasm.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	app.mm.SetOrderInitGenesis(
		capabilitytypes.ModuleName,
		auth.ModuleName, distr.ModuleName, staking.ModuleName, bank.ModuleName,
		slashing.ModuleName, gov.ModuleName, mint.ModuleName, supply.ModuleName,
		token.ModuleName, dex.ModuleName, order.ModuleName, ammswap.ModuleName, farm.ModuleName,
		ibctransfertypes.ModuleName,
		ibchost.ModuleName,
		evm.ModuleName, crisis.ModuleName, genutil.ModuleName, params.ModuleName, evidence.ModuleName,
		erc20.ModuleName,
		mock.ModuleName,
		wasm.ModuleName,
	)

	app.mm.RegisterInvariants(&app.CrisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())
	app.configurator = module.NewConfigurator(app.Codec(), app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)
	app.setupUpgradeModules()

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions
	app.sm = module.NewSimulationManager(
		auth.NewAppModule(app.AccountKeeper),
		bank.NewAppModule(app.BankKeeper, app.AccountKeeper, app.SupplyKeeper),
		supply.NewAppModule(app.SupplyKeeper, app.AccountKeeper),
		gov.NewAppModule(app.GovKeeper, app.SupplyKeeper),
		mint.NewAppModule(app.MintKeeper),
		staking.NewAppModule(app.StakingKeeper, app.AccountKeeper, app.SupplyKeeper),
		distr.NewAppModule(app.DistrKeeper, app.SupplyKeeper),
		slashing.NewAppModule(app.SlashingKeeper, app.AccountKeeper, app.StakingKeeper),
		params.NewAppModule(app.ParamsKeeper), // NOTE: only used for simulation to generate randomized param change proposals
		ibc.NewAppModule(app.IBCKeeper),
		wasm.NewAppModule(*app.marshal, &app.wasmKeeper),
	)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.WasmHandler = wasmkeeper.HandlerOption{
		WasmConfig:        &wasmConfig,
		TXCounterStoreKey: keys[wasm.StoreKey],
	}
	app.SetAnteHandler(ante.NewAnteHandler(app.AccountKeeper, app.EvmKeeper, app.SupplyKeeper, validateMsgHook(app.OrderKeeper), app.WasmHandler, app.IBCKeeper.ChannelKeeper))
	app.SetEndBlocker(app.EndBlocker)
	app.SetGasRefundHandler(refund.NewGasRefundHandler(app.AccountKeeper, app.SupplyKeeper))
	app.SetAccNonceHandler(NewAccHandler(app.AccountKeeper))
	app.SetUpdateFeeCollectorAccHandler(updateFeeCollectorHandler(app.BankKeeper, app.SupplyKeeper))
	app.SetParallelTxLogHandlers(fixLogForParallelTxHandler(app.EvmKeeper))
	app.SetGetTxFeeHandler(getTxFeeHandler())
	app.SetEvmSysContractAddressHandler(NewEvmSysContractAddressHandler(app.EvmKeeper))
	if loadLatest {
		err := app.LoadLatestVersion(app.keys[bam.MainStoreKey])
		if err != nil {
			tmos.Exit(err.Error())
		}
	}

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper

	// NOTE: the IBC mock keeper and application module is used only for testing core IBC. Do
	// note replicate if you do not need to test core IBC or light clients.
	app.ScopedIBCMockKeeper = scopedIBCMockKeeper

	return app
}

func updateFeeCollectorHandler(bk bank.Keeper, sk supply.Keeper) sdk.UpdateFeeCollectorAccHandler {
	return func(ctx sdk.Context, balance sdk.Coins, txFeesplit []*sdk.FeeSplitInfo) error {
		return bk.SetCoins(ctx, sk.GetModuleAccount(ctx, auth.FeeCollectorName).GetAddress(), balance)
	}
}

func fixLogForParallelTxHandler(ek *evm.Keeper) sdk.LogFix {
	return func(tx []sdk.Tx, logIndex []int, hasEnterEvmTx []bool, anteErrs []error, resp []abci.ResponseDeliverTx) (logs [][]byte) {
		return ek.FixLog(tx, logIndex, hasEnterEvmTx, anteErrs, resp)
	}
}

func getTxFeeHandler() sdk.GetTxFeeHandler {
	return func(tx sdk.Tx) (fee sdk.Coins) {
		if feeTx, ok := tx.(authante.FeeTx); ok {
			fee = feeTx.GetFee()
		}

		return
	}
}

func (app *SimApp) SetOption(req abci.RequestSetOption) (res abci.ResponseSetOption) {
	if req.Key == "CheckChainID" {
		if err := evoblockchain.IsValidateChainIdWithGenesisHeight(req.Value); err != nil {
			app.Logger().Error(err.Error())
			panic(err)
		}
		err := evoblockchain.SetChainId(req.Value)
		if err != nil {
			app.Logger().Error(err.Error())
			panic(err)
		}
	}
	return app.BaseApp.SetOption(req)
}

func (app *SimApp) LoadStartVersion(height int64) error {
	return app.LoadVersion(height, app.keys[bam.MainStoreKey])
}

// Name returns the name of the App
func (app *SimApp) Name() string { return app.BaseApp.Name() }

// BeginBlocker updates every begin block
func (app *SimApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker updates every end block
func (app *SimApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	// if appconfig.GetOecConfig().GetEnableDynamicGp() {
	// 	GlobalGpIndex = CalBlockGasPriceIndex(app.blockGasPrice, appconfig.GetOecConfig().GetDynamicGpWeight())
	// 	app.blockGasPrice = app.blockGasPrice[:0]
	// }

	return app.mm.EndBlock(ctx, req)
}

// InitChainer updates at chain initialization
func (app *SimApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {

	var genesisState simapp.GenesisState
	//app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	app.marshal.GetCdc().MustUnmarshalJSON(req.AppStateBytes, &genesisState)
	return app.mm.InitGenesis(ctx, genesisState)
}

// LoadHeight loads state at a particular height
func (app *SimApp) LoadHeight(height int64) error {
	return app.LoadVersion(height, app.keys[bam.MainStoreKey])
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *SimApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[supply.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// SimulationManager implements the SimulationApp interface
func (app *SimApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *SimApp) GetKey(storeKey string) *sdk.KVStoreKey {
	return app.keys[storeKey]
}

func (app *SimApp) GetBaseApp() *bam.BaseApp {
	return app.BaseApp
}

func (app *SimApp) GetStakingKeeper() staking.Keeper {
	return app.StakingKeeper
}
func (app *SimApp) GetIBCKeeper() *ibc.Keeper {
	return app.IBCKeeper
}

func (app *SimApp) GetScopedIBCKeeper() (cap capabilitykeeper.ScopedKeeper) {
	cap = app.ibcScopeKeep
	return
}

func (app *SimApp) AppCodec() *codec.CodecProxy {
	return app.marshal
}

func (app *SimApp) LastCommitID() sdk.CommitID {
	return app.BaseApp.GetCMS().LastCommitID()
}

func (app *SimApp) LastBlockHeight() int64 {
	return app.GetCMS().LastCommitID().Version
}

func (app *SimApp) Codec() *codec.Codec {
	return app.marshal.GetCdc()
}

func (app *SimApp) Marshal() *codec.CodecProxy {
	return app.marshal
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *SimApp) GetSubspace(moduleName string) params.Subspace {
	return app.subspaces[moduleName]
}

var protoCodec = encoding.GetCodec(proto.Name)

func makeInterceptors() map[string]bam.Interceptor {
	m := make(map[string]bam.Interceptor)
	m["/cosmos.tx.v1beta1.Service/Simulate"] = bam.NewRedirectInterceptor("app/simulate")
	m["/cosmos.bank.v1beta1.Query/AllBalances"] = bam.NewRedirectInterceptor("custom/bank/grpc_balances")
	m["/cosmos.staking.v1beta1.Query/Params"] = bam.NewRedirectInterceptor("custom/staking/params4ibc")
	return m
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}

	return dupMaccPerms
}

func validateMsgHook(orderKeeper order.Keeper) ante.ValidateMsgHandler {
	return func(newCtx sdk.Context, msgs []sdk.Msg) error {

		wrongMsgErr := sdk.ErrUnknownRequest(
			"It is not allowed that a transaction with more than one message contains order or evm message")
		var err error

		for _, msg := range msgs {
			switch assertedMsg := msg.(type) {
			case order.MsgNewOrders:
				if len(msgs) > 1 {
					return wrongMsgErr
				}
				_, err = order.ValidateMsgNewOrders(newCtx, orderKeeper, assertedMsg)
			case order.MsgCancelOrders:
				if len(msgs) > 1 {
					return wrongMsgErr
				}
				err = order.ValidateMsgCancelOrders(newCtx, orderKeeper, assertedMsg)
			case *evmtypes.MsgEthereumTx:
				if len(msgs) > 1 {
					return wrongMsgErr
				}
			}

			if err != nil {
				return err
			}
		}
		return nil
	}
}

func NewAccHandler(ak auth.AccountKeeper) sdk.AccNonceHandler {
	return func(
		ctx sdk.Context, addr sdk.AccAddress,
	) uint64 {
		return ak.GetAccount(ctx, addr).GetSequence()
	}
}

func NewEvmSysContractAddressHandler(ak *evm.Keeper) sdk.EvmSysContractAddressHandler {
	if ak == nil {
		panic("NewEvmSysContractAddressHandler ak is nil")
	}
	return func(
		ctx sdk.Context, addr sdk.AccAddress,
	) bool {
		if addr.Empty() {
			return false
		}
		return ak.IsMatchSysContractAddress(ctx, addr)
	}
}

func PreRun(ctx *server.Context) error {
	// set the dynamic config
	appconfig.RegisterDynamicConfig(ctx.Logger.With("module", "config"))

	// check start flag conflicts
	err := sanity.CheckStart()
	if err != nil {
		return err
	}

	// set config by node mode
	//setNodeConfig(ctx)

	//download pprof
	appconfig.PprofDownload(ctx)

	// pruning options
	_, err = server.GetPruningOptionsFromFlags()
	if err != nil {
		return err
	}
	// repair state on start
	// if viper.GetBool(FlagEnableRepairState) {
	// 	repairStateOnStart(ctx)
	// }

	// init tx signature cache
	tmtypes.InitSignatureCache()
	return nil
}

func (app *SimApp) setupUpgradeModules() {
	heightTasks, paramMap, cf, pf, vf := app.CollectUpgradeModules(app.mm)

	app.heightTasks = heightTasks

	app.GetCMS().AppendCommitFilters(cf)
	app.GetCMS().AppendPruneFilters(pf)
	app.GetCMS().AppendVersionFilters(vf)

	vs := app.subspaces
	for k, vv := range paramMap {
		supace, exist := vs[k]
		if !exist {
			continue
		}
		vs[k] = supace.LazyWithKeyTable(subspace.NewKeyTable(vv.ParamSetPairs()...))
	}
}

func (o *SimApp) TxConfig() client.TxConfig {
	return o.txconfig
}

func (o *SimApp) CollectUpgradeModules(m *module.Manager) (map[int64]*upgradetypes.HeightTasks,
	map[string]params.ParamSet, []types.StoreFilter, []types.StoreFilter, []types.VersionFilter) {
	hm := make(map[int64]*upgradetypes.HeightTasks)
	paramsRet := make(map[string]params.ParamSet)
	commitFiltreMap := make(map[*types.StoreFilter]struct{})
	pruneFilterMap := make(map[*types.StoreFilter]struct{})
	versionFilterMap := make(map[*types.VersionFilter]struct{})

	for _, mm := range m.Modules {
		if ada, ok := mm.(upgradetypes.UpgradeModule); ok {
			set := ada.RegisterParam()
			if set != nil {
				if _, exist := paramsRet[ada.ModuleName()]; !exist {
					paramsRet[ada.ModuleName()] = set
				}
			}
			h := ada.UpgradeHeight()
			if h > 0 {
				h++
			}

			cf := ada.CommitFilter()
			if cf != nil {
				if _, exist := commitFiltreMap[cf]; !exist {
					commitFiltreMap[cf] = struct{}{}
				}
			}
			pf := ada.PruneFilter()
			if pf != nil {
				if _, exist := pruneFilterMap[pf]; !exist {
					pruneFilterMap[pf] = struct{}{}
				}
			}
			vf := ada.VersionFilter()
			if vf != nil {
				if _, exist := versionFilterMap[vf]; !exist {
					versionFilterMap[vf] = struct{}{}
				}
			}

			t := ada.RegisterTask()
			if t == nil {
				continue
			}
			if err := t.ValidateBasic(); nil != err {
				panic(err)
			}
			taskList := hm[h]
			if taskList == nil {
				v := make(upgradetypes.HeightTasks, 0)
				taskList = &v
				hm[h] = taskList
			}
			*taskList = append(*taskList, t)
		}
	}

	for _, v := range hm {
		sort.Sort(*v)
	}

	commitFilters := make([]types.StoreFilter, 0)
	pruneFilters := make([]types.StoreFilter, 0)
	versionFilters := make([]types.VersionFilter, 0)
	for pointerFilter, _ := range commitFiltreMap {
		commitFilters = append(commitFilters, *pointerFilter)
	}
	for pointerFilter, _ := range pruneFilterMap {
		pruneFilters = append(pruneFilters, *pointerFilter)
	}
	for pointerFilter, _ := range versionFilterMap {
		versionFilters = append(versionFilters, *pointerFilter)
	}

	return hm, paramsRet, commitFilters, pruneFilters, versionFilters
}
