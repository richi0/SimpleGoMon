package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
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

func runCommand(build string, run string) *exec.Cmd {
	var temp *exec.Cmd
	var cmd *exec.Cmd
	b := strings.Split(build, " ")
	if len(b) > 1 {
		temp = exec.Command(b[0], b[1:]...)
	} else if len(b) == 1 {
		temp = exec.Command(build)
	} else {
		log.Panic("Build command error", build)
	}
	temp.Run()
	r := strings.Split(run, " ")
	if len(r) > 1 {
		cmd = exec.Command(r[0], r[1:]...)
	} else if len(r) == 1 {
		cmd = exec.Command(run)
	} else {
		log.Panic("Run command error", run)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	return cmd
}

func main() {
	buildCmd := flag.String("build", "go build -o=__temp__", "Custom build command")
	runCmd := flag.String("run", "./__temp__", "Custom run command")
	tearDownCmd := flag.String("tearDown", "rm __temp__", "Custom tear down command")
	fileTypes := flag.String("types", "go,htmp,css,js,tmpl", "File types to check for changes, Example: -types=go,html,js ")
	flag.Parse()

	// Teardown
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		var cmd *exec.Cmd
		signalType := <-ch
		signal.Stop(ch)
		fmt.Println("Exit command received. Exiting...", signalType)
		c := strings.Split(*tearDownCmd, " ")
		if len(c) > 1 {
			cmd = exec.Command(c[0], c[1:]...)
		} else if len(c) == 1 {
			cmd = exec.Command(*tearDownCmd)
		} else {
			log.Panic("Build command error", cmd)
		}
		cmd.Run()
		os.Exit(0)
	}()

	pattern := fmt.Sprintf(".+\\.(%s)", strings.ReplaceAll(*fileTypes, ",", "|"))
	setup(pattern)
	cmd := runCommand(*buildCmd, *runCmd)
	for {
		time.Sleep(time.Second * 1)
		if !isEqual(pattern) {
			fmt.Println("Update detected")
			setup(pattern)
			cmd.Process.Kill()
			cmd = runCommand(*buildCmd, *runCmd)
		}
	}
}
