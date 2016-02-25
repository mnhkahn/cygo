/*
* Reference: https://www.zhihu.com/question/25306747
             https://www.zhihu.com/question/29738147
             http://stackoverflow.com/questions/11268943/golang-is-it-possible-to-capture-a-ctrlc-signal-and-run-a-cleanup-function-in
*/
package process_bar

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"
)

type ProcessBar struct {
	percentDuration int
	duration        time.Duration
	ticker          *time.Ticker
	isComplete      bool
}

func NewProcessBar(timer time.Duration) *ProcessBar {
	bar := new(ProcessBar)
	bar.duration = timer
	bar.percentDuration = 100 / DEFAULT_PROCESS_WIDTH
	return bar
}

func (this *ProcessBar) Start(f func() int) {
	this.ticker = time.NewTicker(this.duration)
	for !this.isComplete {
		select {
		case <-this.ticker.C:
			this.process(f())
		}
	}
}

func (this *ProcessBar) process(processCnt int) {
	if processCnt == DEFAULT_PROCESS_100 {
		this.isComplete = true
		this.ticker.Stop()
		fmt.Println("100%", "[", strings.Repeat("=", DEFAULT_PROCESS_WIDTH), "]")
		return
	} else {
		cnt := processCnt / this.percentDuration
		fmt.Printf("%3d", processCnt)
		fmt.Print("% [ ")

		fmt.Print(strings.Repeat("=", cnt))

		fmt.Print(PROCESS_BAR_GRAPH[processCnt%len(PROCESS_BAR_GRAPH)])

		fmt.Print(strings.Repeat(" ", DEFAULT_PROCESS_WIDTH-cnt-1))
		fmt.Print(" ]")

		fmt.Print("\r")
	}
}

var defaultProcessBar *ProcessBar

const DEFAULT_PROCESS_WIDTH = 25
const DEFAULT_PROCESS_100 = 100

var PROCESS_BAR_GRAPH = []string{" ", "-", "\\", "/"}

func init() {
	defaultProcessBar = NewProcessBar(1 * time.Second)

	// handle ^c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for {
			select {
			case <-c:
				fmt.Println()
				os.Exit(1)
			}
		}
	}()
}

func Start(f func() int) {
	defaultProcessBar.Start(f)
}
