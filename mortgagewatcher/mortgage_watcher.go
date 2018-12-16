package mortgagewatcher

import (
	"encoding/json"
	"github.com/btcsuite/btcutil"
	log "github.com/inconshreveable/log15"
	"github.com/yaanhyy/btc_watcher/coinmanager"
	"os"
	"strings"
	"time"

	"github.com/yaanhyy/btc_watcher/dbop"
	"strconv"
	"sync"

	"github.com/spf13/viper"
)

var (
	confirmHeightLDKey = "confirmHeight"
)

//MortgageWatcher 抵押交易监听类
type MortgageWatcher struct {
	sync.Mutex
	bwClient          *coinmanager.BitCoinWatcher
	scanConfirmHeight int64
	coinType          string
	mortgageTxChan    chan *SubTransaction
	federationAddress string
	redeemScript      []byte
	levelDb           *dbop.LDBDatabase
	utxoMonitorCount  sync.Map
	federationMap     sync.Map
	faUtxoInfo        sync.Map
	addrList          []btcutil.Address
	timeout           int
	confirmNum        int64
	firstBlockHeight  int64
	loadMode          string

	levelDbTxMappingPreFix string
	levelDbTxPreFix        string
	levelDbUtxoPreFix      string
}

func openLevelDB(coinType string) (*dbop.LDBDatabase, error) {
	dbPath := viper.GetString("LEVELDB.bch_db_path")
	if coinType == "btc" {
		dbPath = viper.GetString("LEVELDB.btc_db_path")
	}

	info, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		if err := os.Mkdir(dbPath, 0700); err != nil {
			return nil, err
		}
	} else {
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			return nil, err
		}
	}

	db, err := dbop.NewLDBDatabase(dbPath, 16, 16)
	if err != nil {
		return nil, err
	}
	return db, nil

}

//NewMortgageWatcher 创建一个抵押交易监听实例
//coinType bch/btc, confirmHeight监听开始的高度， federationAddress 要监听的多签地址 redeemScript兑现脚本
func NewMortgageWatcher(coinType string, confirmHeight int64, federationAddress string, redeemScript []byte, timeout int) (*MortgageWatcher, error) {

	levelDb, err := openLevelDB(coinType)
	if err != nil {
		log.Error("open level db failed", "err", err.Error())
		return nil, err
	}

	//查看leveldb存储的高度
	height := 0

	value, err := levelDb.Get([]byte(confirmHeightLDKey))
	if value != nil && err == nil {
		height, err = strconv.Atoi(string(value))
		if err != nil {
			log.Error("atoi height failed", "err", err.Error())
			return nil, err
		}
	}

	log.Debug("confirm height", "height", height)

	if int64(height) >= confirmHeight {
		confirmHeight = int64(height)
	}

	bwClient, err := coinmanager.NewBitCoinWatcher(coinType, confirmHeight)
	if err != nil {
		return nil, err
	}

	mw := MortgageWatcher{
		levelDb:           levelDb,
		bwClient:          bwClient,
		scanConfirmHeight: confirmHeight,
		coinType:          coinType,
		mortgageTxChan:    make(chan *SubTransaction, 100),
		federationAddress: federationAddress,
		redeemScript:      redeemScript,
		timeout:           timeout,
	}

	switch coinType {
	case "btc":
		mw.confirmNum = int64(viper.GetInt64("BTC.confirm_block_num"))
		mw.firstBlockHeight = int64(viper.GetInt64("BTC.first_block_height"))
		mw.loadMode = viper.GetString("BTC.load_mode")
	case "bch":
		mw.confirmNum = int64(viper.GetInt64("BCH.confirm_block_num"))
		mw.firstBlockHeight = int64(viper.GetInt64("BCH.first_block_height"))
		mw.loadMode = viper.GetString("BCH.load_mode")
	}

	mw.levelDbUtxoPreFix = strings.Join([]string{coinType, "utxo"}, "_")
	mw.levelDbTxPreFix = strings.Join([]string{coinType, "fa_tx"}, "_")
	mw.levelDbTxMappingPreFix = strings.Join([]string{coinType, "hash_mapping"}, "_")

	mw.federationMap.Store(federationAddress, redeemScript)
	addr, err := coinmanager.DecodeAddress(federationAddress, coinType)
	if err != nil {
		log.Warn("decode address failed", "err", err.Error())
		return nil, err
	}
	mw.addrList = append(mw.addrList, addr)

	mw.loadUtxo()

	return &mw, err
}

