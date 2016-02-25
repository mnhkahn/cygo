package process_bar

import (
	"testing"
)

func TestMain(t *testing.T) {
	Start(test)
}

var a = 0

func test() int {
	a += 1
	return a
}
