package pos

import (
	"testing"

	"github.com/vcaesar/tt"
)

var (
	seg Segmenter
)

func init() {
	seg.LoadDict()
}

func TestCut(t *testing.T) {
	s := seg.Cut("那里湖面总是澄清, 那里空气充满宁静", true)
	tt.Equal(t, "[{那里 r} {湖面 n} {总是 c} {澄清 v} {, x} {  x} {那里 r} {空气 n} {充满 a} {宁静 nr}]", s)

	text := "工信处女干事每月经过下属科室都要亲口交代24口交换机等技术性器件的安装工作"
	s = seg.Cut(text, true)
	tt.Equal(t,
		"[{工信处 n} {女干事 n} {每月 r} {经过 p} {下属 v} {科室 n} {都 d} {要 v} {亲口 n} {交代 n} {24 m} {口 n} {交换 v} {机 ng} {等 u} {技术 n} {性 n} {器件 n} {的 uj} {安装 v} {工作 vn}]", s)

	s = seg.Cut(text, false)
	tt.Equal(t,
		"[{工信处 n} {女干事 n} {每月 r} {经过 p} {下属 v} {科室 n} {都 d} {要 v} {亲口 n} {交代 n} {24 eng} {口 q} {交换 v} {机 n} {等 u} {技术 n} {性 n} {器件 n} {的 uj} {安装 v} {工作 vn}]", s)

	pos := seg.TrimPos(s, "n", "r", "uj")
	tt.Equal(t, 10, len(pos))
	tt.Equal(t, "[{经过 p} {下属 v} {都 d} {要 v} {24 eng} {口 q} {交换 v} {等 u} {安装 v} {工作 vn}]", pos)
}