func (m *MortgageWatcher) utxoMonitor() {
	go func() {
		for {
			m.utxoMonitorCount.Range(func(k, v interface{}) bool {
				utxoID := k.(string)
				count := v.(int)
				count++
				if count >= m.timeout {
					m.utxoMonitorCount.Delete(utxoID)
				} else {
					m.utxoMonitorCount.Store(utxoID, count)
				}
				return true
			})

			time.Sleep(time.Duration(1) * time.Second)
		}
	}()
}

//StartWatch 启动监听已确认和未确认的区块以及新交易，提取抵押交易
func (m *MortgageWatcher) StartWatch() {
	m.utxoMonitor()
	//
	//	m.bwClient.WatchNewTxFromNodeMempool()
	//	m.bwClient.WatchNewBlock()
	//
	//	confirmBlockChan := m.bwClient.GetConfirmChan()
	//	newTxChan := m.bwClient.GetNewTxChan()
	//	newUnconfirmBlockChan := m.bwClient.GetNewUnconfirmBlockChan()
	//
	//	go func() {
	//		for {
	//			select {
	//			case newConfirmBlock := <-confirmBlockChan:
	//				log.Info("process confirm block height:", "height", newConfirmBlock.BlockInfo.Height, "coinType", m.coinType)
	//				m.processConfirmBlock(newConfirmBlock)
	//				if newConfirmBlock.BlockInfo.Height < m.scanConfirmHeight {
	//					//发生回退
	//					log.Info("confirm block height roll back", "height", newConfirmBlock.BlockInfo.Height, "coinType", m.coinType)
	//					m.checkUtxo(newConfirmBlock.BlockInfo.Height)
	//				}
	//				m.scanConfirmHeight = newConfirmBlock.BlockInfo.Height + 1
	//
	//				height := strconv.Itoa(int(m.scanConfirmHeight))
	//				err := m.levelDb.Put([]byte(confirmHeightLDKey), []byte(height))
	//				if err != nil {
	//					log.Error("Save confirmHeight Failed", "err", err.Error(), "height", height)
	//				}
	//
	//			case newTx := <-newTxChan:
	//				m.processNewTx(newTx)
	//			case newUnconfirmBlock := <-newUnconfirmBlockChan:
	//				log.Info("process new block height:", "height", newUnconfirmBlock.BlockInfo.Height, "coinType", m.coinType)
	//				m.processNewUnconfirmBlock(newUnconfirmBlock)
	//			}
	//		}
	//	}()
	//
}

//loadUtxoFromLevelDb 从leveldb中，load utxo进内存
func (m *MortgageWatcher) loadUtxoFromLevelDb() {
	iter := m.levelDb.NewIteratorWithPrefix([]byte(m.levelDbUtxoPreFix))
	for iter.Next() {
		utxo := &coinmanager.UtxoInfo{}
		err := json.Unmarshal(iter.Value(), utxo)
		log.Debug("load utxo from leveldb", "key", string(iter.Key()), "utxo", utxo)
		if err != nil {
			log.Warn("Unmarshal UTXO FROM LEVELDB ERR", "err", err.Error(), "coinType", m.coinType)
			continue
		}

		if utxo.SpendType == 3 {
			continue
		}

		m.faUtxoInfo.Store(string(iter.Key())[9:], utxo)

	}
}

func (m *MortgageWatcher) loadUtxo() {
	switch m.loadMode {
	case "leveldb":
		m.loadUtxoFromLevelDb()
	case "chain":
		//m.loadUtxoFromChain()
	}
}
