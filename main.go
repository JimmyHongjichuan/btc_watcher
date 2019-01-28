package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"log"

	"github.com/JimmyHongjichuan/btc_watcher/primitives"
	//"primitives"
)


func getInputAddress(){
	chainParams :=  &chaincfg.MainNetParams;
	PubKey , err := hex.DecodeString("03fcad68904be57b0a8133293ef4c3f49218df2da8886ed1257bc9f7e7d5a70e28")
	if err != nil {
		panic(err)
	}
	addressPubKey, err := btcutil.NewAddressPubKey(PubKey, chainParams)
	if err != nil {
		log.Fatal(err)
	}
	address := addressPubKey.EncodeAddress()
	fmt.Println (address)
}
func genTx() {
	privWif := "cS5LWK2aUKgP9LmvViG3m9HkfwjaEJpGVbrFHuGZKvW2ae3W9aUe"
	txHash := "12e0d25258ec29fadf75a3f569fccaeeb8ca4af5d2d34e9a48ab5a6fdc0efc1e"
	destination := "mrdKfqWEkwferzEQus5NpgK2Dtpq7Qcgif"
	amount := int64(11650795)
	txFee := int64(500000)
	sourceUTXOIndex := uint32(1)
	chainParams := &chaincfg.TestNet3Params

	decodedWif, err := btcutil.DecodeWIF(privWif)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Decoded WIF: %v\n", decodedWif) // Decoded WIF: cS5LWK2aUKgP9LmvViG3m9HkfwjaEJpGVbrFHuGZKvW2ae3W9aUe

	addressPubKey, err := btcutil.NewAddressPubKey(decodedWif.PrivKey.PubKey().SerializeUncompressed(), chainParams)
	if err != nil {
		log.Fatal(err)
	}

	sourceUTXOHash, err := chainhash.NewHashFromStr(txHash)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("UTXO hash: %s\n", sourceUTXOHash) // utxo hash: 12e0d25258ec29fadf75a3f569fccaeeb8ca4af5d2d34e9a48ab5a6fdc0efc1e

	sourceUTXO := wire.NewOutPoint(sourceUTXOHash, sourceUTXOIndex)
	sourceTxIn := wire.NewTxIn(sourceUTXO, nil, nil)
	destinationAddress, err := btcutil.DecodeAddress(destination, chainParams)
	if err != nil {
		log.Fatal(err)
	}

	sourceAddress, err := btcutil.DecodeAddress(addressPubKey.EncodeAddress(), chainParams)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Source Address: %s\n", sourceAddress) // Source Address: mgjHgKi1g6qLFBM1gQwuMjjVBGMJdrs9pP

	destinationPkScript, err := txscript.PayToAddrScript(destinationAddress)
	if err != nil {
		log.Fatal(err)
	}

	sourcePkScript, err := txscript.PayToAddrScript(sourceAddress)
	if err != nil {
		log.Fatal(err)
	}

	sourceTxOut := wire.NewTxOut(amount, sourcePkScript)

	redeemTx := wire.NewMsgTx(wire.TxVersion)
	redeemTx.AddTxIn(sourceTxIn)
	redeemTxOut := wire.NewTxOut((amount - txFee), destinationPkScript)
	redeemTx.AddTxOut(redeemTxOut)

	sigScript, err := txscript.SignatureScript(redeemTx, 0, sourceTxOut.PkScript, txscript.SigHashAll, decodedWif.PrivKey, false)
	if err != nil {
		log.Fatal(err)
	}

	redeemTx.TxIn[0].SignatureScript = sigScript
	fmt.Printf("Signature Script: %v\n", hex.EncodeToString(sigScript)) // Signature Script: 473...b67

	// validate signature
	flags := txscript.StandardVerifyFlags
	vm, err := txscript.NewEngine(sourceTxOut.PkScript, redeemTx, 0, flags, nil, nil, amount)
	if err != nil {
		log.Fatal(err)
	}

	if err := vm.Execute(); err != nil {
		log.Fatal(err)
	}

	buf := bytes.NewBuffer(make([]byte, 0, redeemTx.SerializeSize()))
	redeemTx.Serialize(buf)

	fmt.Printf("Redeem Tx: %v\n", hex.EncodeToString(buf.Bytes())) // redeem Tx: 01000000011efc...5bb88ac00000000
}

func main() {
	//address, err := primitives.BitCoinHashToAddress("9d3e745da6c1e85960f8a72fc57d0e9c41287d39", primitives.P2SH)
	address, err := primitives.BitCoinHashToAddress("78c0c1bd8bf13af60f4b0371c2f5f9353de777c9", primitives.P2PKH)

	if err != nil {
		panic(err)
	}
	fmt.Println("BITCOIN ADDRESS: ", address)
	getInputAddress()
}
