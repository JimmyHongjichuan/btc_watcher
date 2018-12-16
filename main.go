package main

import (
	"fmt"

	"github.com/JimmyHongjichuan/btc_watcher/log"
	"github.com/JimmyHongjichuan/btc_watcher/mortgagewatcher"
	"github.com/JimmyHongjichuan/btc_watcher/util"

	"github.com/ofgp/ofgp-core/cluster"

	"github.com/JimmyHongjichuan/btc_watcher/dgwdb"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"

	"path"
)
var (
	nodeLogger        = log.New(viper.GetString("loglevel"), "node")
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

func openDbOrDie(dbPath string) (db *dgwdb.LDBDatabase, newlyCreated bool) {
	if len(dbPath) == 0 {
		homeDir, err := util.GetHomeDir()
		if err != nil {
			panic("Cannot detect the home dir for the current user.")
		}
		dbPath = path.Join(homeDir, "braftdb")
	}

	fmt.Println("open db path ", dbPath)
	info, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		if err := os.Mkdir(dbPath, 0700); err != nil {
			panic(fmt.Errorf("Cannot create db path %v", dbPath))
		}
		newlyCreated = true
	} else {
		if err != nil {
			panic(fmt.Errorf("Cannot get info of %v", dbPath))
		}
		if !info.IsDir() {
			panic(fmt.Errorf("Datavse path (%v) is not a directory", dbPath))
		}
		if c, _ := ioutil.ReadDir(dbPath); len(c) == 0 {
			newlyCreated = true
		} else {
			newlyCreated = false
		}
	}

	db, err = dgwdb.NewLDBDatabase(dbPath, cluster.DbCache, cluster.DbFileHandles)
	if err != nil {
		panic(fmt.Errorf("Failed to open database at %v", dbPath))
	}
	return
}

//func initWatchHeight(db *dgwdb.LDBDatabase) {
//	height := primitives.GetCurrentHeight(db, "bch")
//	if height > 0 {
//		viper.Set("DGW.bch_height", height)
//	}
//	height = primitives.GetCurrentHeight(db, "eth")
//	if height > 0 {
//		viper.Set("DGW.eth_height", height)
//	}
//}

func main() {
	configFile := "./config.toml"
	viper.SetConfigFile(configFile)
	viper.ReadInConfig()

	/*
	db, newlyCreated := openDbOrDie(viper.GetString("DGW.dbpath"))
	if newlyCreated {
		nodeLogger.Debug("initializing new db")
		primitives.InitDB(db, primitives.GenesisBlockPack)
	}
	*/
	//initWatchHeight(db)
	viper.Set("DGW.bch_height", 100);
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
	defer func() { select {} }()
}
