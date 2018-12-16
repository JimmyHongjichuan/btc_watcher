package main

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/yaanhyy/btc_watcher/mortgagewatcher"
	"github.com/yaanhyy/btc_watcher/util"
	"path"
)

func init() {
	homeDir, _ := util.GetHomeDir()
	dbPath := path.Join(homeDir, "btc_db")
	viper.SetDefault("LEVELDB.btc_db_path", dbPath)
	dbPath = path.Join(homeDir, "bch_db")
	viper.SetDefault("LEVELDB.bch_db_path", dbPath)
	viper.SetDefault("BTC.load_mode", "leveldb")
	viper.SetDefault("BCH.load_mode", "leveldb")

}

// MultiSigInfo 多签信息
type MultiSigInfo struct {
	BtcAddress      string `json:"btc_address"`
	BtcRedeemScript []byte `json:"btc_redeem_script"`
	BchAddress      string `json:"bch_address"`
	BchRedeemScript []byte `json:"bch_redeem_script"`
}

var defaultUtxoLockTime = 60

func main() {
	configFile := "./config.toml"
	viper.SetConfigFile(configFile)
	viper.ReadInConfig()
	// bchFederationAddress, bchRedeem, btcFederationAddress, btcRedeem := getFederationAddress()
	multiSig := new(MultiSigInfo)
	multiSig.BtcAddress = viper.GetString("BTC.btc_multisig")
	multiSig.BtcRedeemScript = []byte(viper.GetString("BTC.btc_redeem_script"))
	fmt.Sprintf("get multisig address", "btc", multiSig.BtcAddress, "bch", multiSig.BchAddress)

	utxoLockTime := viper.GetInt("DGW.utxo_lock_time")
	if utxoLockTime == 0 {
		utxoLockTime = defaultUtxoLockTime
	}

	var (
		btcWatcher *mortgagewatcher.MortgageWatcher
		//bchWatcher *btcwatcher.MortgageWatcher
		//ethWatcher *ew.Client
		//xinWatcher *eoswatcher.EOSWatcher
		//eosWatcher *eoswatcher.EOSWatcherMain
		err error
	)

	if len(viper.GetString("BTC.rpc_server")) > 0 {
		btcWatcher, err = mortgagewatcher.NewMortgageWatcher("btc", viper.GetInt64("DGW.btc_height"),
			multiSig.BtcAddress, multiSig.BtcRedeemScript, utxoLockTime)
		if err != nil {
			panic(fmt.Sprintf("new btc watcher failed, err: %v", err))
		}
	}
	btcWatcher.StartWatch()
}
