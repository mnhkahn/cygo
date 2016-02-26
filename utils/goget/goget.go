package goget

import (
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	DEFAULT_DOWNLOAD_BLOCK int64 = 1048576 // 2^20
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
	// ContentLength  int64
	// CompleteLength int64
	// DownloadRange [][]int64
	File      *os.File
	TempFiles []*os.File
	WG        sync.WaitGroup
	raw       []byte
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

func (this *GoGetSchedules) NextJob() *GoGetBlock {
	job := new(GoGetBlock)

	var i int64
	for i = 0; i < this.ContentLength; i++ {
		if this.processes[i] == STATUS_NO_START {
			job.Start = i
		}
	}
	job.End = job.Start + this.DownloadBlock
	for i = job.Start; i-job.Start <= this.DownloadBlock && i < this.ContentLength; i++ {
		if this.processes[i] == STATUS_FINISH {
			job.End = i - 1
			break
		}
	}
	return job
}

func (this *GoGetSchedules) FinishJob(job *GoGetBlock) {

}

func (this *GoGetSchedules) Percent() float32 {
	return float32(this.CompleteLength) / float32(this.ContentLength)
}

var urlFlag = flag.String("u", "http://7b1h1l.com1.z0.glb.clouddn.com/bryce.jpg", "Fetch file url")
var cntFlag = flag.Int("c", 1, "Fetch concurrently counts")

func NewGoGet() *GoGet {
	get := new(GoGet)
	get.FilePath = "./"
	get.GetClient = new(http.Client)

	flag.Parse()
	get.Url = *urlFlag
	get.Cnt = *cntFlag

	return get
}

func (get *GoGet) producer(jobs chan *GoGetBlock) {
	for {
		job := new(GoGetBlock)
		// job.Start = get.CompleteLength
		// job.End = job.Start + get.DownloadBlock
		jobs <- job
	}
}

func (get *GoGet) consumer(jobs chan *GoGetBlock) {
	for {
		select {
		case job := <-jobs:
			go get.Download(job)
		}
	}
}

func (get *GoGet) Download(job *GoGetBlock) {
	fmt.Println(job.Start, job.End, "Download")
	time.Sleep(5 * time.Second)
}

func (get *GoGet) Start() {
	req, err := http.NewRequest("HEAD", get.Url, nil)
	resp, err := get.GetClient.Do(req)
	get.Header = resp.Header
	if err != nil {
		log.Panicf("Get %s error %v.\n", get.Url, err)
	}
	get.MediaType, get.MediaParams, _ = mime.ParseMediaType(get.Header.Get("Content-Disposition"))
	get.Schedule.ContentLength = resp.ContentLength

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
	get.File, err = os.Create(get.FilePath + ".tmp")
	if err != nil {
		log.Panicf("Create file %s error %v.\n", get.FilePath, err)
	}
	log.Printf("Get %s MediaType:%s, Filename:%s, Size %d.\n", get.Url, get.MediaType, get.MediaParams["filename"], get.Schedule.ContentLength)
	if get.Header.Get("Accept-Ranges") != "" {
		log.Printf("Server %s support Range by %s.\n", get.Header.Get("Server"), get.Header.Get("Accept-Ranges"))
	} else {
		log.Printf("Server %s doesn't support Range.\n", get.Header.Get("Server"))
	}

	log.Printf("Start to download %s with %d thread.\n", get.MediaParams["filename"], get.Cnt)

	channels := make(chan *GoGetBlock, get.Cnt)
	go get.producer(channels)
	go get.consumer(channels)

	time.Sleep(15 * time.Second)
	// 	var range_start int64 = 0
	// 	for i := 0; i < get.Cnt; i++ {
	// 		if i != get.Cnt-1 {
	// 			get.DownloadRange = append(get.DownloadRange, []int64{range_start, range_start + get.DownloadBlock - 1})
	// 		} else {
	// 			// 最后一块
	// 			get.DownloadRange = append(get.DownloadRange, []int64{range_start, get.ContentLength - 1})
	// 		}
	// 		range_start += get.DownloadBlock
	// 	}
	// 	// Check if the download has paused.
	// 	for i := 0; i < len(get.DownloadRange); i++ {
	// 		range_i := fmt.Sprintf("%d-%d", get.DownloadRange[i][0], get.DownloadRange[i][1])
	// 		temp_file, err := os.OpenFile(get.FilePath+"."+range_i, os.O_RDONLY|os.O_APPEND, 0)
	// 		if err != nil {
	// 			temp_file, _ = os.Create(get.FilePath + "." + range_i)
	// 		} else {
	// 			fi, err := temp_file.Stat()
	// 			if err == nil {
	// 				get.DownloadRange[i][0] += fi.Size()
	// 			}
	// 		}
	// 		get.TempFiles = append(get.TempFiles, temp_file)
	// 	}

	// 	go get.Watch()
	// 	get.Latch = get.Cnt
	// 	for i, _ := range get.DownloadRange {
	// 		get.WG.Add(1)
	// 		go get.Download(i)
	// 	}

	// 	get.WG.Wait()

	// 	for i := 0; i < len(get.TempFiles); i++ {
	// 		temp_file, _ := os.Open(get.TempFiles[i].Name())
	// 		cnt, err := io.Copy(get.File, temp_file)
	// 		if cnt <= 0 || err != nil {
	// 			log.Printf("Download #%d error %v.\n", i, err)
	// 		}
	// 		temp_file.Close()
	// 	}
	// 	get.File.Close()
	// 	log.Printf("Download complete and store file %s with %v.\n", get.FilePath, time.Now().Sub(download_start))
	// 	defer func() {
	// 		for i := 0; i < len(get.TempFiles); i++ {
	// 			err := os.Remove(get.TempFiles[i].Name())
	// 			if err != nil {
	// 				log.Printf("Remove temp file %s error %v.\n", get.TempFiles[i].Name(), err)
	// 			} else {
	// 				log.Printf("Remove temp file %s.\n", get.TempFiles[i].Name())
	// 			}
	// 		}
	// 	}()
	// }

	// func (get *GoGet) Download(i int) {
	// 	defer get.WG.Done()
	// 	if get.DownloadRange[i][0] > get.DownloadRange[i][1] {
	// 		return
	// 	}
	// 	range_i := fmt.Sprintf("%d-%d", get.DownloadRange[i][0], get.DownloadRange[i][1])
	// 	log.Printf("Download #%d bytes %s.\n", i, range_i)

	// 	defer get.TempFiles[i].Close()

	// 	req, err := http.NewRequest("GET", get.Url, nil)
	// 	req.Header.Set("Range", "bytes="+range_i)
	// 	resp, err := get.GetClient.Do(req)
	// 	defer func() {
	// 		if resp != nil && resp.Body != nil {
	// 			resp.Body.Close()
	// 		}
	// 	}()
	// 	if err != nil {
	// 		log.Printf("Download #%d error %v.\n", i, err)
	// 	} else {
	// 		cnt, err := io.Copy(get.TempFiles[i], resp.Body)
	// 		if cnt == int64(get.DownloadRange[i][1]-get.DownloadRange[i][0]+1) {
	// 			log.Printf("Download #%d complete.\n", i)
	// 		} else {
	// 			req_dump, _ := httputil.DumpRequest(req, false)
	// 			resp_dump, _ := httputil.DumpResponse(resp, true)
	// 			log.Printf("Download error %d %v, expect %d-%d, but got %d.\nRequest: %s\nResponse: %s\n", resp.StatusCode, err, get.DownloadRange[i][0], get.DownloadRange[i][1], cnt, string(req_dump), string(resp_dump))
	// 		}
	// 	}
}

func (get *GoGet) Stop() {

}

// http://stackoverflow.com/questions/15714126/how-to-update-command-line-output
func (get *GoGet) Watch() {
	fmt.Printf("[=================>]\n")
}

func main() {
	DEFAULT_GET.Start()
}

var DEFAULT_GET *GoGet

func init() {
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
