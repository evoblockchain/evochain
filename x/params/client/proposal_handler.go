package client

import (
	govclient "github.com/evoblockchain/evochain/x/gov/client"
	"github.com/evoblockchain/evochain/x/params/client/cli"
	"github.com/evoblockchain/evochain/x/params/client/rest"
)

// ProposalHandler is the param change proposal handler in cmsdk
var ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
