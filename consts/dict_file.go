package consts

const (
	// dict file type to loading

	// LoadDictTypeIDF dict of TFIDF to loading
	LoadDictTypeIDF = iota + 1

	// LoadDictTypeTFIDF dict of IDF to loading
	LoadDictTypeTFIDF

	// LoadDictTypeBM25 dict of BM25 to loading
	LoadDictTypeBM25

	// LoadDictTypeWithPos dict of with position to loading
	LoadDictTypeWithPos
)

const (
	BM25DefaultK1 = 1.25
	BM25DefaultB  = 0.8
)
