package coinmanager

import (
	"github.com/btcsuite/btcd/wire"
	log "github.com/inconshreveable/log15"
	"github.com/spf13/viper"
)

//BitCoinWatcher BTC/BCH监听类
type BitCoinWatcher struct {
	scanConfirmHeight     int64
	watchHeight           int64
	coinType              string
	bitcoinClient         *BitCoinClient
	confirmBlockChan      chan *BlockData
	newUnconfirmBlockChan chan *BlockData
	newTxChan             chan *wire.MsgTx
	confirmNeedNum        int64
	mempoolTxs            map[string]int
	//zmqClient             *zmq.Socket
	freshBlockList []*BlockData
}

//NewBitCoinWatcher 创建一个BTC/BCH监听实例
func NewBitCoinWatcher(coinType string, confirmHeight int64) (*BitCoinWatcher, error) {
	bw := BitCoinWatcher{
		scanConfirmHeight:     confirmHeight,
		coinType:              coinType,
		confirmBlockChan:      make(chan *BlockData, 32),
		newUnconfirmBlockChan: make(chan *BlockData, 32),
		newTxChan:             make(chan *wire.MsgTx, 100),
		mempoolTxs:            make(map[string]int),
		watchHeight:           -1,
		freshBlockList:        nil,
	}

	//bw.zmqClient, _ = zmq.NewSocket(zmq.SUB)
	switch coinType {
	case "btc":
		bw.confirmNeedNum = viper.GetInt64("BTC.confirm_block_num")
		//bw.zmqClient.Connect(viper.GetString("BTC.zmq_server"))
	case "bch":
		bw.confirmNeedNum = viper.GetInt64("BCH.confirm_block_num")
		//bw.zmqClient.Connect(viper.GetString("BCH.zmq_server"))

	}
	//bw.zmqClient.SetSubscribe("hashblock")
	bitcoinClient, err := NewBitCoinClient(coinType)
	if err != nil {
		log.Error("Create btc Client failed:", "err", err.Error())
		return &bw, err
	}
	bw.bitcoinClient = bitcoinClient

	return &bw, nil
}
