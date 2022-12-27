package params

import (
	"fmt"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/codec"
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
	sdkerrors "github.com/evoblockchain/evochain/libs/cosmos-sdk/types/errors"
	abci "github.com/evoblockchain/evochain/libs/tendermint/abci/types"

	"github.com/evoblockchain/evochain/x/params/types"
)

// NewQuerier returns all query handlers
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case types.QueryParams:
			return queryParams(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown params query endpoint")
		}
	}
}

func queryParams(ctx sdk.Context, _ abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	bz, err := codec.MarshalJSONIndent(keeper.cdc, keeper.GetParams(ctx))
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, fmt.Sprintf("could not marshal result to JSON %s", err.Error()))
	}
	return bz, nil
}
