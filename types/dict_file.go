package types

type LoadDictFile struct {
	File     string
	FileType int
}

type LoadBM25DictFile struct {
	File     string
	FileType int
	Number   float64
}
