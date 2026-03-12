package types

type LoadDictFile struct {
	// file path to load dict
	FilePath string

	// file type to load, defind in file `consts/dict_file.go`
	FileType int
}

type BM25Setting struct {
	// K1 calculation factors for the formula of bm25
	// this factors generally takes values in the range of 1.2 to 2
	// and we define K1 = 1.25 for defult value in `consts/dict_file.go`
	K1 float64

	// B calculation factors for the formula of bm25
	// this factors generally takes values 0.75
	// and we define B = 0.75 for defult value in `consts/dict_file.go`
	B float64
}

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
