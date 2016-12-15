package goget

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	// "strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"

	// "github.com/mnhkahn/cygo/container/interval"
	"github.com/mnhkahn/cygo/utils/process_bar"
	"github.com/pyk/byten"
)

const (
	DEFAULT_DOWNLOAD_BLOCK int64 = 1024 * 16 // 2^20
	// DEFAULT_DOWNLOAD_BLOCK int64 = 1048576 // 2^20
)

type GoGet struct {
	Url string
	// u           *url.URL
	// Ips         []net.IP
	Cnt         int
	FailCnt     int // 连续失败次数
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
	DebugLog    *log.Logger
}

// 前闭后闭区间
type GoGetBlock struct {
	start int64
	end   int64
}

func NewGoGetBlock(s, e int64) *GoGetBlock {
	n := new(GoGetBlock)
	n.start, n.end = s, e
	return n
}

func (this *GoGetBlock) Start() int64 {
	return this.start
}

func (this *GoGetBlock) End() int64 {
	return this.end
}

func (this *GoGetBlock) SetStart(start int64) {
	this.start = start
}

func (this *GoGetBlock) SetEnd(end int64) {
	this.end = end
}

const (
	STATUS_NO_START = byte(0)
	STATUS_START    = byte(1)
	STATUS_FINISH   = byte(2)
)

type GoGetSchedules struct {
	processes []byte
	// noStartInterval *interval.Interval
	// startedInterval *interval.Interval
	// finishInterval  *interval.Interval
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
	speed := "0"
	if elaspe > 0 {
		speed = byten.Size(int64(float64(this.CompleteLength) / elaspe))
	}
	return fmt.Sprintf("%s/S     ", speed)
}

// get next download block size
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
			job.SetStart(i)
			// job.Start = i
			break
		}
	}

	if i >= this.ContentLength {
		job.SetStart(-1)
		job.SetEnd(-1)
		// job.Start = -1
		// job.End = -1
		return job
	}

	job.SetEnd(job.Start() + this.DownloadBlock)

	// job.End = job.Start + this.DownloadBlock
	for i = job.Start(); i-job.Start() < this.DownloadBlock && i < this.ContentLength; i++ {
		if this.processes[i] == STATUS_FINISH {
			job.SetEnd(i - 1)
			// job.End = i - 1
			break
		}
		job.SetEnd(i)
		// job.End = i
		this.processes[i] = STATUS_START
	}

	return job
}

func (this *GoGetSchedules) FinishJob(job *GoGetBlock) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for i := job.Start(); i < job.End(); i++ {
		this.processes[i] = STATUS_FINISH
	}
	this.CompleteLength += (job.End() - job.Start() + 1)
}

// func (this *GoGetSchedules) ResetJob(job *GoGetBlock) {
// 	this.lock.Lock()
// 	defer this.lock.Unlock()

// 	for i := job.Start; i < job.End; i++ {
// 		this.processes[i] = STATUS_NO_START
// 	}
// }

func (this *GoGetSchedules) IsComplete() bool {
	for _, process := range this.processes {
		if process != STATUS_FINISH {
			return false
		}
	}

	return true
}

func NewGoGet() *GoGet {
	get := new(GoGet)

	if runtime.GOOS == "windows" {
		// http://windowsitpro.com/systems-management/what-environment-variables-are-available-windows
		get.FilePath = strings.Replace(os.Getenv("HOMEDRIVE")+os.Getenv("HOMEPATH"), "\\", "/", -1) + "/Downloads/"
	} else if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		get.FilePath = os.Getenv("HOME") + "/Downloads/"
	} else {
		get.FilePath = "./"
	}

	debuglogFile, logErr := os.OpenFile(get.FilePath+"debug.log", os.O_CREATE|os.O_RDWR|os.O_TRUNC|os.O_APPEND, 0666)
	if logErr != nil {
		log.Println(logErr)
	}

	os.Stderr = debuglogFile
	get.DebugLog = log.New(debuglogFile, "", log.Lshortfile)

	// http 长连接 https://serholiu.com/go-http-client-keepalive
	get.GetClient = new(http.Client)
	get.GetClient.Transport = &http.Transport{
		Dial: KeepAliveDialTimeout,
	}

	get.processBar = process_bar.NewProcessBar(0)

	return get
}

