/*
* Reference: https://www.zhihu.com/question/25306747
             https://www.zhihu.com/question/29738147
             http://stackoverflow.com/questions/11268943/golang-is-it-possible-to-capture-a-ctrlc-signal-and-run-a-cleanup-function-in
*/
package process_bar

import (
	"fmt"
	"strings"
	"time"

	"github.com/buger/goterm"
)

type ProcessBar struct {
	percentDuration float32
	percentWidth    int
	duration        time.Duration
	ticker          *time.Ticker
	isComplete      bool
}

const (
	DEFAULT_PROCESS_100 = 100
	PERCENT_WIDTH       = 4
	SPEED_WIDTH         = 7
	BLANK_WIDTH         = 6 // " [  ] "
)

func NewProcessBar(timer time.Duration) *ProcessBar {
	bar := new(ProcessBar)
	bar.duration = timer
	bar.percentWidth = (goterm.Width() - PERCENT_WIDTH - SPEED_WIDTH - BLANK_WIDTH)
	if bar.percentWidth > DEFAULT_PROCESS_100 {
		bar.percentWidth = DEFAULT_PROCESS_100
		bar.percentDuration = 1.0
	} else {
		bar.percentDuration = float32(DEFAULT_PROCESS_100) / float32(bar.percentWidth)
	}

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
	if processCnt > DEFAULT_PROCESS_100 {
		processCnt = DEFAULT_PROCESS_100
	}
	if processCnt == DEFAULT_PROCESS_100 {
		this.isComplete = true
		if this.ticker != nil {
			this.ticker.Stop()
		}
		fmt.Println("100%", "[", strings.Repeat("=", this.percentWidth), "]", message)
		return
	} else {
		cnt := int(float32(processCnt) / this.percentDuration)
		fmt.Printf("%3d", processCnt)
		fmt.Print("% [ ")

		fmt.Print(strings.Repeat("=", cnt))

		fmt.Print(PROCESS_BAR_GRAPH[processCnt%len(PROCESS_BAR_GRAPH)])

		fmt.Print(strings.Repeat(" ", this.percentWidth-cnt-1))
		// fmt.Print 不会自动追加空格
		fmt.Print(" ] ", message)

		fmt.Print("\r")
	}
}

var defaultProcessBar *ProcessBar

var PROCESS_BAR_GRAPH = []string{" ", "-", "\\", "/"}

// func init() {
// 	defaultProcessBar = NewProcessBar(1 * time.Second)

// 	// handle ^c
// 	c := make(chan os.Signal, 1)
// 	signal.Notify(c, os.Interrupt)
// 	go func() {
// 		for {
// 			select {
// 			case <-c:
// 				fmt.Println()
// 				os.Exit(1)
// 			}
// 		}
// 	}()
// }

func Start(f func() (int, string)) {
	defaultProcessBar.Start(f)
}
