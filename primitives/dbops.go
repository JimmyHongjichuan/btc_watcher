package primitives

//// GetCurrentHeight 获取上次某条公链监听到的高度
//func GetCurrentHeight(db *dgwdb.LDBDatabase, chainType string) int64 {
//	k := append(keyHeightPrefix, []byte(chainType)...)
//	data, err := db.Get(k)
//	if err != nil {
//		return 0
//	}
//	height, err := util.BytesToI64(data)
//	if err != nil {
//		return 0
//	}
//	return height
//}
//
//// InitDB 初始化leveldb数据
//func InitDB(db *dgwdb.LDBDatabase, genesis *pb.BlockPack) {
//	db.Delete(keyTerm)
//	db.Delete(keyLastTermAccuse)
//	db.Delete(keyWeakAccuses)
//	db.Delete(keyFresh)
//	db.Delete(keyVotie)
//	db.Delete(keyLastBchBlockHeader)
//	db.DeleteWithPrefix(keyCommitChainPrefix)
//	db.DeleteWithPrefix(keyCommitHsByIdsPrefix)
//
//	SetNodeTerm(db, 0)
//	SetVotie(db, genesis.ToVotie())
//	JustCommitIt(db, genesis)
//}
