package ts

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Counter struct {
	name  string
	count int
}

type ByCount []Counter

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCount) Less(i, j int) bool { return a[i].count > a[j].count }

type Bucket struct {
	ip        []Counter
	section   []Counter
	bytes     int64
	hits      int
	timestamp time.Time
}

type TimeSeries []*Bucket

func (ts TimeSeries) TotalHits() int64 {
	hits := int64(0)
	for _, v := range ts {
		hits += int64(v.hits)
	}
	return hits
}

func (ts TimeSeries) TotalBytes() int64 {
	bytes := int64(0)
	for _, v := range ts {
		bytes += int64(v.bytes)
	}
	return bytes
}

func (ts TimeSeries) AverageHits(buckets int) (average_hits int) {
	average_hits = 0
	b := int(math.Abs(math.Min(float64(len(ts)), float64(buckets))))
	for _, v := range ts[len(ts)-b:] {
		average_hits += v.hits
	}
	average_hits = average_hits / b
	return
}

func (ts TimeSeries) AverageBytes(buckets int) (average_bytes int64) {
	average_bytes = 0
	b := int64(math.Min(float64(len(ts)), float64(buckets)))
	for _, v := range ts[int64(len(ts))-b:] {
		average_bytes += v.bytes
	}
	average_bytes = average_bytes / b
	return
}

func (ts TimeSeries) LastBucket() *Bucket {
	bucket := ts[len(ts)-1]
	return bucket
}

type RawEvent struct {
	IP    string
	Time  time.Time
	Verb  string
	Path  string
	Proto string
	Bytes int64
}

func CollectBuckets(buckets TimeSeries, refreshInterval time.Duration, alertThreshold int, eventChan chan RawEvent) {
	events := make([]*RawEvent, 0)
	refreshTimer := time.Tick(refreshInterval)

	for {
		select {
		case event := <-eventChan:
			events = append(events, &event)
		case <-refreshTimer:

			//_ip := make(map[string]int)
			//_section := make(map[string]int)

			ipCounters := map[string]Counter{}
			sectionCounters := map[string]Counter{}

			//ipCounters := make([]Counter, 0)
			//sectionCounters := make([]Counter, 0)

			var bytes int64 = 0
			var hits int = 0
			timestamp := time.Now().Local()

			for _, event := range events {

				c, ok := ipCounters[event.IP]
				if !ok {
					ipCounters[event.IP] = Counter{name: event.IP, count: 1}
				} else {
					c.count += 1
					ipCounters[event.IP] = c
				}

				path := strings.Split(event.Path, "/")[1]

				s, ok := sectionCounters[path]
				if !ok {
					sectionCounters[path] = Counter{name: path, count: 1}
				} else {
					s.count += 1
					sectionCounters[path] = s
				}

				bytes += event.Bytes
				hits++
			}

			events = events[0:0]

			//sort the inputs

			list := make([]Counter, 0)
			for _, v := range ipCounters {
				list = append(list, v)
			}

			sList := make([]Counter, 0)
			for _, v := range sectionCounters {
				sList = append(sList, v)
			}

			sort.Sort(ByCount(list))
			sort.Sort(ByCount(sList))

			//this is a race conditions
			buckets = append(buckets, &Bucket{list, sList, bytes, hits, timestamp})

			//something to draw the updated stats

			//alert on number of threshold hits

			//monitorHits
			go checkThreshold(buckets, alertThreshold)
		}
	}
}

var alert_fail_state = false

func checkThreshold(buckets TimeSeries, threshold int) {
	avg := buckets.AverageHits(6)

	if avg > threshold {
		if !alert_fail_state {
			alert_fail_state = true
			message := []string{"avg hits- ", strconv.Itoa(avg), " in last 2m exceeded alert_threshold of ", strconv.Itoa(threshold), " at ", time.Now().Local().String()}
			fmt.Println(message)
		}

		if alert_fail_state {
			message := []string{"avg hits- ", strconv.Itoa(avg), " in last 2m exceeded alert_threshold of ", strconv.Itoa(threshold), " at ", time.Now().Local().String()}
			fmt.Println(message)
		}
	}

	if avg < threshold && alert_fail_state {
		alert_fail_state = false
		message := []string{"avg hits- ", strconv.Itoa(avg), " in last 2m below alert_threshold of ", strconv.Itoa(threshold), " at ", time.Now().Local().String()}
		fmt.Println(message)
	}
}
