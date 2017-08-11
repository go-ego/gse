package gse

import (
	"testing"
)

var (
	prodSeg = Segmenter{}
)

func TestSplit(t *testing.T) {
	expect(t, "中/国/有/十/三/亿/人/口/",
		bytesToString(splitTextToWords([]byte(
			"中国有十三亿人口"))))

	expect(t, "github/ /is/ /a/ /web/-/based/ /hosting/ /service/,/ /for/ /software/ /development/ /projects/./",
		bytesToString(splitTextToWords([]byte(
			"GitHub is a web-based hosting service, for software development projects."))))

	expect(t, "中/国/雅/虎/yahoo/!/ /china/致/力/于/，/领/先/的/公/益/民/生/门/户/网/站/。/",
		bytesToString(splitTextToWords([]byte(
			"中国雅虎Yahoo! China致力于，领先的公益民生门户网站。"))))

	expect(t, "こ/ん/に/ち/は/", bytesToString(splitTextToWords([]byte("こんにちは"))))

	expect(t, "안/녕/하/세/요/", bytesToString(splitTextToWords([]byte("안녕하세요"))))

	expect(t, "Я/ /тоже/ /рада/ /Вас/ /видеть/", bytesToString(splitTextToWords([]byte("Я тоже рада Вас видеть"))))

	expect(t, "¿/cómo/ /van/ /las/ /cosas/", bytesToString(splitTextToWords([]byte("¿Cómo van las cosas"))))

	expect(t, "wie/ /geht/ /es/ /ihnen/", bytesToString(splitTextToWords([]byte("Wie geht es Ihnen"))))

	expect(t, "je/ /suis/ /enchanté/ /de/ /cette/ /pièce/",
		bytesToString(splitTextToWords([]byte("Je suis enchanté de cette pièce"))))
}

func TestSegment(t *testing.T) {
	var seg Segmenter
	seg.LoadDict("testdata/test_dict1.txt,testdata/test_dict2.txt")
	expect(t, "12", seg.dict.NumTokens())
	segments := seg.Segment([]byte("中国有十三亿人口"))
	expect(t, "中国/ 有/p3 十三亿/ 人口/p12 ", ToString(segments, false))
	expect(t, "4", len(segments))
	expect(t, "0", segments[0].start)
	expect(t, "6", segments[0].end)
	expect(t, "6", segments[1].start)
	expect(t, "9", segments[1].end)
	expect(t, "9", segments[2].start)
	expect(t, "18", segments[2].end)
	expect(t, "18", segments[3].start)
	expect(t, "24", segments[3].end)
}

func TestLargeDictionary(t *testing.T) {
	prodSeg.LoadDict("data/dictionary.txt")
	expect(t, "中国/ns 人口/n ", ToString(prodSeg.Segment(
		[]byte("中国人口")), false))

	expect(t, "中国/ns 人口/n ", ToString(prodSeg.internalSegment(
		[]byte("中国人口"), false), false))

	expect(t, "中国/ns 人口/n ", ToString(prodSeg.internalSegment(
		[]byte("中国人口"), true), false))

	expect(t, "中华人民共和国/ns 中央人民政府/nt ", ToString(prodSeg.internalSegment(
		[]byte("中华人民共和国中央人民政府"), true), false))

	expect(t, "中华人民共和国中央人民政府/nt ", ToString(prodSeg.internalSegment(
		[]byte("中华人民共和国中央人民政府"), false), false))

	expect(t, "中华/nz 人民/n 共和/nz 共和国/ns 人民共和国/nt 中华人民共和国/ns 中央/n 人民/n 政府/n 人民政府/nt 中央人民政府/nt 中华人民共和国中央人民政府/nt ", ToString(prodSeg.Segment(
		[]byte("中华人民共和国中央人民政府")), true))
}
