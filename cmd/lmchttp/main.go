package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wallywest/lmchttp/tailer"
	"github.com/wallywest/lmchttp/ts"
)

func main() {
	//gui initialization

	//logChan := make(chan RawLogEvent)

	logFile := flag.String("log-file", "", "path to the log file to process")
	refreshInterval := flag.String("refresh-interval", "10s", "default refresh interval")
	alertThreshold := flag.Int("alert-threshold", 100, "hit threshold to trigger a warning")

	flag.Parse()

	if *logFile == "" {
		fmt.Println("must specify a valid file")
		os.Exit(0)
	}

	duration, err := time.ParseDuration(*refreshInterval)
	if err != nil {
		fmt.Println("invalid time duration")
		os.Exit(0)
	}

	buckets := make([]*ts.Bucket, 0)

	eventChan := make(chan ts.RawEvent)
	logDumpChan := make(chan string)
	quitChan := make(chan bool)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := tailer.ReadLog(*logFile, eventChan, logDumpChan)
		if err != nil {
			fmt.Println(err)
			quitChan <- true
		}
		fmt.Println("finished")
	}()

	go ts.CollectBuckets(buckets, duration, *alertThreshold, eventChan)

	go func() {
		sig := <-sigs
		fmt.Println(sig)
		quitChan <- true
	}()

	<-quitChan
	fmt.Println("exiting")
}
