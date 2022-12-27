package client

import (
	"github.com/evoblockchain/evochain/x/dex/client/cli"
	"github.com/evoblockchain/evochain/x/dex/client/rest"
	govclient "github.com/evoblockchain/evochain/x/gov/client"
)

// param change proposal handler
var (
	// DelistProposalHandler alias gov NewProposalHandler
	DelistProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitDelistProposal, rest.DelistProposalRESTHandler)
)
