package consts

const (
	// dict file type to loading

	// LoadDictTypeIDF dict of IDF to loading
	LoadDictTypeIDF = iota + 1

	// LoadDictTypeTFIDF dict of TFIDF to loading
	LoadDictTypeTFIDF

	// LoadDictTypeBM25 dict of BM25 to loading
	LoadDictTypeBM25

	// LoadDictTypeWithPos dict of with position to loading
	LoadDictTypeWithPos

	// LoadDictCorpus dict of corpus to loading
	LoadDictCorpus
)

const (
	// BM25DefaultK1 default k1 value for calculate bm25
	BM25DefaultK1 = 1.25

	// BM25DefaultK1 default B value for calculate bm25
	BM25DefaultB = 0.75
)
