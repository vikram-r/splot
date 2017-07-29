package splot

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

type point struct {
	x, y int
}

type dataSet struct {
	data                   []point
	xName, yName           string
	xMin, xMax, yMin, yMax int
}

func (ds *dataSet) sort() {
	sort.Slice(ds.data, func(i, j int) bool { return ds.data[i].x < ds.data[j].x })
}

func (ds *dataSet) xTickInterval(numTicks int) float64 {
	return prettyInterval(ds.xMin, ds.xMax, numTicks)
}

func (ds *dataSet) yTickInterval(numTicks int) float64 {
	return prettyInterval(ds.yMin, ds.yMax, numTicks)
}

// prettyInterval calculates a visually appealing interval per tick given the minimum value on the axis, maximum value
// on the axis and the number of ticks required.
// More information: https://stackoverflow.com/questions/326679/choosing-an-attractive-linear-scale-for-a-graphs-y-axis
func prettyInterval(min int, max int, numTicks int) float64 {
	var trueInterval float64
	if min == max {
		trueInterval = float64(min)
	} else {
		trueInterval = float64(max) / float64(numTicks)
	}
	factor := math.Pow(10, math.Ceil(math.Log10(trueInterval)-1))
	return math.Ceil(trueInterval/factor) * factor
}

type pixel struct {
	char          rune
	ansiColorCode string
}

type canvas struct {
	board         [][]*pixel // [h,w]
	data          *dataSet
	height        int
	width         int
	xNumTicks     int
	yNumTicks     int
	xTickInterval float64
	yTickInterval float64
	xRatio        float64
	yRatio        float64
	graphHeight   int
	graphWidth    int
	colorConfig   ColorConfig
}

type ColorConfig struct {
	Point      string
	Line       string
	XAxis      string
	YAxis      string
	XAxisTitle string
	YAxisTitle string
	Tick       string
	TickLabel  string
}

func newCanvas(board [][]*pixel, dataSet *dataSet, width int, height int, xNumTicks int, yNumTicks int, colorConfig ColorConfig) *canvas {
	graphHeight := height - xAxisOffset
	graphWidth := width - yAxisOffset - 1

	xTickInterval := dataSet.xTickInterval(xNumTicks)
	yTickInterval := dataSet.yTickInterval(yNumTicks)

	xRatio := float64(graphWidth/xNumTicks) / xTickInterval
	yRatio := float64(graphHeight/yNumTicks) / yTickInterval

	canvas := canvas{
		board:         board,
		data:          dataSet,
		width:         width,
		height:        height,
		xNumTicks:     xNumTicks,
		yNumTicks:     yNumTicks,
		xTickInterval: xTickInterval,
		yTickInterval: yTickInterval,
		xRatio:        xRatio,
		yRatio:        yRatio,
		graphHeight:   graphHeight,
		graphWidth:    graphWidth,
		colorConfig: colorConfig,
	}

	canvas.drawXAxis()
	canvas.drawYAxis()
	for i, p := range dataSet.data {
		canvas.addPointPixels(p, 'X')
		if i != 0 {
			// draw line from previous point
			canvas.addLinePixels(dataSet.data[i-1], p, '.')
		}
	}

	return &canvas
}

// prev offset + prev height + actual offset
const xAxisTitleOffset = 1
const xAxisTickLabelOffset = xAxisTitleOffset + 1 + 1
const xAxisOffset = xAxisTickLabelOffset + 1 + 0

func (c *canvas) drawXAxis() {
	midWidth := int(math.Floor(float64(c.graphWidth / 2)))
	xLabelMiddle := int(math.Floor(float64(len(c.data.xName) / 2)))

	graphTick := c.graphWidth / c.xNumTicks
	tickIndices := map[int]string{}
	for t := 1; t <= c.xNumTicks; t++ {
		tickIndices[yAxisOffset+(graphTick*t)] = strconv.FormatFloat(float64(t)*c.xTickInterval, 'f', 1, 64)
	}

	for j := yAxisOffset; j < c.width; j++ {
		// draw axis title
		if j >= midWidth-xLabelMiddle && j < midWidth+xLabelMiddle {
			charIndex := j - (midWidth - xLabelMiddle)
			r, _ := utf8.DecodeRune([]byte{c.data.xName[charIndex]})

			c.addXAxisTitlePixel(j, c.height-xAxisTitleOffset, r)
		}

		// draw tick
		if s, ok := tickIndices[j]; ok {
			label := []rune(s)
			labelBegin := j - len(label)
			for lidx := 0; lidx < len(label); lidx++ {
				c.addTickLablePixel(labelBegin+lidx, c.height-xAxisTickLabelOffset, label[lidx])
			}
			c.addTickPixel(j, c.height-xAxisOffset, '+')
		} else {
			// draw axis
			c.addXAxisPixel(j, c.graphHeight, '_')
		}
	}
}

