package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
)

func Build() error {
	args := append([]string{"build", "-o", NewGGConfig().AppName + NewGGConfig().AppSuffix}, NewGGConfig().MainApplication...)
	cmd := exec.Command("go", args...)
	log.Println(strings.Join(cmd.Args, " "))
	var err_output bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &err_output

	if err := cmd.Start(); err != nil { //Use start, not run
		log.Println("An error occured: ", err) //replace with logger, or anything you want
	}

	if err := cmd.Wait(); err != nil {
		log.Printf("Build error: %v. %s.\n", err, string(err_output.Bytes()))
		return err
	}
	log.Println("Build Success.")
	return nil
}
