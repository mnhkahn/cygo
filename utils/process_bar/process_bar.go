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

	"github.com/buger/goterm"
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
	bar.percentDuration = 100 / goterm.Width()
	return bar
}

func (this *ProcessBar) Start(f func() (int, string)) {
	this.ticker = time.NewTicker(this.duration)
	for !this.isComplete {
		select {
		case <-this.ticker.C:
			this.Process(f())
		}
	}
}

func (this *ProcessBar) Process(processCnt int, message string) {
	if processCnt == DEFAULT_PROCESS_100 {
		this.isComplete = true
		if this.ticker != nil {
			this.ticker.Stop()
		}
		fmt.Println("100%", "[", strings.Repeat("=", goterm.Width()-len(message)-10), "]", message)
		return
	} else {
		cnt := processCnt / this.percentDuration
		fmt.Printf("%3d", processCnt)
		fmt.Print("% [ ")

		fmt.Print(strings.Repeat("=", cnt))

		fmt.Print(PROCESS_BAR_GRAPH[processCnt%len(PROCESS_BAR_GRAPH)])

		fmt.Print(strings.Repeat(" ", goterm.Width()-len(message)-cnt-1-10))
		fmt.Print(" ]", message)

		fmt.Print("\r")
	}
}

var defaultProcessBar *ProcessBar

// const DEFAULT_PROCESS_WIDTH = 25
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

func Start(f func() (int, string)) {
	defaultProcessBar.Start(f)
}
