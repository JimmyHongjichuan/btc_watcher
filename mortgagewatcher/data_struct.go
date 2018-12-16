package mortgagewatcher

//AddressInfo 充币地址信息
type AddressInfo struct {
	Address string
	Amount  int64
}

//SubTransaction 铸币/熔币交易信息
type SubTransaction struct {
	ScTxid       string
	Amount       int64
	RechargeList []*AddressInfo //
	From         string         //from chain
	To           string         //to chain
	TokenFrom    uint32
	TokenTo      uint32
}
