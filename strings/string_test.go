package strings

import (
	"fmt"
	gostr "strings"
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

func TestIndexAll(t *testing.T) {
	fmt.Println(gostr.Index("谷歌地图创始人拉斯离开谷歌加盟Facebook", "谷歌"))
	fmt.Println(gostr.LastIndex("谷歌地图创始人拉斯离开谷歌加盟Facebook", "谷歌"))
	fmt.Println(IndexAll("谷歌地图创始人拉斯离开谷歌加盟Facebook", "谷歌"))
}
