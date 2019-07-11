package main

import (
	"time"

	"github.com/sbinet/go-gnuplot"
)

func (cr *gnuChartRenderer) Start() {
	fname := ""
	persist := false
	debug := false

	p, err := gnuplot.NewPlotter(fname, persist, debug)
	if err != nil {
		panic(err)
	}
	defer p.Close()

	for {
		time.Sleep(time.Second * 5)
		toPlot := []float64{}
		for _, res := range *cr.parser.GetResults() {
			toPlot = append(toPlot, float64(res.totalDurationMillis))
		}

		if len(toPlot) > 0 {
			p.ResetPlot()
			p.PlotX(toPlot, "Req duration (ms)")

		}
	}
}
