package coinmanager

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/wire"
)

//BlockData 区块数据
type BlockData struct {
	BlockInfo *btcjson.GetBlockVerboseResult
	MsgBolck  *wire.MsgBlock
}

//SpendType -1:已移除 0:未确认 1:已确认 2:使用中 3:已使用
type UtxoInfo struct {
	Address       string `json:"address"`
	Txid          string `json:"vout_txid"`
	Vout          uint32 `json:"vout_index"`
	Value         int64  `json:"value"`
	Confirmations int64  `json:"confirmations"`
	SpendType     int    `json:"spend_type"`
	BlockHeight   int64  `json:"block_height"`
}
