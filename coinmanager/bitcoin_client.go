package coinmanager

import (
	"github.com/btcsuite/btcd/rpcclient"
	log "github.com/inconshreveable/log15"
	"github.com/spf13/viper"
)

//BitCoinClient BTC/BCH RPC操作类
type BitCoinClient struct {
	rpcClient  *rpcclient.Client
	confirmNum uint64
	coinType   string
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
