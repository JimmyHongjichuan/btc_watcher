package coinmanager

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/cpacia/bchutil"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("net_param", "mainnet")
}

func getNetParams() *chaincfg.Params {
	switch viper.GetString("net_param") {
	case "mainnet":
		return &chaincfg.MainNetParams
	case "testnet":
		return &chaincfg.TestNet3Params
	case "regtest":
		return &chaincfg.RegressionNetParams
	default:
		return nil
	}

}

//DecodeAddress 从地址字符串中decode Address
func DecodeAddress(addr string, coinType string) (btcutil.Address, error) {
	switch coinType {
	case "btc":
		return btcutil.DecodeAddress(addr, getNetParams())
	case "bch":
		return bchutil.DecodeAddress(addr, getNetParams())
	default:
		return nil, nil
	}

}
