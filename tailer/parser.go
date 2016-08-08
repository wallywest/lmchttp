package tailer

import (
	"regexp"
	"strconv"
	"time"

	"github.com/hpcloud/tail"
	"github.com/wallywest/lmchttp/ts"
)

var commonRegex = regexp.MustCompile(`^(?P<ip>[\d\.]+) - - \[(?P<timestamp>.*)\] "(?P<verb>.*) (?P<query>.*) (?P<proto>.*)" (?P<status>\d+) (?P<bytes>\d+) "(?P<referer>.*)" "(?P<useragent>.*)"`)

func ReadLog(logFile string, eventChan chan ts.RawEvent, logDumpChan chan string) error {

	var seek = tail.SeekInfo{Offset: 0, Whence: 2}

	t, err := tail.TailFile(logFile, tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: true,
		Location:  &seek,
		Logger:    tail.DiscardingLogger,
	})

	if err != nil {
		return err
	}

	for line := range t.Lines {
		match := commonRegex.FindStringSubmatch(line.Text)

		if len(match) == 0 {
			continue
		}

		ip := match[1]
		current_time := time.Now().Local()
		verb := match[3]
		query := match[4]
		proto := match[5]
		//status, _ := strconv.Atoi(match[6])
		bytes, _ := strconv.ParseInt(match[7], 10, 64)

		event := ts.RawEvent{ip, current_time, verb, query, proto, bytes}

		eventChan <- event
		//logDumpChan <- match[0]
	}

	return nil
}
