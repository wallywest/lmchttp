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

func (c *Counter) Name() string {
	return c.name
}

func (c *Counter) Count() int {
	return c.count
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

func (b *Bucket) Section() []Counter {
	return b.section
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

func CollectBuckets(buckets TimeSeries, refreshInterval time.Duration, alertThreshold int, eventChan chan RawEvent, cuiChan chan TimeSeries, alertChan chan string) {
	events := make([]*RawEvent, 0)
	refreshTimer := time.Tick(refreshInterval)

	for {
		select {
		case event := <-eventChan:
			events = append(events, &event)
		case <-refreshTimer:

			ipCounters := map[string]Counter{}
			sectionCounters := map[string]Counter{}

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

			//fmt.Println(buckets)
			go checkThreshold(buckets, alertThreshold, alertChan)
			//go printDebug(buckets)

			cuiChan <- buckets
			//something to draw the updated stats

			//alert on number of threshold hits
			//monitorHits
		}
	}
}

var alert_fail_state = false

func printDebug(timeseries TimeSeries) {
	message := ""

	message += fmt.Sprint(" Avg Hits: ", timeseries.AverageHits(12))
	message += fmt.Sprint("  Avg Bytes: ", timeseries.AverageBytes(12))
	message += fmt.Sprint(" Total Hits: ", timeseries.TotalHits())
	message += fmt.Sprint("  Total Bytes: ", timeseries.TotalBytes())

	fmt.Println(message)

	sectionMessage := ""
	for _, counter := range timeseries.LastBucket().Section() {
		sectionMessage += fmt.Sprint(" /", counter.Name(), " : ", strconv.Itoa(counter.Count()), "\n")
	}

	fmt.Println(sectionMessage)
}

func checkThreshold(buckets TimeSeries, threshold int, alertChan chan string) {
	avg := buckets.AverageHits(12)

	if avg > threshold {
		if !alert_fail_state {
			alert_fail_state = true
		}
		message := []string{"avg hits - ", strconv.Itoa(avg), " in last 2m exceeded Alert Threshold of ", strconv.Itoa(threshold), " at ", time.Now().Local().String()}
		alertChan <- strings.Join(message, "")
		return
	}

	if avg < threshold && alert_fail_state {
		alert_fail_state = false
		message := []string{"avg hits - ", strconv.Itoa(avg), " in last 2m are below Alert Treshold of ", strconv.Itoa(threshold), " at ", time.Now().Local().String()}
		alertChan <- strings.Join(message, "")
		return
	}
}
