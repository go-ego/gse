package pos

import (
	"testing"

	"github.com/vcaesar/tt"
)

var (
	seg Segmenter
)

func TestCut(t *testing.T) {
	seg.LoadDict()

	s := seg.Cut("那里湖面总是澄清, 那里空气充满宁静", true)
	tt.Equal(t, "[{那里 r} {湖面 n} {总是 c} {澄清 v} {, x} {  x} {那里 r} {空气 n} {充满 a} {宁静 nr}]", s)
}
