package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/wallywest/lmchttp/cui"
	"github.com/wallywest/lmchttp/tailer"
	"github.com/wallywest/lmchttp/ts"
)

func main() {
	logFile := flag.String("log-file", "", "path to the log file to process")
	refreshInterval := flag.String("refresh-interval", "10s", "default refresh interval")
	alertThreshold := flag.Int("alert-threshold", 100, "hit threshold to trigger a warning")
	debug := flag.Bool("debug", false, "enter debug mode which doesnt display a gui")

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
	cuiChan := make(chan ts.TimeSeries)
	alertChan := make(chan string)
	quitChan := make(chan bool)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := tailer.ReadLog(*logFile, eventChan, logDumpChan)
		if err != nil {
			fmt.Println(err)
			quitChan <- true
		}
		fmt.Println("finished")
	}()

	go ts.CollectBuckets(buckets, duration, *alertThreshold, eventChan, cuiChan, alertChan)

	if *debug {
		go func() {
			sigs := <-sigChan
			fmt.Println(sigs)
			quitChan <- true
		}()
	} else {
		go func() {
			gui := cui.New(cuiChan, logDumpChan, alertChan)

			err := gui.MainLoop()
			if err != nil && err == gocui.ErrQuit {
				gui.Close()
				quitChan <- true
			}
		}()
	}

	<-quitChan
}