// prev offset + prev width + actual offset
const yAxisTitleOffset = 1
const yAxisTickLabelOffset = yAxisTitleOffset + 1 + 2
const yAxisOffset = yAxisTickLabelOffset + 4 + 2

func (c *canvas) drawYAxis() {
	midHeight := int(math.Floor(float64(c.graphHeight / 2)))
	yLabelMiddle := int(math.Floor(float64(len(c.data.yName) / 2)))

	graphTick := c.graphHeight / c.yNumTicks
	tickIndices := map[int]string{}
	// TODO should 0 be considered a tick?
	for t := 1; t <= c.yNumTicks; t++ {
		tickIndices[c.graphHeight-(graphTick*t)] = strconv.FormatFloat(float64(t)*c.yTickInterval, 'f', 1, 64)
	}

	for i := 0; i < c.graphHeight; i++ {
		// draw axis title
		if i >= midHeight-yLabelMiddle && i < midHeight+yLabelMiddle {
			charIndex := i - (midHeight - yLabelMiddle)
			r, _ := utf8.DecodeRune([]byte{c.data.yName[charIndex]})

			c.addYAxisTitlePixel(yAxisTitleOffset, i, r)
		}

		// draw tick
		if s, ok := tickIndices[i]; ok {
			label := []rune(s)
			for lidx := 0; lidx < len(label); lidx++ {
				c.addTickLablePixel(yAxisTickLabelOffset+lidx, i, label[lidx])
			}
			c.addTickPixel(yAxisOffset, i, '+')
		} else {
			// draw axis
			c.addYAxisPixel(yAxisOffset, i, '|')
		}

	}
}

func (c *canvas) getExactCoordinate(p point) (width int64, height int64) {
	width = int64(yAxisOffset) + int64(float64(p.x)*c.xRatio)
	height = int64(c.height) - int64(xAxisOffset+(float64(p.y)*c.yRatio))
	return width, height
}

func (c *canvas) addPointPixels(p point, char rune) {
	w, h := c.getExactCoordinate(p)
	c.addPixel(w, h, pixel{char: char, ansiColorCode: c.colorConfig.Point})
}

func (c *canvas) addLinePixels(from point, to point, char rune) {
	fromW, fromH := c.getExactCoordinate(from)
	toW, toH := c.getExactCoordinate(to)

	deltaY := float64((toH-fromH)*-1) / float64(toW-fromW)

	for i := int64(1); i < toW-fromW; i++ {
		w := i + fromW
		h := int64(float64(fromH) - (float64(i) * deltaY))

		c.addPixel(w, h, pixel{char: char, ansiColorCode: c.colorConfig.Line})
	}
}

func (c *canvas) addXAxisPixel(w int, h int, char rune) {
	c.addPixel(int64(w), int64(h), pixel{char: char, ansiColorCode: c.colorConfig.XAxis})
}

func (c *canvas) addYAxisPixel(w int, h int, char rune) {
	c.addPixel(int64(w), int64(h), pixel{char: char, ansiColorCode: c.colorConfig.YAxis})
}

func (c *canvas) addXAxisTitlePixel(w int, h int, char rune) {
	c.addPixel(int64(w), int64(h), pixel{char: char, ansiColorCode: c.colorConfig.XAxisTitle})
}

func (c *canvas) addYAxisTitlePixel(w int, h int, char rune) {
	c.addPixel(int64(w), int64(h), pixel{char: char, ansiColorCode: c.colorConfig.YAxisTitle})
}

