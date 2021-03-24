// +build go1.16

package gse

import (
	"testing"

	"github.com/vcaesar/tt"
)

func TestLoadDictEmbed(t *testing.T) {
	var seg1 Segmenter
	err := seg1.LoadDictEmbed()
	tt.Nil(t, err)

	f, ok := seg1.Find("1号店")
	tt.Bool(t, ok)
	tt.Equal(t, 3, f)

	f, ok = seg1.Find("八千一百三十七万七千二百三十六口")
	tt.Bool(t, ok)
	tt.Equal(t, 2, f)
}
