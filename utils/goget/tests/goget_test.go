package tests

import (
	"fmt"
	"testing"

	"github.com/mnhkahn/cygo/utils/goget"
)

func TestProcess(t *testing.T) {
	schedule := goget.NewGoGetSchedules(2)
	schedule.SetDownloadBlock(1)

	job := schedule.NextJob()
	fmt.Println(job)
	schedule.FinishJob(job)

	fmt.Println(schedule.Percent(), "percent")

	job = schedule.NextJob()
	fmt.Println(job)
	schedule.FinishJob(job)
}
