package cui

import (
	"fmt"
	"os"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/wallywest/lmchttp/ts"
)

func New(cuiChan chan ts.TimeSeries, logDumpChan chan string, alertChan chan string) *gocui.Gui {
	gui := gocui.NewGui()

	if err := gui.Init(); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	gui.SetLayout(layout)

	err := keybindings(gui)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	t := time.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case timeSeries := <-cuiChan:
				go updateSectionView(gui, timeSeries)
				go updateAveragesView(gui, timeSeries)
				go updateTotalView(gui, timeSeries)
			case logDump := <-logDumpChan:
				go updateLogView(gui, logDump)
			case alertMessage := <-alertChan:
				go updateAlertView(gui, alertMessage)
			case <-t.C:
			}
		}
	}()

	return gui
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	mainView, err := g.SetView("sections", 0, 3, maxX/3, maxY-8)

	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}

	mainView.Autoscroll = true
	mainView.Frame = true
	mainView.Title = "Top Hit Sections"

	g.SetCurrentView("sections")

	if averages_view, err := g.SetView("averages", 0, 0, maxX-(2*maxX/3), 2); err != nil &&
		err != gocui.ErrUnknownView {
		return err
	} else {
		averages_view.Frame = true
		averages_view.FgColor = gocui.ColorBlue
	}

	if total_view, err := g.SetView("totals", maxX/3, 0, maxX-(maxX/3), 2); err != nil &&
		err != gocui.ErrUnknownView {
		return err
	} else {
		total_view.Frame = true
	}

	if time_view, err := g.SetView("info", (2 * (maxX / 3)), 0, maxX-1, 2); err != nil &&
		err != gocui.ErrUnknownView {
		return err
	} else {
		time_view.Frame = true
		options := fmt.Sprint(" Refresh Interval: 10s  Alert Threshold: 100  Log File: access_log")
		fmt.Fprintln(time_view, options)
	}

	alerts_view, err := g.SetView("alerts", maxX/3, 3, maxX-1, maxY-8)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	} else {
		alerts_view.Frame = true
		alerts_view.Title = "Threshold Alerts"
		alerts_view.Autoscroll = true
		alerts_view.Highlight = false
		alerts_view.Wrap = true
	}

	return nil
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	return nil
}
