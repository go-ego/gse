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
