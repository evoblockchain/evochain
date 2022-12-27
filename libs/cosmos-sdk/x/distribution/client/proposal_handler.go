package client

import (
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/distribution/client/cli"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/distribution/client/rest"
	govclient "github.com/evoblockchain/evochain/libs/cosmos-sdk/x/gov/client"
)

// param change proposal handler
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
)
