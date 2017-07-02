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

func (ds *dataSet) xRange() (int, int) {
	return ds.data[0].x, ds.data[len(ds.data)-1].x
}

func (ds *dataSet) yRange() (int, int) {
	return ds.data[0].y, ds.data[len(ds.data)-1].y
}

// TODO tickIntervalX, if required

func (ds *dataSet) tickIntervalY(numTicks int) float64 {
	return prettyInterval(ds.yMin, ds.yMax, numTicks)
}

// prettyInterval calculates a visually appealing interval per tick given the minimum value on the axis, maximum value
// on the axis and the number of ticks required.
// More information: https://stackoverflow.com/questions/326679/choosing-an-attractive-linear-scale-for-a-graphs-y-axis
func prettyInterval(min int, max int, numTicks int) float64 {
	trueInterval := float64(max-min) / float64(numTicks)
	factor := math.Pow(10, math.Ceil(math.Log10(trueInterval)-1))
	return math.Ceil(trueInterval/factor) * factor
}

type canvas struct {
	board     [][]rune
	data      *dataSet
	width     int
	height    int
	numTicksY int
}

const xAxisTitleOffset = 2
const xAxisTickLabelOffset = xAxisTitleOffset + 2
const xAxisOffset = xAxisTickLabelOffset + 1

func (c *canvas) graphHeight() int {
	return c.height - xAxisOffset
}

func (c *canvas) graphWidth() int {
	return c.width - yAxisOffset
}

func (c *canvas) drawXAxis() {
	midWidth := int(math.Floor(float64(c.graphWidth() / 2)))
	xLabelMiddle := int(math.Floor(float64(len(c.data.xName) / 2)))

	for j := yAxisOffset; j < c.width; j++ {
		if j >= midWidth-xLabelMiddle && j < midWidth+xLabelMiddle {
			charIndex := j - (midWidth - xLabelMiddle)
			r, _ := utf8.DecodeRune([]byte{c.data.xName[charIndex]})

			c.board[c.height-xAxisTitleOffset][j] = r

		}
		c.board[c.graphHeight()][j] = '_'
	}
}

const yAxisTitleOffset = 1
const yAxisTickLabelOffset = yAxisTitleOffset + 4
const yAxisOffset = yAxisTickLabelOffset + 2

func (c *canvas) drawYAxis() {
	tick := c.data.tickIntervalY(c.numTicksY)
	lowerBound := tick * float64(int64((float64(c.data.yMin)/tick)+.5))
	upperBound := tick * float64(int64(1+(float64(c.data.yMax)/tick)+.5))
	fmt.Println("Pretty Interval: ", tick)
	fmt.Println("Bounds: ", lowerBound, ", ", upperBound)

	midHeight := int(math.Floor(float64(c.graphHeight() / 2)))
	yLabelMiddle := int(math.Floor(float64(len(c.data.yName) / 2)))

	graphHeight := c.graphHeight()
	graphTick := graphHeight / c.numTicksY
	tickIndices := map[int]bool{}
	// TODO should 0 be considered a tick?
	for t := 1; t <= c.numTicksY; t++ {
		tickIndices[graphHeight-(graphTick*t)] = true
	}

	for i := 0; i < graphHeight; i++ {
		// draw axis title
		if i >= midHeight-yLabelMiddle && i < midHeight+yLabelMiddle {
			charIndex := i - (midHeight - yLabelMiddle)
			r, _ := utf8.DecodeRune([]byte{c.data.yName[charIndex]})

			c.board[i][yAxisTitleOffset] = r
		}

		// draw tick
		if _, ok := tickIndices[i]; ok {
			c.board[i][yAxisOffset] = '+'
		} else {
			// draw axis
			c.board[i][yAxisOffset] = '|'
		}

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

func Render(dataSource io.Reader, to io.Writer, width int, height int, numTicksY int) error {
	ds, err := loadData(dataSource)
	if err != nil {
		return err
	}

	b := make([][]rune, height)
	for i := range b {
		b[i] = make([]rune, width)
	}

	canvas := canvas{
		board:     b,
		data:      ds,
		width:     width,
		height:    height,
		numTicksY: numTicksY,
	}

	canvas.drawXAxis()
	canvas.drawYAxis()
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
	dataSet.sort()

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
