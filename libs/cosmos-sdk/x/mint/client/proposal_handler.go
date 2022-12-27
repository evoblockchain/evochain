package client

import (
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/mint/client/cli"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/mint/client/rest"
	govcli "github.com/evoblockchain/evochain/x/gov/client"
)

var (
	ManageTreasuresProposalHandler = govcli.NewProposalHandler(
		cli.GetCmdManageTreasuresProposal,
		rest.ManageTreasuresProposalRESTHandler,
	)
)
