# LMCHTTP

## Summary

Consume an actively written-to w3c-formatted HTTP access log (https://en.wikipedia.org/wiki/Common_Log_Format)
Every 10s, display in the console the sections of the web site with the most hits (a section is defined as being what's before the second '/' in a URL. i.e. the section for "http://my.site.com/pages/create' is "http://my.site.com/pages"), as well as interesting summary statistics on the traffic as a whole.
Make sure a user can keep the console app running and monitor traffic on their machine
Whenever total traffic for the past 2 minutes exceeds a certain number on average, add a message saying that “High traffic generated an alert - hits = {value}, triggered at {time}”
Whenever the total traffic drops again below that value on average for the past 2 minutes, add another message detailing when the alert recovered
Make sure all messages showing when alerting thresholds are crossed remain visible on the page for historical reasons.
Write a test for the alerting logic

## Requirements

- Go 1.6+
- govendor


## Setup

```bash
make setup
make build
```

```bash
./bin/lmchttp -log-file access_log
```

## Command Line Options

```bash
  -alert-threshold int
      hit threshold to trigger a warning (default 100)
  -debug
      enter debug mode which doesnt display a gui
  -log-file string
      path to the log file to process
  -refresh-interval string
      default refresh interval (default "10s")
```

## Future Improvements

* The consol ui could be improved big time
* The code in general could be refactored in several places where there are possible data races.
* More accurate calculation of average with EWMA.
* Different type of alerts with different thresholds like status code alerts.
* Average is a bad metric in general, a histogram would be useful to show.
* Better data caching
* Separating the UI from the daemon process running
