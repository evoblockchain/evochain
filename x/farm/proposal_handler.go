package farm

import (
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
	"github.com/evoblockchain/evochain/x/common"
	"github.com/evoblockchain/evochain/x/farm/types"
	govTypes "github.com/evoblockchain/evochain/x/gov/types"
)

// NewManageWhiteListProposalHandler handles "gov" type message in "farm"
func NewManageWhiteListProposalHandler(k *Keeper) govTypes.Handler {
	return func(ctx sdk.Context, proposal *govTypes.Proposal) (err sdk.Error) {
		switch content := proposal.Content.(type) {
		case types.ManageWhiteListProposal:
			return handleManageWhiteListProposal(ctx, k, content)
		default:
			return common.ErrUnknownProposalType(DefaultCodespace, content.ProposalType())
		}
	}
}

func handleManageWhiteListProposal(ctx sdk.Context, k *Keeper, p types.ManageWhiteListProposal) sdk.Error {
	if sdkErr := k.CheckMsgManageWhiteListProposal(ctx, p); sdkErr != nil {
		return sdkErr
	}

	if p.IsAdded {
		// add pool name into whitelist
		k.SetWhitelist(ctx, p.PoolName)
		return nil
	}

	// remove pool name from whitelist
	k.DeleteWhiteList(ctx, p.PoolName)
	return nil
}
