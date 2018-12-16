package coinmanager

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/cpacia/bchutil"
	"github.com/spf13/viper"
	log "github.com/inconshreveable/log15"
	"reflect"
	"strings"
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

//ExtractPkScriptAddr 从输出脚本中提取地址
func ExtractPkScriptAddr(PkScript []byte, coinType string) string {
	scriptClass, addresses, _, err := txscript.ExtractPkScriptAddrs(
		PkScript, getNetParams())

	if err != nil {
		log.Warn("ExtractPkScriptAddrs:", "err", err.Error())
		return ""
	}

	if scriptClass == txscript.NullDataTy {
		return ""
	}

	if len(addresses) > 0 {
		//address := addresses[0].EncodeAddress()

		var addressList []string
		for _, addr := range addresses {
			//log.Debug("addr type", "addr type", reflect.TypeOf(addr).String())
			if coinType == "btc" {
				addressList = append(addressList, addr.EncodeAddress())
			} else {
				switch reflect.TypeOf(addr).String() {
				case "*btcutil.AddressPubKey":
					addr2, err := bchutil.NewCashAddressPubKeyHash(btcutil.Hash160(addr.ScriptAddress()), getNetParams())
					if err != nil {
						log.Warn("NewCashAddressPubKeyHash failed:", "err", err.Error())
						return ""
					}
					addressList = append(addressList, addr2.String())
				case "*btcutil.AddressPubKeyHash":
					addr2, err := bchutil.NewCashAddressPubKeyHash(addr.ScriptAddress(), getNetParams())
					if err != nil {
						log.Warn("NewCashAddressPubKeyHash failed:", "err", err.Error())
						return ""
					}
					addressList = append(addressList, addr2.String())
				case "*btcutil.AddressScriptHash":
					addr2, err := bchutil.NewCashAddressScriptHashFromHash(addr.ScriptAddress(), getNetParams())
					if err != nil {
						log.Warn("NewCashAddressScriptHashFromHash failed:", "err", err.Error())
						return ""
					}
					addressList = append(addressList, addr2.String())
				}
			}

		}
		return strings.Join(addressList, "|")

		/*
			switch scriptClass {
			case txscript.PubKeyTy:
				addr2, err := bchutil.NewCashAddressPubKeyHash(btcutil.Hash160(addresses[0].ScriptAddress()), getNetParams())
				if err != nil {
					log.Warn("NewCashAddressPubKeyHash failed:", "err", err.Error())
					return ""
				}
				return addr2.String()
			case txscript.PubKeyHashTy:
				addr2, err := bchutil.NewCashAddressPubKeyHash(addresses[0].ScriptAddress(), getNetParams())
				if err != nil {
					log.Warn("NewCashAddressPubKeyHash failed:", "err", err.Error())
					return ""
				}
				return addr2.String()
			case txscript.ScriptHashTy:
				addr2, err := bchutil.NewCashAddressScriptHashFromHash(addresses[0].ScriptAddress(), getNetParams())
				if err != nil {
					log.Warn("NewCashAddressScriptHashFromHash failed:", "err", err.Error())
					return ""
				}
				return addr2.String()
			}*/
	}

	return ""
}

