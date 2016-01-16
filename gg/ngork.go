package main

import (
// "bytes"
// "log"
// "os"
// "os/exec"
// "strings"

// "github.com/mnhkahn/gg/conf"
)

func Ngrok() {
	// args := append([]string{"http", conf.NewGGConfig().NgrokPort})
	// cmd := exec.Command("ngrok", args...)
	// log.Println(strings.Join(cmd.Args, " "))
	// var err_output bytes.Buffer
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = &err_output

	// if err := cmd.Start(); err != nil { //Use start, not run
	// 	log.Println("An error occured: ", err) //replace with logger, or anything you want
	// }

	// if err := cmd.Wait(); err != nil {
	// 	log.Printf("Ngrok error: %v. %s.\n", err, string(err_output.Bytes()))
	// 	return
	// }
	// log.Println("Ngrok Success.")
}
