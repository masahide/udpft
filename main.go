package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"sync"

	"github.com/masahide/udpft/lib"
)

var checkSize = true
var checkMD5 = false
var workNum = 10
var cpPath = ""
var destPath = ""
var show_version = false
var serverMode = true
var startPort = 20300
var version string

func main() {
	// Parse the command-line flags.
	flag.BoolVar(&show_version, "version", false, "Show version")
	flag.BoolVar(&serverMode, "server", serverMode, "Server mode")
	flag.BoolVar(&checkSize, "checksize", true, "check size")
	flag.BoolVar(&checkMD5, "checkmd5", false, "check md5")
	flag.IntVar(&startPort, "port", startPort, "use port number")
	flag.IntVar(&workNum, "n", workNum, "max workers")
	flag.Parse()

	if show_version {
		fmt.Printf("version: %s\n", version)
		return
	}

	if flag.NArg() < 3 {
		fmt.Printf("Usage:\n %s [options] <filePath>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
		os.Exit(1)
	}
	cpPath = flag.Args()[0]

	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)

	// Generate Task
	done := make(chan struct{})
	defer close(done)

	// Start workers
	dq := make(chan lib.DataQueue, 100)
	results := make(chan error)
	var wg sync.WaitGroup
	wg.Add(workNum)
	for i := 0; i < workNum; i++ {
		go func() {
			lib.SendWorker(fmt.Sprintf("out%02d.dat", i), dq, results)
			wg.Done()
		}()
	}

	// wait work
	go func() {
		wg.Wait()
		close(results)
	}()

	end := make(chan int)
	errc := lib.FileToMemory(done, cpPath, dq, end)

	// Merge results
	for result := range results {
		log.Printf("%v", result)
	}

	// Check whether the work failed.
	if err := <-errc; err != nil {
		log.Printf("Error: %v", err)
	}

}
