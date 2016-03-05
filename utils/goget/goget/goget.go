package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mnhkahn/cygo/utils/goget"
)

func main() {
	flag.Parse()
	if len(os.Args) != 2 {
		fmt.Println("Error. Use it like 'goget http://7b1h1l.com1.z0.glb.clouddn.com/bryce.jpg'")
	} else {
		goget.DEFAULT_GET.Start(os.Args[1], *cntFlag)
	}
}

var cntFlag = flag.Int("n", 10, "Fetch concurrently counts")