func (get *GoGet) producer() {
	// downloadOnce := false
	for {
		job := get.Schedule.NextJob()

		if job.Start() == -1 && get.Schedule.IsComplete() {
			break
		}
		if job.Start() != -1 && job.End() != -1 {
			get.jobs <- job
		} else if job.Start() == -1 && job.End() == -1 {
			// downloadOnce = true
			break
		}

		// // 下载完成一次之后，1s钟检查一次
		// if downloadOnce {
		// 	time.Sleep(1 * time.Second)
		// }
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
	defer func() {
		if err := recover(); err != nil {
			get.DebugLog.Println(err)
			debug.PrintStack()
		}
	}()

	range_i := fmt.Sprintf("%d-%d", job.Start(), job.End())

	get.DebugLog.Printf("Download block [%s].", range_i)

	req, err := http.NewRequest("GET", get.Url, nil)
	req.Header.Set("Range", "bytes="+range_i)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.116 Safari/537.36")
	resp, err := get.GetClient.Do(req)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if err != nil || (resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK) {
		get.FailCnt++
		// get.Schedule.ResetJob(job)
		go get.Download(job)
		if resp == nil {
			get.DebugLog.Printf("Download %s error %v.\n", range_i, err)
		} else {
			get.DebugLog.Printf("Download %s error %v, %d.\n", range_i, err, resp.StatusCode)
		}
	} else {
		res, err := ioutil.ReadAll(resp.Body)
		if err != nil || int64(len(res)) != job.End()-job.Start()+1 {
			get.FailCnt++
			go get.Download(job)
			// get.Schedule.ResetJob(job)
			get.DebugLog.Printf("Download %s error %v.\n", range_i, err)
		} else {
			get.FailCnt = 0
			// Change to io.Copy
			for i := 0; i < len(res); i++ {
				get.raw[int64(i)+job.Start()] = res[i]
			}
			get.Schedule.FinishJob(job)
		}
	}

	<-get.jobStatus
}

func (get *GoGet) Start(config *GoGetConfig) {
	defer func() {
		if err := recover(); err != nil {
			get.DebugLog.Println(err)
			debug.PrintStack()
		}
	}()

	var err error

	get.Url = config.Url
	get.Cnt = config.Cnt

	// // cache dns
	// get.u, err = url.Parse(get.Url)
	// if err != nil {
	// 	panic(err)
	// }
	// get.Ips, err = net.LookupIP(get.u.Host)
	// fmt.Println(get.Ips, err, "AAAAAAAAAAAAa")

	if config.ProxyType == PROXYHTTP {
		proxy := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(config.Proxy)
		}

		get.GetClient.Transport = &http.Transport{Proxy: proxy}
	} else if config.ProxyType == PROXYSOCKS5 {
		dialer, err := proxy.SOCKS5("tcp", config.Proxy,
			nil,
			&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			},
		)
		if err != nil {
			return
		}

		get.GetClient.Transport = &http.Transport{
			Proxy:               nil,
			Dial:                dialer.Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		}
	}

	// var err error
	for i := 0; i < 5; i++ {
		err := get.InitLen()
		if err == nil {
			break
		} else {
			get.DebugLog.Println("Retry", i)
		}
	}

	get.File, err = os.Create(get.FilePath)
	if err != nil {
		log.Printf("Create file %s error %v.\n", get.FilePath, err)
		return
	}
	// log.Printf("Get %s MediaType:%s, Filename:%s, Size %d.\n", get.Url, get.MediaType, get.MediaParams["filename"], get.Schedule.ContentLength)
	if get.Header.Get("Accept-Ranges") != "" {
		// log.Printf("Server %s support Range by %s.\n", get.Header.Get("Server"), get.Header.Get("Accept-Ranges"))
	} else {
		log.Printf("Server %s doesn't support Range.\n", get.Header.Get("Server"))
	}

	log.Printf("Start to download %s(%s) with %d thread.\n", get.FilePath, byten.Size(get.Schedule.ContentLength), get.Cnt)

	get.jobs = make(chan *GoGetBlock, get.Cnt)
	get.jobStatus = make(chan *GoGetBlock, get.Cnt)
	go get.producer()
	go get.consumer()

	for get.Schedule.Percent() != 1 /*&& get.FailCnt < 300*/ {
		// get.processBar.Process(int(get.Schedule.Percent()*100), get.Schedule.Speed())
		get.processBar.Process(int(get.Schedule.Percent()*100), fmt.Sprintf("%s %d/%d %s", byten.Size(get.Schedule.DownloadBlock), len(get.jobs), cap(get.jobs), get.Schedule.Speed()))
		// fmt.Println(len(get.jobs), cap(get.jobs), get.Cnt)
		time.Sleep(1 * time.Second)
	}
	if get.Schedule.Percent() == 1 {
		// fmt.Println(get.Schedule.CompleteLength, get.Schedule.ContentLength, get.Schedule.Percent())
		get.processBar.Process(100, get.Schedule.Speed())
	}

	_, err = get.File.Write(get.raw)
	if err != nil {
		panic(err)
	}
	get.File.Close()

	h := md5.New()
	_, err = h.Write(get.raw)
	if err != nil {
		panic(err)
	}
	log.Printf("Download complete and store file %s. MD5: %x.\n", get.FilePath, h.Sum(nil))
	get.DebugLog.Println("========================================")
}

