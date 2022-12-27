package ante

import (
	"github.com/evoblockchain/evochain/app/crypto/ethsecp256k1"
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
	sdkerrors "github.com/evoblockchain/evochain/libs/cosmos-sdk/types/errors"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/auth"
	authante "github.com/evoblockchain/evochain/libs/cosmos-sdk/x/auth/ante"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/auth/types"
	channelkeeper "github.com/evoblockchain/evochain/libs/ibc-go/modules/core/04-channel/keeper"
	ibcante "github.com/evoblockchain/evochain/libs/ibc-go/modules/core/ante"
	"github.com/evoblockchain/evochain/libs/system/trace"
	tmcrypto "github.com/evoblockchain/evochain/libs/tendermint/crypto"
	wasmkeeper "github.com/evoblockchain/evochain/x/wasm/keeper"
)

func init() {
	ethsecp256k1.RegisterCodec(types.ModuleCdc)
}

const (
	// TODO: Use this cost per byte through parameter or overriding NewConsumeGasForTxSizeDecorator
	// which currently defaults at 10, if intended
	// memoCostPerByte     sdk.Gas = 3
	secp256k1VerifyCost uint64 = 21000
)

// NewAnteHandler returns an ante handler responsible for attempting to route an
// Ethereum or SDK transaction to an internal ante handler for performing
// transaction-level processing (e.g. fee payment, signature verification) before
// being passed onto it's respective handler.
func NewAnteHandler(ak auth.AccountKeeper, evmKeeper EVMKeeper, sk types.SupplyKeeper, validateMsgHandler ValidateMsgHandler, option wasmkeeper.HandlerOption, ibcChannelKeepr channelkeeper.Keeper) sdk.AnteHandler {
	var stdTxAnteHandler, evmTxAnteHandler sdk.AnteHandler

	stdTxAnteHandler = sdk.ChainAnteDecorators(
		authante.NewSetUpContextDecorator(),                                             // outermost AnteDecorator. SetUpContext must be called first
		wasmkeeper.NewLimitSimulationGasDecorator(option.WasmConfig.SimulationGasLimit), // after setup context to enforce limits early
		wasmkeeper.NewCountTXDecorator(option.TXCounterStoreKey),
		NewAccountBlockedVerificationDecorator(evmKeeper), //account blocked check AnteDecorator
		authante.NewMempoolFeeDecorator(),
		authante.NewValidateBasicDecorator(),
		authante.NewValidateMemoDecorator(ak),
		authante.NewConsumeGasForTxSizeDecorator(ak),
		authante.NewSetPubKeyDecorator(ak), // SetPubKeyDecorator must be called before all signature verification decorators
		authante.NewValidateSigCountDecorator(ak),
		authante.NewDeductFeeDecorator(ak, sk),
		authante.NewSigGasConsumeDecorator(ak, sigGasConsumer),
		authante.NewSigVerificationDecorator(ak),
		authante.NewIncrementSequenceDecorator(ak), // innermost AnteDecorator
		NewValidateMsgHandlerDecorator(validateMsgHandler),
		ibcante.NewAnteDecorator(ibcChannelKeepr),
	)

	evmTxAnteHandler = sdk.ChainAnteDecorators(
		NewEthSetupContextDecorator(), // outermost AnteDecorator. EthSetUpContext must be called first
		NewGasLimitDecorator(evmKeeper),
		NewEthMempoolFeeDecorator(evmKeeper),
		authante.NewValidateBasicDecorator(),
		NewEthSigVerificationDecorator(),
		NewAccountBlockedVerificationDecorator(evmKeeper), //account blocked check AnteDecorator
		NewAccountAnteDecorator(ak, evmKeeper, sk),
	)

	return func(
		ctx sdk.Context, tx sdk.Tx, sim bool,
	) (newCtx sdk.Context, err error) {
		var anteHandler sdk.AnteHandler
		switch tx.GetType() {
		case sdk.StdTxType:
			anteHandler = stdTxAnteHandler

		case sdk.EvmTxType:
			if ctx.IsWrappedCheckTx() {
				anteHandler = sdk.ChainAnteDecorators(
					NewNonceVerificationDecorator(ak),
					NewIncrementSenderSequenceDecorator(ak),
				)
			} else {
				anteHandler = evmTxAnteHandler
			}

		default:
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", tx)
		}

		return anteHandler(ctx, tx, sim)
	}
}

// sigGasConsumer overrides the DefaultSigVerificationGasConsumer from the x/auth
// module on the SDK. It doesn't allow ed25519 nor multisig thresholds.
func sigGasConsumer(
	meter sdk.GasMeter, _ []byte, pubkey tmcrypto.PubKey, _ types.Params,
) error {
	switch pubkey.(type) {
	case ethsecp256k1.PubKey:
		meter.ConsumeGas(secp256k1VerifyCost, "ante verify: secp256k1")
		return nil
	case tmcrypto.PubKey:
		meter.ConsumeGas(secp256k1VerifyCost, "ante verify: tendermint secp256k1")
		return nil
	default:
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey, "unrecognized public key type: %T", pubkey)
	}
}

func pinAnte(trc *trace.Tracer, tag string) {
	if trc != nil {
		trc.RepeatingPin(tag)
	}
}
