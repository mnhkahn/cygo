package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mnhkahn/cygo/utils/goget"
)

func main() {
	flag.Parse()
	if len(os.Args) < 2 {
		fmt.Println("Error. Use it like 'goget http://7b1h1l.com1.z0.glb.clouddn.com/bryce.jpg'")
	} else {
		config := goget.NewGoGetConfig1(os.Args[len(os.Args)-1], *cntFlag, *proxyType, *proxy)
		goget.DEFAULT_GET.Start(config)
	}
}

var cntFlag = flag.Int("n", 10, "Fetch concurrently counts")
var proxyType = flag.String("t", "", "Proxy Type")
var proxy = flag.String("p", "", "Proxy")
