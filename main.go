package main

import (
	"context"
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

type Program struct {
	newFiles      []*File
	oldFiles      []*File
	root          string
	pattern       string
	setupCMD      string
	updateCMD     string
	updateContext context.Context
	teardownCMD   string
	interval      int
	debug         bool
}

type File struct {
	path      string
	timestamp int64
}

func (p *Program) getFiles() []*File {
	p.newFiles = make([]*File, 0)
	filepath.Walk(p.root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			panic("error walking files")
		}
		if !info.IsDir() && p.isWatchFile(path) {
			p.newFiles = append(p.newFiles, &File{path: path, timestamp: modTime(path)})
		}
		return nil
	})
	return p.newFiles
}

func (p *Program) isWatchFile(path string) bool {
	re := regexp.MustCompile(p.pattern)
	return re.Match([]byte(path))
}

func (p *Program) setup() {
	p.log(fmt.Sprintf("Run setup command -- %s", p.setupCMD))
	cmd := makeCmd(p.setupCMD, context.Background())
	cmd.Run()
}

func (p *Program) update() {
	p.log(fmt.Sprintf("Run update command -- %s", p.updateCMD))
	p.updateContext = context.Background()
	cmd := makeCmd(p.updateCMD, p.updateContext)
	cmd.Start()
}

func (p *Program) teardown() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		<-ch
		signal.Stop(ch)
		p.log(fmt.Sprintf("Run teardown command -- %s", p.teardownCMD))
		cmd := makeCmd(p.teardownCMD, context.Background())
		cmd.Run()
		os.Exit(0)
	}()
}

func (p *Program) isOutOfDate() bool {
	if len(p.oldFiles) != len(p.newFiles) {
		return true
	}
	for i, file := range p.newFiles {
		if file.path != p.oldFiles[i].path || file.timestamp != p.oldFiles[i].timestamp {
			return true
		}
	}
	return false
}

func (p *Program) run() {
	p.setup()
	p.update()
	p.teardown()
	p.getFiles()
	p.oldFiles = p.newFiles
	for {
		p.log(fmt.Sprintf("Sleep for %d seconds", p.interval))
		time.Sleep(time.Second * time.Duration(p.interval))
		p.getFiles()
		if p.isOutOfDate() {
			p.oldFiles = p.newFiles
			p.log("Detected file change")
			p.updateContext.Done()
			p.update()
		}
	}
}

func modTime(path string) int64 {
	file, err := os.Stat(path)
	if err != nil {
		panic("cannot find file")
	}
	return file.ModTime().UnixMilli()
}

func makeCmd(c string, ctx context.Context) *exec.Cmd {
	var cmd *exec.Cmd
	command := strings.Split(c, " ")
	if len(command) > 1 {
		cmd = exec.CommandContext(ctx, command[0], command[1:]...)
	} else {
		cmd = exec.CommandContext(ctx, c)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func (p *Program) log(s string) {
	if p.debug {
		log.Println(s)
	}
}

func main() {
	setup := flag.String("setup", "go build -o=__temp__", "Custom build command")
	update := flag.String("update", "./__temp__", "Custom run command")
	teardown := flag.String("teardown", "rm __temp__", "Custom teardown command")
	fileTypes := flag.String("types", "go,html,css,js,tmpl", "File types to monitor")
	root := flag.String("root", ".", "Root folder to monitor")
	debug := flag.Bool("debug", false, "Activates debug logs")
	flag.Parse()

	pattern := fmt.Sprintf(".+\\.(%s)", strings.ReplaceAll(*fileTypes, ",", "|"))
	program := &Program{
		nil,
		nil,
		*root,
		pattern,
		*setup,
		*update,
		context.Background(),
		*teardown,
		1,
		*debug}
	program.run()
}
