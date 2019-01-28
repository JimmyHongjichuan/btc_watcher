package main

import (
	"fmt"

	"github.com/JimmyHongjichuan/btc_watcher/primitives"
	//"primitives"
)

func main() {
	//address, err := primitives.BitCoinHashToAddress("9d3e745da6c1e85960f8a72fc57d0e9c41287d39", primitives.P2SH)
	//address, err := primitives.BitCoinHashToAddress("78c0c1bd8bf13af60f4b0371c2f5f9353de777c9", primitives.P2PKH)
	address, err := primitives.BitCoinHashToAddress("03635e5c184951e14fcfecc83b15960594f4fceec729e09a4a517b0a03a7f4b9", primitives.P2PKH)
	if err != nil {
		panic(err)
	}
	fmt.Println("BITCOIN ADDRESS: ", address)
}