func (c *canvas) addTickPixel(w int, h int, char rune) {
	c.addPixel(int64(w), int64(h), pixel{char: char, ansiColorCode: c.colorConfig.Tick})
}

func (c *canvas) addTickLablePixel(w int, h int, char rune) {
	c.addPixel(int64(w), int64(h), pixel{char: char, ansiColorCode: c.colorConfig.TickLabel})
}

func (c *canvas) addPixel(w int64, h int64, p pixel) {
	c.board[h][w] = &p
}

func (c *canvas) render(writer io.Writer) {
	for i := 0; i < c.height; i++ {
		for j := range c.board[i] {
			p := c.board[i][j]
			if p == nil {
				fmt.Fprint(writer, " ")
			} else if p.ansiColorCode != "" {
				fmt.Fprint(writer, colorString(p.ansiColorCode, string(p.char)))
			} else {
				fmt.Fprint(writer, string(p.char))
			}
		}
		fmt.Fprintln(writer)
	}
}

const ansi_color_off = "\033[0m"

func colorString(ansiCode string, text string) string {
	return fmt.Sprintf("%s%s%s", ansiCode, text, ansi_color_off)
}

type RowError struct {
	rowNum  int
	rowText string
	reason  error
}

func (r *RowError) Error() string {
	return fmt.Sprintf("Could not parse row %s: \"%s\", reason: %q", strconv.Itoa(r.rowNum), r.rowText, r.reason)
}

func RenderWithColor(dataSource io.Reader, to io.Writer, width int, height int, xNumTicks, yNumTicks int, colorConfig ColorConfig) error {
	ds, err := loadData(dataSource)
	if err != nil {
		return err
	}

	b := make([][]*pixel, height)
	for i := range b {
		b[i] = make([]*pixel, width)
	}

	ds.sort()

	canvas := newCanvas(b, ds, width, height, xNumTicks, yNumTicks, colorConfig)

	canvas.render(to)

	return nil
}

func Render(dataSource io.Reader, to io.Writer, width int, height int, xNumTicks, yNumTicks int) error {
	return RenderWithColor(dataSource, to, width, height, xNumTicks, yNumTicks, ColorConfig{})
}

func loadData(input io.Reader) (*dataSet, error) {
	scanner := bufio.NewScanner(input)

	if !scanner.Scan() {
		return nil, &RowError{0, "", errors.New("No data found")}
	}
	header := strings.Split(scanner.Text(), ",")
	if len(header) < 2 || header[0] == "" || header[1] == "" {
		return nil, &RowError{0, strings.Join(header, ","), errors.New("Header with 2 elements required")}
	}
	xAxis := header[0]
	yAxis := header[1]

	dataSet := dataSet{
		xName: xAxis,
		yName: yAxis,
		data:  []point{},
	}

	var xMin, xMax, yMin, yMax int
	rowNum := 1
	for scanner.Scan() {
		row := scanner.Text()
		p, err := parseRow(row)
		if err != nil {
			return nil, &RowError{rowNum, row, err}
		}

		if rowNum == 1 {
			xMin, xMax, yMin, yMax = p.x, p.x, p.y, p.y
		}

		dataSet.data = append(dataSet.data, p)

		if p.x < xMin {
			xMin = p.x
		}
		if p.x > xMax {
			xMax = p.x
		}
		if p.y < yMin {
			yMin = p.y
		}
		if p.y > yMax {
			yMax = p.y
		}
		rowNum++
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	dataSet.xMin = xMin
	dataSet.xMax = xMax
	dataSet.yMin = yMin
	dataSet.yMax = yMax

	return &dataSet, nil
}

func parseRow(row string) (point, error) {
	r := strings.Split(row, ",")
	if len(r) < 2 {
		return point{}, errors.New("coordinates require 2 values")
	}

	x, err := strconv.Atoi(r[0])
	if err != nil {
		return point{}, fmt.Errorf("%v is not a number", r[0])
	}

	y, err := strconv.Atoi(r[1])
	if err != nil {
		return point{}, fmt.Errorf("%v is not a number", r[1])
	}

	return point{x, y}, nil
}
