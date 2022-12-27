package client

import (
	"github.com/evoblockchain/evochain/x/farm/client/cli"
	"github.com/evoblockchain/evochain/x/farm/client/rest"
	govcli "github.com/evoblockchain/evochain/x/gov/client"
)

var (
	// ManageWhiteListProposalHandler alias gov NewProposalHandler
	ManageWhiteListProposalHandler = govcli.NewProposalHandler(cli.GetCmdManageWhiteListProposal, rest.ManageWhiteListProposalRESTHandler)
)
