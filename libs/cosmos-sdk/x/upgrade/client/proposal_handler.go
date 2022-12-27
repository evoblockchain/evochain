package client

import (
	govclient "github.com/evoblockchain/evochain/libs/cosmos-sdk/x/gov/client"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/upgrade/client/cli"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/upgrade/client/rest"
)

var ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitUpgradeProposal, rest.ProposalRESTHandler)
