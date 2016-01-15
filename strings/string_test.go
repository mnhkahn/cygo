package strings

import (
	"testing"
)

func TestSplitEachAfter(t *testing.T) {
	str := "hello,world,hahaha,yeah"

	var f = func(sub string) {
		t.Log(sub)
	}
	SplitEachAfter(str, ",", f)
	if str != "hello,world,hahaha,yeah" {
		t.Error("str changed", str)
	}

	t.Log("================================")

	str = ""
	SplitEachAfter(str, "", f)

	t.Log("================================")

	str = "你好啊哈哈哈啊"
	SplitEachAfter(str, "啊", f)
}
