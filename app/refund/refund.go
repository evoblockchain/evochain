package refund

import (
	"math/big"
	"sync"

	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/auth/ante"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/auth/keeper"

	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
	sdkerrors "github.com/evoblockchain/evochain/libs/cosmos-sdk/types/errors"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/auth"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/auth/types"
)

func NewGasRefundHandler(ak auth.AccountKeeper, sk types.SupplyKeeper) sdk.GasRefundHandler {
	evmGasRefundHandler := NewGasRefundDecorator(ak, sk)

	return func(
		ctx sdk.Context, tx sdk.Tx,
	) (refundFee sdk.Coins, err error) {
		var gasRefundHandler sdk.GasRefundHandler

		if tx.GetType() == sdk.EvmTxType {
			gasRefundHandler = evmGasRefundHandler
		} else {
			return nil, nil
		}
		return gasRefundHandler(ctx, tx)
	}
}

type Handler struct {
	ak           keeper.AccountKeeper
	supplyKeeper types.SupplyKeeper
}

func (handler Handler) GasRefund(ctx sdk.Context, tx sdk.Tx) (refundGasFee sdk.Coins, err error) {
	currentGasMeter := ctx.GasMeter()
	ctx.SetGasMeter(sdk.NewInfiniteGasMeter())

	gasLimit := currentGasMeter.Limit()
	gasUsed := currentGasMeter.GasConsumed()

	if gasUsed >= gasLimit {
		return nil, nil
	}

	feeTx, ok := tx.(ante.FeeTx)
	if !ok {
		return nil, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feePayer := feeTx.FeePayer(ctx)
	feePayerAcc := handler.ak.GetAccount(ctx, feePayer)
	if feePayerAcc == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %s does not exist", feePayer)
	}

	gas := feeTx.GetGas()
	fees := feeTx.GetFee()
	gasFees := calculateRefundFees(gasUsed, gas, fees)

	newCoins := feePayerAcc.GetCoins().Add(gasFees...)
	if err = feePayerAcc.SetCoins(newCoins); err != nil {
		return nil, err
	}
	handler.ak.SetAccount(ctx, feePayerAcc)

	return gasFees, nil
}

func NewGasRefundDecorator(ak auth.AccountKeeper, sk types.SupplyKeeper) sdk.GasRefundHandler {
	chandler := Handler{
		ak:           ak,
		supplyKeeper: sk,
	}
	return chandler.GasRefund
}

var bigIntsPool = &sync.Pool{
	New: func() interface{} {
		return &[2]big.Int{}
	},
}

func calculateRefundFees(gasUsed uint64, gas uint64, fees sdk.DecCoins) sdk.Coins {
	bitInts := bigIntsPool.Get().(*[2]big.Int)
	defer bigIntsPool.Put(bitInts)

	refundFees := make(sdk.Coins, len(fees))
	for i, fee := range fees {
		gasPrice := bitInts[0].SetUint64(gas)
		gasPrice = gasPrice.Div(fee.Amount.Int, gasPrice)

		gasConsumed := bitInts[1].SetUint64(gasUsed)
		gasConsumed = gasConsumed.Mul(gasPrice, gasConsumed)

		gasCost := sdk.NewDecCoinFromDec(fee.Denom, sdk.NewDecWithBigIntAndPrec(gasConsumed, sdk.Precision))
		gasRefund := fee.Sub(gasCost)

		refundFees[i] = gasRefund
	}
	return refundFees
}

// CalculateRefundFees provides the way to calculate the refunded gas with gasUsed, fees and gasPrice,
// as refunded gas = fees - gasPrice * gasUsed
func CalculateRefundFees(gasUsed uint64, fees sdk.DecCoins, gasPrice *big.Int) sdk.Coins {
	gas := new(big.Int).Div(fees[0].Amount.BigInt(), gasPrice).Uint64()
	return calculateRefundFees(gasUsed, gas, fees)
}
