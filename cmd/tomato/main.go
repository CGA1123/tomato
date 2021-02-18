package main

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/CGA1123/tomato"
	"github.com/soellman/pidfile"
)

func serverRunning() bool {
	pid, err := pidfileContents(tomato.PidFile)
	if err != nil {
		return false
	}

	return pidIsRunning(pid)
}

func pidfileContents(filename string) (int, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(contents)))
	if err != nil {
		return 0, pidfile.ErrFileInvalid
	}

	return pid, nil
}

func pidIsRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))

	if err != nil && err.Error() == "no such process" {
		return false
	}

	if err != nil && err.Error() == "os: process already finished" {
		return false
	}

	return true
}

// TODO: add flags and parse them, implement cli
func main() {
	log.SetFlags(0)
	log.SetPrefix(tomato.LogPrefix)

	if !serverRunning() {
		log.Printf("tomato-server is not running")
		log.Printf("you need to start it before attempting to use tomato")
		return
	}

	client, err := tomato.NewClient(tomato.Socket)
	if err != nil {
		log.Printf("error creating client: %v", err)
		log.Printf("is the server running? start it with tomato-server")
		return
	}

	ends, err := client.Start()
	if err != nil {
		if err.Error() == tomato.ErrTomatoIsRunning.Error() {
			log.Printf("tomato is still running!")
			log.Printf("use tomato stop or tomato start -f")
		} else {
			log.Printf("error: %v", err)
		}
		return
	}

	log.Printf("got: %v", ends)
}
