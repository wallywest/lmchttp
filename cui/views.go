package cui

import (
	"fmt"
	"strconv"

	"github.com/jroimartin/gocui"
	"github.com/wallywest/lmchttp/ts"
)

func updateSectionView(gui *gocui.Gui, timeseries ts.TimeSeries) {
	gui.Execute(func(g *gocui.Gui) error {
		v, err := g.View("sections")
		if err != nil {
			return err
		}
		v.Clear()
		message := ""

		for _, counter := range timeseries.LastBucket().Section() {
			message += fmt.Sprint(" /", counter.Name(), " : ", strconv.Itoa(counter.Count()), "\n")
		}

		fmt.Fprintln(v, message)
		return nil
	})
}

func updateLogView(gui *gocui.Gui, logEvent string) {
	gui.Execute(func(g *gocui.Gui) error {
		v, err := g.View("logs")
		if err != nil {
			return err
		}
		fmt.Fprintln(v, logEvent)
		return nil
	})
}

func updateAveragesView(gui *gocui.Gui, timeseries ts.TimeSeries) {
	gui.Execute(func(g *gocui.Gui) error {
		v, err := g.View("averages")
		if err != nil {
			return err
		}
		v.Clear()

		message := ""

		message += fmt.Sprint(" Avg Hits: ", timeseries.AverageHits(12))
		message += fmt.Sprint("  Avg Bytes: ", timeseries.AverageBytes(12))

		fmt.Fprintln(v, message)
		return nil
	})
}

func updateTotalView(gui *gocui.Gui, timeseries ts.TimeSeries) {
	gui.Execute(func(g *gocui.Gui) error {
		v, err := g.View("totals")
		if err != nil {
			return err
		}
		v.Clear()

		message := ""

		message += fmt.Sprint(" Total Hits: ", timeseries.TotalHits())
		message += fmt.Sprint("  Total Bytes: ", timeseries.TotalBytes())

		fmt.Fprintln(v, message)
		return nil
	})
}

func updateAlertView(gui *gocui.Gui, message string) {
	gui.Execute(func(g *gocui.Gui) error {
		v, err := g.View("alerts")
		if err != nil {
			return err
		}
		fmt.Fprintln(v, message)
		return nil
	})
}
