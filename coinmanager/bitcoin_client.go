package coinmanager

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	log "github.com/inconshreveable/log15"
	"github.com/spf13/viper"
)

//BitCoinClient BTC/BCH RPC操作类
type BitCoinClient struct {
	rpcClient  *rpcclient.Client
	confirmNum uint64
	coinType   string
}

//GetRawMempool 从全节点内存中获取内存中的交易数据
func (b *BitCoinClient) GetRawMempool() ([]*chainhash.Hash, error) {
	result, err := b.rpcClient.GetRawMempool()
	if err != nil {
		log.Warn("GetRawMempool FAILED:", "err", err.Error())
		return nil, err
	}
	return result, err
}


//NewBitCoinClient 创建一个bitcoin操作客户端
func NewBitCoinClient(coinType string) (*BitCoinClient, error) {
	connCfg := &rpcclient.ConnConfig{
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	bc := &BitCoinClient{}

	bc.coinType = coinType
	switch coinType {
	case "btc":
		connCfg.Host = viper.GetString("BTC.rpc_server")
		connCfg.User = viper.GetString("BTC.rpc_user")
		connCfg.Pass = viper.GetString("BTC.rpc_password")
		bc.confirmNum = uint64(viper.GetInt64("BTC.confirm_block_num"))
	case "bch":
		connCfg.Host = viper.GetString("BCH.rpc_server")
		connCfg.User = viper.GetString("BCH.rpc_user")
		connCfg.Pass = viper.GetString("BCH.rpc_password")
		bc.confirmNum = uint64(viper.GetInt64("BCH.confirm_block_num"))
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Error("GET_BTC_RPC_CLIENT FAIL:", "err", err.Error())
	}

	bc.rpcClient = client

	return bc, err
}

//GetRawTransaction 根据txhash从区块链上查询交易数据
func (b *BitCoinClient) GetRawTransaction(txHash string) (*btcutil.Tx, error) {
	hash, err := chainhash.NewHashFromStr(txHash)
	if err != nil {
		log.Warn("NEW_HASH_FAILED:", "err", err.Error(), "hash", txHash)
		return nil, err
	}

	txRaw, err := b.rpcClient.GetRawTransaction(hash)
	if err != nil {
		log.Warn("GetRawTransaction FAILED:", "err", err.Error(), "hash", txHash)
		return nil, err
	}
	return txRaw, nil
}

//GetBlockCount 获取当前区块链高度
func (b *BitCoinClient) GetBlockCount() int64 {
	blockHeight, err := b.rpcClient.GetBlockCount()
	if err != nil {
		log.Warn("GET_BLOCK_COUNT FAIL:", "err", err.Error())
		return -1
	}
	return blockHeight
}

//GetBlockInfoByHeight 根据区块高度获取区块信息
func (b *BitCoinClient) GetBlockInfoByHeight(height int64) *BlockData {
	blockHash, err := b.rpcClient.GetBlockHash(height)
	if err != nil {
		log.Warn("GET_BLOCK_HASH FAIL:", "err", err.Error())
		return nil
	}

	blockVerbose, err := b.rpcClient.GetBlockVerbose(blockHash)
	if err != nil {
		log.Warn("GET_BLOCK_VERBOSE FAIL:", "err", err.Error())
		return nil
	}

	blockEntity, err := b.rpcClient.GetBlock(blockHash)
	if err != nil {
		log.Warn("GET_BLOCK FAIL:", "err", err.Error())
		return nil
	}

	return &BlockData{
		BlockInfo: blockVerbose,
		MsgBolck:  blockEntity,
	}
}