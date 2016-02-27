package goget

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mnhkahn/cygo/utils/process_bar"
)

const (
	DEFAULT_DOWNLOAD_BLOCK int64 = 1048576 // 2^20
)

var (
	DebugLog *log.Logger
)

type GoGet struct {
	Url         string
	Cnt         int
	Schedule    *GoGetSchedules
	Latch       int
	Header      http.Header
	MediaType   string
	MediaParams map[string]string
	FilePath    string // 包括路径和文件名
	GetClient   *http.Client
	File        *os.File
	TempFiles   []*os.File
	raw         []byte
	jobs        chan *GoGetBlock
	jobStatus   chan *GoGetBlock
	processBar  *process_bar.ProcessBar
}

// 前开后闭区间？？？
type GoGetBlock struct {
	Start int64
	End   int64
}

const (
	STATUS_NO_START = byte(0)
	STATUS_START    = byte(1)
	STATUS_FINISH   = byte(2)
)

type GoGetSchedules struct {
	processes      []byte
	DownloadBlock  int64
	ContentLength  int64
	CompleteLength int64
	startTime      time.Time
	lock           sync.RWMutex
}

func NewGoGetSchedules(contentLength int64) *GoGetSchedules {
	schedules := new(GoGetSchedules)
	schedules.DownloadBlock = DEFAULT_DOWNLOAD_BLOCK
	schedules.ContentLength = contentLength
	schedules.processes = make([]byte, schedules.ContentLength, schedules.ContentLength)
	return schedules
}

func (this *GoGetSchedules) SetDownloadBlock(block int64) {
	this.DownloadBlock = block
}

func (this *GoGetSchedules) Percent() float32 {
	return float32(this.CompleteLength) / float32(this.ContentLength)
}

func (this *GoGetSchedules) Speed() string {
	elaspe := time.Now().Sub(this.startTime).Seconds()
	return fmt.Sprintf(" %d KB/S     ", this.CompleteLength/(int64(elaspe*1000)))
}

func (this *GoGetSchedules) NextJob() *GoGetBlock {
	this.lock.Lock()
	defer this.lock.Unlock()

	if this.startTime.IsZero() {
		this.startTime = time.Now()
	}

	job := new(GoGetBlock)

	var i int64
	for i = 0; i < this.ContentLength; i++ {
		if this.processes[i] == STATUS_NO_START {
			job.Start = i
			break
		}
	}

	if i >= this.ContentLength {
		job.Start = -1
		job.End = -1
		return job
	}

	job.End = job.Start + this.DownloadBlock
	for i = job.Start; i-job.Start < this.DownloadBlock && i < this.ContentLength; i++ {
		if this.processes[i] == STATUS_FINISH {
			job.End = i - 1
			break
		}
		job.End = i
		this.processes[i] = STATUS_START
	}

	return job
}

func (this *GoGetSchedules) FinishJob(job *GoGetBlock) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for i := job.Start; i < job.End; i++ {
		this.processes[i] = STATUS_FINISH
	}
	this.CompleteLength += (job.End - job.Start + 1)
}

var urlFlag = flag.String("u", "http://7b1h1l.com1.z0.glb.clouddn.com/bryce.jpg", "Fetch file url")
var cntFlag = flag.Int("c", 1, "Fetch concurrently counts")

func NewGoGet() *GoGet {
	get := new(GoGet)
	get.FilePath = "./"
	get.GetClient = new(http.Client)
	get.processBar = process_bar.NewProcessBar(0)

	flag.Parse()
	get.Url = *urlFlag
	get.Cnt = *cntFlag

	return get
}

func (get *GoGet) producer() {
	for {
		job := get.Schedule.NextJob()
		if job.Start == -1 {
			break
		}
		get.jobs <- job
	}
}

func (get *GoGet) consumer() {
	for {
		select {
		case job := <-get.jobs:
			get.jobStatus <- job
			go get.Download(job)
		}
	}
}

func (get *GoGet) Download(job *GoGetBlock) {
	range_i := fmt.Sprintf("%d-%d", job.Start, job.End)

	DebugLog.Printf("Download block [%s].", range_i)

	req, err := http.NewRequest("GET", get.Url, nil)
	req.Header.Set("Range", "bytes="+range_i)
	resp, err := get.GetClient.Do(req)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()
	if err != nil {
		DebugLog.Printf("Download %s error %v.\n", range_i, err)
	} else {
		res, _ := ioutil.ReadAll(resp.Body)
		for i := 0; i < len(res); i++ {
			get.raw[int64(i)+job.Start] = res[i]
		}
		get.Schedule.FinishJob(job)
	}

	<-get.jobStatus
}

func (get *GoGet) Start() {
	req, err := http.NewRequest("HEAD", get.Url, nil)
	resp, err := get.GetClient.Do(req)
	get.Header = resp.Header
	if err != nil {
		log.Panicf("Get %s error %v.\n", get.Url, err)
	}
	get.MediaType, get.MediaParams, _ = mime.ParseMediaType(get.Header.Get("Content-Disposition"))
	get.raw = make([]byte, resp.ContentLength, resp.ContentLength)
	get.Schedule = NewGoGetSchedules(resp.ContentLength)

	if strings.HasSuffix(get.FilePath, "/") {
		get.FilePath += get.MediaParams["filename"]
	} else {
	}

	if get.MediaParams["filename"] == "" {
		i := strings.LastIndex(get.Url, "/")
		if i != -1 && i+1 <= len(get.Url) {
			get.FilePath += get.Url[i+1:]
		}
	}
	get.File, err = os.Create(get.FilePath)
	if err != nil {
		log.Panicf("Create file %s error %v.\n", get.FilePath, err)
	}
	log.Printf("Get %s MediaType:%s, Filename:%s, Size %d.\n", get.Url, get.MediaType, get.MediaParams["filename"], get.Schedule.ContentLength)
	if get.Header.Get("Accept-Ranges") != "" {
		log.Printf("Server %s support Range by %s.\n", get.Header.Get("Server"), get.Header.Get("Accept-Ranges"))
	} else {
		log.Printf("Server %s doesn't support Range.\n", get.Header.Get("Server"))
	}

	log.Printf("Start to download %s(%d bytes) with %d thread.\n", get.MediaParams["filename"], get.Schedule.ContentLength, get.Cnt)

	get.jobs = make(chan *GoGetBlock, get.Cnt)
	get.jobStatus = make(chan *GoGetBlock, get.Cnt)
	go get.producer()
	go get.consumer()

	for get.Schedule.Percent() != 1 {
		get.processBar.Process(int(get.Schedule.Percent()*100), get.Schedule.Speed())
		time.Sleep(1 * time.Second)
	}
	if get.Schedule.Percent() == 1 {
		get.processBar.Process(100, get.Schedule.Speed())
	}

	get.File.Write(get.raw)
	get.File.Close()
	log.Printf("Download complete and store file %s.\n", get.FilePath)
}

func (get *GoGet) Stop() {

}

var DEFAULT_GET *GoGet

func init() {
	debuglogFile, logErr := os.OpenFile("debug.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)

	if logErr != nil {
		fmt.Println("Fail to find", "debug.log", " start Failed")
	}
	DebugLog = log.New(debuglogFile, "", log.LstdFlags)

	DEFAULT_GET = NewGoGet()
	// // handle ^c
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt)
	// go func() {
	// 	for {
	// 		select {
	// 		case <-c:
	// 			b, _ := json.Marshal(DEFAULT_GET)
	// 			io.Copy(DEFAULT_GET.File, bytes.NewReader(b))
	// 		}
	// 	}
	// }()
}
