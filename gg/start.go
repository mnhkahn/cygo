package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/howeyc/fsnotify"
)

var (
	cmd       *exec.Cmd
	state     sync.Mutex
	eventTime = make(map[string]int64)
)

func Start() {
	cmd = exec.Command("./"+NewGGConfig().AppName+NewGGConfig().AppSuffix, NewGGConfig().RunFlag)

	log.Println(strings.Join(cmd.Args, " "))
	var err_output bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &err_output

	if err := cmd.Start(); err != nil { //Use start, not run
		log.Println("An error occured: ", err) //replace with logger, or anything you want
	}

	if err := cmd.Wait(); err != nil {
		log.Printf("Start error: %v. %s.\n", err, string(err_output.Bytes()))
		return
	}
	log.Println("Start Success.")
}

func Watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)

	go func() {
		for {
			select {
			case e := <-watcher.Event:
				isbuild := true
				mt := getFileModTime(e.Name)
				if t := eventTime[e.Name]; mt == t {
					log.Printf("[SKIP] # %s #\n", e.String())
					isbuild = false
				}

				eventTime[e.Name] = mt

				if !strings.HasSuffix(e.Name, ".go") {
					isbuild = false
				}
				if isbuild {
					log.Println("event:", e)
					time.Sleep(1 * time.Second)
					AutoBuild()
				}
			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	for dir, _ := range NewGGConfig().FileWatcher {
		log.Println("Watchs on:", dir)
		err = watcher.Watch(dir)
		if err != nil {
			log.Fatal(err)
		}
	}

	<-done

	defer watcher.Close()
}

func AutoBuild() {
	state.Lock()
	defer state.Unlock()

	Kill()
	if err := Build(); err == nil {
		go Start()
	}
}

func Kill() {
	defer func() {
		if e := recover(); e != nil {
			log.Println("Kill.recover -> ", e)
		}
	}()
	if cmd != nil && cmd.Process != nil {
		err := cmd.Process.Kill()
		if err != nil {
			log.Println("Kill -> ", err)
		}
	}
}

func getFileModTime(path string) int64 {
	path = strings.Replace(path, "\\", "/", -1)
	f, err := os.Open(path)
	if err != nil {
		log.Printf("[ERRO] Fail to open file[ %s ]\n", err)
		return time.Now().Unix()
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		log.Printf("[ERRO] Fail to get file information[ %s ]\n", err)
		return time.Now().Unix()
	}

	return fi.ModTime().Unix()
}

func getFileInfo(path string) os.FileInfo {
	path = strings.Replace(path, "\\", "/", -1)
	f, err := os.Open(path)
	if err != nil {
		log.Printf("[ERRO] Fail to open file[ %s ]\n", err)
		return nil
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		log.Printf("[ERRO] Fail to get file information[ %s ]\n", err)
		return nil
	}

	return fi
}
