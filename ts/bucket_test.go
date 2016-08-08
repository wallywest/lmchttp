package ts

import (
	"fmt"
	"regexp"
	"testing"
	"time"
)

type _TestMonitorHitsConfig struct {
	buckets_to_create    int
	expected_alert_text  string
	expected_alert_state bool
	avg_hits_per_bucket  int
}

func _TestHits(t *testing.T, c _TestMonitorHitsConfig) {
	buckets := make([]*Bucket, 0)

	for i := 0; i < c.buckets_to_create; i++ {
		counts := int64(c.avg_hits_per_bucket)
		b := BuildBucket(counts)
		buckets = append(buckets, &b)
	}

	alertChan := make(chan string)

	go checkThreshold(buckets, 100, alertChan)

	message := <-alertChan

	if match, _ := regexp.MatchString(c.expected_alert_text, message); !match {
		t.Error("Actual alert text [", message, "] did not match expected string [", c.expected_alert_text, "]")
	}
}

func BuildBucket(counts ...int64) Bucket {
	hits := 400
	if len(counts) > 0 {
		hits = int(counts[0])
	}

	bytes := int64(10000)

	if len(counts) > 1 {
		bytes = counts[1]
	}

	bucket := Bucket{
		[]Counter{{"2.3.4.5", 10}, {"3.4.5.6", 20}, {"4.5.6.7", 30}},
		[]Counter{{"foo", 10}, {"bar", 20}, {"bam", 30}},
		bytes,
		hits,
		time.Now(),
	}

	return bucket
}

func TestExceedHitsThreshold(t *testing.T) {
	c := _TestMonitorHitsConfig{}
	c.buckets_to_create = 200
	c.avg_hits_per_bucket = 101
	c.expected_alert_text = fmt.Sprint("avg hits - ", c.avg_hits_per_bucket, " in last 2m exceeded Alert Threshold of ", 100, " at ")
	c.expected_alert_state = true

	_TestHits(t, c)
}

func TestBelowHitsThreshold(t *testing.T) {
	c := _TestMonitorHitsConfig{}
	c.buckets_to_create = 200
	c.avg_hits_per_bucket = 99
	c.expected_alert_text = fmt.Sprint("avg hits - ", c.avg_hits_per_bucket, " in last 2m are below Alert Threshold of ", 100, " at ")
	c.expected_alert_state = false

	_TestHits(t, c)
}
