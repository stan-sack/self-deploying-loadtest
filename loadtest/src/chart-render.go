package main

type gnuChartRenderer struct {
	parser LogParser
}

type ChartRenderer interface {
	Start()
}

func makeChartRenderer(parser LogParser) ChartRenderer {
	return &gnuChartRenderer{
		parser: parser,
	}
}