func (get *GoGet) InitLen() error {
	req, err := http.NewRequest("HEAD", get.Url, nil)
	if err != nil {
		return err
	}

	resp, err := get.GetClient.Do(req)
	if err != nil {
		log.Printf("Get %s error %v.\n", get.Url, err)
		return err
	}
	get.Header = resp.Header
	get.MediaType, get.MediaParams, _ = mime.ParseMediaType(get.Header.Get("Content-Disposition"))
	if resp.ContentLength <= 0 {
		// log.Printf("ContentLength error", resp.ContentLength)
		return fmt.Errorf("ContentLength error %d", resp.ContentLength)
	}

	get.raw = make([]byte, resp.ContentLength, resp.ContentLength)
	get.Schedule = NewGoGetSchedules(resp.ContentLength)

	if get.MediaParams["filename"] != "" {
		get.FilePath += get.MediaParams["filename"]
	} else if i := strings.LastIndex(get.Url, "/"); i != -1 && i+1 <= len(get.Url) {
		get.FilePath += get.Url[i+1:]
	} else {
		hash := md5.New()
		hash.Write([]byte(get.Url))
		get.FilePath += base64.StdEncoding.EncodeToString(hash.Sum(nil))
	}
	return nil
}

func (get *GoGet) Stop() {

}

const (
	NOPROXY     = 0
	PROXYHTTP   = 1
	PROXYSOCKS5 = 2
)

var (
	Proxys = map[string]int{"": NOPROXY, "http": PROXYHTTP, "socks5": PROXYSOCKS5}
)

type GoGetConfig struct {
	Url       string
	Cnt       int
	ProxyType int // 0 no proxy; 1 http; 2 socks5
	Proxy     string
}

func NewGoGetConfig() *GoGetConfig {
	config := new(GoGetConfig)
	return config
}

func NewGoGetConfig1(Url string, Cnt int, ProxyType string, Proxy string) *GoGetConfig {
	config := NewGoGetConfig()
	config.Url = Url
	config.Cnt = Cnt
	config.ProxyType = Proxys[ProxyType]
	config.Proxy = Proxy
	return config
}

var DEFAULT_GET *GoGet

func init() {
	DEFAULT_GET = NewGoGet()
	// handle ^c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for {
			select {
			case <-c:
				// b, _ := json.Marshal(DEFAULT_GET)
				// io.Copy(DEFAULT_GET.File, bytes.NewReader(b))
				DEFAULT_GET.DebugLog.Println("========================================")
				fmt.Println("")
				fmt.Println("")
				os.Exit(1)
			}
		}
	}()
}
