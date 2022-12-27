package gasprice

import (
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/params"
	"github.com/spf13/viper"

	appconfig "github.com/evoblockchain/evochain/app/config"
	"github.com/evoblockchain/evochain/app/types"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/server"
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
)

var (
	maxPrice     = big.NewInt(500 * params.GWei)
	defaultPrice = getDefaultGasPrice()
)

type GPOConfig struct {
	Weight  int
	Default *big.Int `toml:",omitempty"`
	Blocks  int
}

func NewGPOConfig(weight int, checkBlocks int) GPOConfig {
	return GPOConfig{
		Weight:  weight,
		Default: defaultPrice,
		Blocks:  checkBlocks,
	}
}

// Oracle recommends gas prices based on the content of recent blocks.
type Oracle struct {
	CurrentBlockGPs *types.SingleBlockGPs
	// hold the gas prices of the latest few blocks
	BlockGPQueue *types.BlockGPResults
	lastPrice    *big.Int
	weight       int
}

// NewOracle returns a new gasprice oracle which can recommend suitable
// gasprice for newly created transaction.
func NewOracle(params GPOConfig) *Oracle {
	cbgp := types.NewSingleBlockGPs()
	bgpq := types.NewBlockGPResults(params.Blocks)
	weight := params.Weight
	if weight < 0 {
		weight = 0
	}
	if weight > 100 {
		weight = 100
	}
	return &Oracle{
		CurrentBlockGPs: cbgp,
		BlockGPQueue:    bgpq,
		lastPrice:       params.Default,
		weight:          weight,
	}
}

func (gpo *Oracle) RecommendGP() *big.Int {

	maxGasUsed := appconfig.GetOecConfig().GetDynamicGpMaxGasUsed()
	maxTxNum := appconfig.GetOecConfig().GetDynamicGpMaxTxNum()
	allTxsLen := int64(len(gpo.CurrentBlockGPs.GetAll()))
	// If the current block's total gas consumption is more than maxGasUsed,
	// or the number of tx in the current block is more than maxTxNum,
	// then we consider the chain to be congested.
	isCongested := (int64(gpo.CurrentBlockGPs.GetGasUsed()) >= maxGasUsed) || (allTxsLen >= maxTxNum)

	// When the network is congested, increase the recommended gas price.
	adoptHigherGp := (appconfig.GetOecConfig().GetDynamicGpMode() == types.CongestionHigherGpMode) && isCongested

	txPrices := gpo.BlockGPQueue.ExecuteSamplingBy(gpo.lastPrice, adoptHigherGp)

	price := new(big.Int).Set(gpo.lastPrice)
	if len(txPrices) > 0 {
		sort.Sort(types.BigIntArray(txPrices))
		price.Set(txPrices[(len(txPrices)-1)*gpo.weight/100])
	}

	if price.Cmp(maxPrice) > 0 {
		price.Set(maxPrice)
	}
	gpo.lastPrice.Set(price)
	return price
}

func getDefaultGasPrice() *big.Int {
	gasPrices, err := sdk.ParseDecCoins(viper.GetString(server.FlagMinGasPrices))
	if err == nil && gasPrices != nil && len(gasPrices) > 0 {
		return gasPrices[0].Amount.BigInt()
	}
	//return the default gas price : DefaultGasPrice
	return sdk.NewDecFromBigIntWithPrec(big.NewInt(1), sdk.Precision/2+1).BigInt()
}
