package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/mitchellh/go-ps"
)

func Backup() error {
	return deleteFile([]string{NewGGConfig().RunDirectory + NewGGConfig().AppName + NewGGConfig().AppSuffix})
}

func Deploy() {
	if err := CopyFile(NewGGConfig().AppPath, NewGGConfig().RunDirectory+NewGGConfig().AppName+".tar.gz"); err != nil {
		log.Println("Copy file error: ", err)
		return
	}

	if err := unPackFile(NewGGConfig().RunDirectory + NewGGConfig().AppName + ".tar.gz"); err != nil {
		log.Println("Tar package error: ", err)
	} else {
		pss, _ := ps.Processes()
		for _, p := range pss {
			if strings.Index(p.Executable(), NewGGConfig().AppName) != -1 {
				log.Printf("[pgrep %s] got pid: %d.\n", NewGGConfig().AppName, p.Pid())
				killProcess(p.Pid())
			}
		}
	}
}

func CopyFile(src, dst string) (err error) {
	log.Println("Copy file", src, "to", dst)

	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return err
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func killProcess(pid int) {
	log.Println("kill process", pid)
	cmd := fmt.Sprintf("kill %v", pid)
	runCommand(cmd)
}

func runCommand(cmd string) (string, error) {
	res, err := exec.Command("/bin/sh", "-c", cmd).Output()
	return string(res), err
}

func deleteFile(includeDir []string) error {
	for i := len(includeDir) - 1; i >= 0; i-- {
		err := os.RemoveAll(includeDir[i])
		if err != nil {
			return err
		} else {
			log.Println("Delete file", includeDir[i])
		}
	}
	return nil
}
