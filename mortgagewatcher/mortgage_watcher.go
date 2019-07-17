package mortgagewatcher

import (
	"encoding/json"
	"github.com/JimmyHongjichuan/btc_watcher/coinmanager"
	"github.com/JimmyHongjichuan/btc_watcher/dbop"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	log "github.com/inconshreveable/log15"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
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
//GetUtxoInfoByID 从leveldb中获取utxo信息
func (m *MortgageWatcher) GetUtxoInfoByID(utxoID string) *coinmanager.UtxoInfo {
	t, ok := m.faUtxoInfo.Load(utxoID)
	if ok {
		utxoInfo := t.(*coinmanager.UtxoInfo)
		return utxoInfo
	}

	return nil
}
func (m *MortgageWatcher) storeUtxo(utxoID string) bool {
	t, ok := m.faUtxoInfo.Load(utxoID)
	if ok {
		utxoInfo := t.(*coinmanager.UtxoInfo)
		log.Debug("store utxo", "utxoID", utxoID, "utxo", utxoInfo)
		data, err := json.Marshal(utxoInfo)
		if err != nil {
			log.Warn("Marshal utxo failed", "err", err.Error(), "coinType", m.coinType)
			return false
		}
		key := strings.Join([]string{m.levelDbUtxoPreFix, utxoID}, "_")
		retErr := m.levelDb.Put([]byte(key), []byte(data))
		if retErr != nil {
			log.Warn("Marshal utxo failed", "err", err.Error(), "coinType", m.coinType)
			return false
		}

		return true
	}

	return false
}

//存储tx 签名前与签名后的交易hash映射
func (m *MortgageWatcher) storeHashMapping(tx *wire.MsgTx) bool {
	hashAfterSign := tx.TxHash().String()

	copyTx := tx.Copy()
	for _, vin := range copyTx.TxIn {
		vin.SignatureScript = nil
	}
	hashBeforeSign := copyTx.TxHash().String()
	mappingKey := strings.Join([]string{m.levelDbTxMappingPreFix, hashBeforeSign}, "_")
	log.Debug("storeHashMap", "hash_before_sign", hashBeforeSign, "hash_after_sign", hashAfterSign, "coinType", m.coinType)

	retErr := m.levelDb.Put([]byte(mappingKey), []byte(hashAfterSign))
	if retErr != nil {
		log.Warn("save hashmap failed", "err", retErr.Error(), "coinType", m.coinType)
		return false
	}

	txKey := strings.Join([]string{m.levelDbTxPreFix, hashAfterSign}, "_")
	retErr = m.levelDb.Put([]byte(txKey), []byte(hashAfterSign))
	if retErr != nil {
		log.Warn("save federation hash failed", "err", retErr.Error(), "coinType", m.coinType)
		return false
	}

	return true
}



func (m *MortgageWatcher) processConfirmBlock(blockData *coinmanager.BlockData) {
	for _, tx := range blockData.MsgBolck.Transactions {
		txHash := tx.TxHash().String()

		hasReturn := false
		isFedAddr := false
		isFromFedAddr := false
		var value int64
		var message *Message
		var err error

		//update utxo status

		for _, vin := range tx.TxIn {
			utxoID := strings.Join([]string{vin.PreviousOutPoint.Hash.String(), strconv.Itoa(int(vin.PreviousOutPoint.Index))}, "_")
			utxoInfo := m.GetUtxoInfoByID(utxoID)
			if utxoInfo != nil {
				isFromFedAddr = true
				utxoInfo.SpendType = 3
				m.storeUtxo(utxoID)
				m.faUtxoInfo.Delete(utxoID)

			}

		}

		if isFromFedAddr {
			m.storeHashMapping(tx)
		}

		for voutIndex, vout := range tx.TxOut {
			address := coinmanager.ExtractPkScriptAddr(vout.PkScript, m.coinType)
			if address != "" {
				if _, ok := m.federationMap.Load(address); ok {
					isFedAddr = true
					value = vout.Value
					utxoID := strings.Join([]string{txHash, strconv.Itoa(voutIndex)}, "_")

					t, ok := m.faUtxoInfo.Load(utxoID)
					if !ok {
						newUtxo := coinmanager.UtxoInfo{
							Address:     address,
							Txid:        txHash,
							Vout:        uint32(voutIndex),
							Value:       vout.Value,
							SpendType:   1,
							BlockHeight: blockData.BlockInfo.Height,
						}
						m.faUtxoInfo.Store(utxoID, &newUtxo)
					} else {
						utxoInfo := t.(*coinmanager.UtxoInfo)
						if utxoInfo.SpendType < 1 {
							utxoInfo.SpendType = 1
						}
					}
					m.storeUtxo(utxoID)

					log.Debug("FIND NEW UTXO", "id", utxoID, "value", value, "coinType", m.coinType)

				}
			} else {
				message, err = ParserPayLoadScript(vout.PkScript)
				if err == nil {
					hasReturn = true
				}
			}
		}

		// make mortgage tx
		if isFedAddr && hasReturn {
			mortgageTx := SubTransaction{
				ScTxid:    tx.TxHash().String(),
				Amount:    value,
				From:      m.coinType,
				To:        message.ChainName,
				TokenFrom: 0,
				TokenTo:   message.APPNumber,
				RechargeList: []*AddressInfo{
					{
						Address: message.Address,
						Amount:  value,
					},
				},
			}

			log.Debug("push mortgage tx", "tx", mortgageTx, "coinType", m.coinType)
			m.mortgageTxChan <- &mortgageTx
		}
	}

}

func (m *MortgageWatcher) processNewTx(newTx *wire.MsgTx) {
	txHash := newTx.TxHash().String()
	//update utxo status
	isFromFedAddr := false

	for _, vin := range newTx.TxIn {
		utxoID := strings.Join([]string{vin.PreviousOutPoint.Hash.String(), strconv.Itoa(int(vin.PreviousOutPoint.Index))}, "_")
		utxoInfo := m.GetUtxoInfoByID(utxoID)
		if utxoInfo != nil {
			isFromFedAddr = true
			utxoInfo.SpendType = 2
			m.storeUtxo(utxoID)
		}
	}

	for voutIndex, vout := range newTx.TxOut {
		address := coinmanager.ExtractPkScriptAddr(vout.PkScript, m.coinType)
		if address != "" {
			if _, ok := m.federationMap.Load(address); ok {
				id := strings.Join([]string{txHash, strconv.Itoa(voutIndex)}, "_")
				newUtxo := coinmanager.UtxoInfo{
					Address:   address,
					Txid:      txHash,
					Vout:      uint32(voutIndex),
					Value:     vout.Value,
					SpendType: 0,
				}

				_, ok := m.faUtxoInfo.Load(id)
				if !ok {
					m.faUtxoInfo.Store(id, &newUtxo)
					m.storeUtxo(id)
				}
			}
		}
	}

	if isFromFedAddr {
		log.Info("process tx", "tx_hash", txHash, "coinType", m.coinType)
		m.storeHashMapping(newTx)
	}
}

func (m *MortgageWatcher) processNewUnconfirmBlock(blockData *coinmanager.BlockData) {
	for _, tx := range blockData.MsgBolck.Transactions {
		m.processNewTx(tx)
	}
}
//StartWatch 启动监听已确认和未确认的区块以及新交易，提取抵押交易
func (m *MortgageWatcher) StartWatch() {

	m.utxoMonitor()

	m.bwClient.WatchNewTxFromNodeMempool()
	m.bwClient.WatchNewBlock()

	confirmBlockChan := m.bwClient.GetConfirmChan()
	newTxChan := m.bwClient.GetNewTxChan()
	newUnconfirmBlockChan := m.bwClient.GetNewUnconfirmBlockChan()

	go func() {
		for {
			select {
			case newConfirmBlock := <-confirmBlockChan:
				log.Info("process confirm block height:", "height", newConfirmBlock.BlockInfo.Height, "coinType", m.coinType)
				m.processConfirmBlock(newConfirmBlock)
				if newConfirmBlock.BlockInfo.Height < m.scanConfirmHeight {
					//发生回退
					log.Info("confirm block height roll back", "height", newConfirmBlock.BlockInfo.Height, "coinType", m.coinType)
					//m.checkUtxo(newConfirmBlock.BlockInfo.Height)
				}
				m.scanConfirmHeight = newConfirmBlock.BlockInfo.Height + 1

				height := strconv.Itoa(int(m.scanConfirmHeight))
				err := m.levelDb.Put([]byte(confirmHeightLDKey), []byte(height))
				if err != nil {
					log.Error("Save confirmHeight Failed", "err", err.Error(), "height", height)
				}

			case newTx := <-newTxChan:
				m.processNewTx(newTx)
			case newUnconfirmBlock := <-newUnconfirmBlockChan:
				log.Info("process new block height:", "height", newUnconfirmBlock.BlockInfo.Height, "coinType", m.coinType)
				m.processNewUnconfirmBlock(newUnconfirmBlock)
			}
		}
	}()

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
