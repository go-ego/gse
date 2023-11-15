package types

type LoadDictFile struct {
	File     string
	FileType int
}

type LoadBM25DictFile struct {
	FilePath string
	FileType int
	Number   float64
}

type BM25Setting struct {
	K1 float64
	B  float64
}
