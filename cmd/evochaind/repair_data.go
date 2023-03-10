package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/evoblockchain/evochain/app"
	"github.com/evoblockchain/evochain/app/utils/appstatus"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/server"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/store/flatkv"
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
	tmiavl "github.com/evoblockchain/evochain/libs/iavl"
	"github.com/evoblockchain/evochain/libs/system/trace"
	sm "github.com/evoblockchain/evochain/libs/tendermint/state"
	tmtypes "github.com/evoblockchain/evochain/libs/tendermint/types"
	types2 "github.com/evoblockchain/evochain/x/evm/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func repairStateCmd(ctx *server.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repair-state",
		Short: "Repair the SMB(state machine broken) data of node",
		PreRun: func(_ *cobra.Command, _ []string) {
			setExternalPackageValue()
		},
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("--------- repair data start ---------")

			go func() {
				pprofAddress := viper.GetString(pprofAddrFlag)
				err := http.ListenAndServe(pprofAddress, nil)
				if err != nil {
					fmt.Println(err)
				}
			}()
			app.RepairState(ctx, false)
			log.Println("--------- repair data success ---------")
		},
	}
	cmd.Flags().Int64(app.FlagStartHeight, 0, "Set the start block height for repair")
	cmd.Flags().Bool(flatkv.FlagEnable, false, "Enable flat kv storage for read performance")
	cmd.Flags().String(app.Elapsed, app.DefaultElapsedSchemas, "schemaName=1|0,,,")
	cmd.Flags().Bool(trace.FlagEnableAnalyzer, false, "Enable auto open log analyzer")
	cmd.Flags().BoolVar(&types2.TrieUseCompositeKey, types2.FlagTrieUseCompositeKey, true, "Use composite key to store contract state")
	cmd.Flags().Int(sm.FlagDeliverTxsExecMode, 0, "execution mode for deliver txs, (0:serial[default], 1:deprecated, 2:parallel)")
	cmd.Flags().String(sdk.FlagDBBackend, tmtypes.DBBackend, "Database backend: goleveldb | rocksdb")
	cmd.Flags().Bool(sdk.FlagMultiCache, true, "Enable multi cache")
	cmd.Flags().StringP(pprofAddrFlag, "p", "0.0.0.0:6060", "Address and port of pprof HTTP server listening")

	return cmd
}

func setExternalPackageValue() {
	tmiavl.SetEnableFastStorage(appstatus.IsFastStorageStrategy())
}
