package main

import (
	"crypto/md5"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"
	"time"
)

var m = make(map[string][16]byte)

func errorHandling(err error) {
	if err != nil {
		log.Fatalf("Fatal Error: %s", err)
	}
}

func getAllFiles(root string) []string {
	var files []string
	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		errorHandling(err)
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files
}

func getWatchFiles(root string, pattern string) []string {
	var files []string
	re := regexp.MustCompile(pattern)
	for _, file := range getAllFiles(root) {
		if re.Match([]byte(file)) {
			files = append(files, file)
		}
	}
	return files
}

func fileContenHash(path string) [16]byte {
	content, err := os.ReadFile(path)
	errorHandling(err)
	return md5.Sum(content)
}

func setup(pattern string) {
	m = make(map[string][16]byte)
	files := getWatchFiles(".", pattern)
	for _, file := range files {
		m[file] = fileContenHash(file)
	}
}

func isEqual(pattern string) bool {
	files := getWatchFiles(".", pattern)
	for _, file := range files {
		val, ok := m[file]
		if !ok || val != fileContenHash(file) || len(files) != len(m) {
			return false
		}
	}
	return true
}

func runCommand() *exec.Cmd {
	temp := exec.Command("go", "build", "-o=__temp__")
	temp.Run()
	cmd := exec.Command("./__temp__")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	return cmd
}

func main() {
	// Teardown
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		signalType := <-ch
		signal.Stop(ch)
		fmt.Println("Exit command received. Exiting...", signalType)
		os.Remove("temp")
		os.Exit(0)
	}()

	pattern := ".+\\.(go|tmpl|html)"
	setup(pattern)
	cmd := runCommand()
	for {
		time.Sleep(time.Second * 1)
		if !isEqual(pattern) {
			fmt.Println("Update detected")
			setup(pattern)
			cmd.Process.Kill()
			cmd = runCommand()
		}
	}
}
