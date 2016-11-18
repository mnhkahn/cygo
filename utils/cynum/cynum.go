package main

import (
	"bytes"
	"flag"
	"fmt"
	"strconv"
)

var (
	t   = flag.String("t", "string", "Type of number. Support string/int.")
	sep = flag.String("sep", ",", "Seperator.")
	s   = flag.Int("s", 0, "Start number.")
	e   = flag.Int("e", 10, "End number.")
)

func main() {
	flag.Parse()

	res := bytes.NewBuffer(nil)

	for i := *s; i <= *e; i++ {
		switch *t {
		case "string":
			res.WriteString(`"`)
			res.WriteString(strconv.Itoa(i))
			res.WriteString(`"`)
		case "int":
			res.WriteString(strconv.Itoa(i))
		}
		if i < *e {
			res.WriteString(*sep)
		}

	}

	fmt.Println(res.String())
}
