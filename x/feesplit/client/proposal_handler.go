package client

import (
	"github.com/evoblockchain/evochain/x/feesplit/client/cli"
	"github.com/evoblockchain/evochain/x/feesplit/client/rest"
	govcli "github.com/evoblockchain/evochain/x/gov/client"
)

var (
	// FeeSplitSharesProposalHandler alias gov NewProposalHandler
	FeeSplitSharesProposalHandler = govcli.NewProposalHandler(
		cli.GetCmdFeeSplitSharesProposal,
		rest.FeeSplitSharesProposalRESTHandler,
	)
)
