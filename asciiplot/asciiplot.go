package asciiplot

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

type canvas struct {
	board         [][]rune
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
}

func newCanvas(board [][]rune, dataSet *dataSet, width int, height int, xNumTicks int, yNumTicks int) *canvas {
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
	}

	canvas.drawXAxis()
	canvas.drawYAxis()
	for i, p := range dataSet.data {
		canvas.drawPoint(p, '*')
		if i != 0 {
			// draw line from previous point
			canvas.drawLine(dataSet.data[i-1], p, '.')
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

			c.board[c.height-xAxisTitleOffset][j] = r
		}

		// draw tick
		if s, ok := tickIndices[j]; ok {
			label := []rune(s)
			labelBegin := j - len(label)
			for lidx := 0; lidx < len(label); lidx++ {
				c.board[c.height-xAxisTickLabelOffset][labelBegin+lidx] = label[lidx]
			}
			c.board[c.height-xAxisOffset][j] = '+'
		} else {
			// draw axis
			c.board[c.graphHeight][j] = '_'
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

			c.board[i][yAxisTitleOffset] = r
		}

		// draw tick
		if s, ok := tickIndices[i]; ok {
			label := []rune(s)
			for lidx := 0; lidx < len(label); lidx++ {
				c.board[i][yAxisTickLabelOffset+lidx] = label[lidx]
			}
			c.board[i][yAxisOffset] = '+'
		} else {
			// draw axis
			c.board[i][yAxisOffset] = '|'
		}

	}
}

func (c *canvas) getExactCoordinate(p point) (width int64, height int64) {
	width = int64(yAxisOffset) + int64(float64(p.x)*c.xRatio)
	height = int64(c.height) - int64(xAxisOffset+(float64(p.y)*c.yRatio))
	return width, height
}

func (c *canvas) drawPoint(p point, char rune) {
	w, h := c.getExactCoordinate(p)
	c.drawExactPoint(w, h, char)
}

func (c *canvas) drawExactPoint(w int64, h int64, char rune) {
	c.board[h][w] = char
}

func (c *canvas) drawLine(from point, to point, char rune) {
	fromW, fromH := c.getExactCoordinate(from)
	toW, _ := c.getExactCoordinate(to)

	slope := float64(to.y-from.y) / float64(to.x-from.x)

	xScale := 1 / c.xRatio

	for i := int64(1); i < toW-fromW; i++ {
		w := i + fromW
		h := fromH - int64(((float64(i)*xScale)*slope)+.5)
		c.drawExactPoint(w, h, char)
	}
}

func (c *canvas) render(writer io.Writer) {
	for i := 0; i < c.height; i++ {
		for j := range c.board[i] {
			if c.board[i][j] == 0 {
				fmt.Fprint(writer, " ")
			} else {
				fmt.Fprint(writer, string(c.board[i][j]))
			}
		}
		fmt.Fprintln(writer)
	}
}

type RowError struct {
	rowNum  int
	rowText string
	reason  error
}

func (r *RowError) Error() string {
	return fmt.Sprintf("Could not parse row %s: \"%s\", reason: %q", strconv.Itoa(r.rowNum), r.rowText, r.reason)
}

func Render(dataSource io.Reader, to io.Writer, width int, height int, xNumTicks, yNumTicks int) error {
	ds, err := loadData(dataSource)
	if err != nil {
		return err
	}

	b := make([][]rune, height)
	for i := range b {
		b[i] = make([]rune, width)
	}

	ds.sort()

	canvas := newCanvas(b, ds, width, height, xNumTicks, yNumTicks)

	canvas.render(to)

	return nil
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
