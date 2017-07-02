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

func (ds dataSet) sort() {
	sort.Slice(ds.data, func(i, j int) bool { return ds.data[i].x < ds.data[j].x })
}

func (ds dataSet) xRange() (int, int) {
	return ds.data[0].x, ds.data[len(ds.data)-1].x
}

func (ds dataSet) yRange() (int, int) {
	return ds.data[0].y, ds.data[len(ds.data)-1].y
}

type canvas struct {
	board  [][]rune
	data   *dataSet
	width  int
	height int
}

type RowError struct {
	rowNum  int
	rowText string
	reason  error
}

func (r *RowError) Error() string {
	return fmt.Sprintf("Could not parse row %s: \"%s\", reason: %q", strconv.Itoa(r.rowNum), r.rowText, r.reason)
}

func Render(dataSource io.Reader, to io.Writer, width int, height int, numTicksX int, numTicksY int) error {
	ds, err := loadData(dataSource)
	if err != nil {
		return err
	}

	b := make([][]rune, height)
	for i := range b {
		b[i] = make([]rune, width)
	}

	canvas := canvas{
		board:  b,
		data:   ds,
		width:  width,
		height: height,
	}

	setAxes(canvas)
	renderCanvas(canvas, to)

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

const axisOffset = 6
const axisLabelOffset = axisOffset - 2
const axisTitleOffset = axisLabelOffset - 2

func setAxes(canvas canvas) {
	midHeight := int(math.Floor(float64(canvas.height / 2)))
	yLabelMiddle := int(math.Floor(float64(len(canvas.data.yName) / 2)))
	for i := 0; i < canvas.height; i++ {
		if i >= midHeight-yLabelMiddle && i <= midHeight+yLabelMiddle-1 {
			currOffset := i - (midHeight - yLabelMiddle)
			r, _ := utf8.DecodeRune([]byte{canvas.data.yName[currOffset]})

			canvas.board[i][axisTitleOffset] = r
		}
		canvas.board[i][axisOffset] = '|'
	}

	midWidth := int(math.Floor(float64(canvas.width / 2)))
	xLabelMiddle := int(math.Floor(float64(len(canvas.data.xName) / 2)))
	for j := 0; j < canvas.width; j++ {
		if j >= midWidth-xLabelMiddle && j <= midWidth+xLabelMiddle-1 {
			currOffset := j - (midWidth - xLabelMiddle)
			r, _ := utf8.DecodeRune([]byte{canvas.data.xName[currOffset]})

			canvas.board[canvas.height-axisTitleOffset][j] = r

		}
		canvas.board[canvas.height-axisOffset+1][j] = '_'
	}
}

func renderCanvas(canvas canvas, writer io.Writer) {
	for i := 0; i < canvas.height; i++ {
		for j := range canvas.board[i] {
			if canvas.board[i][j] == 0 {
				fmt.Fprint(writer, " ")
			} else {
				fmt.Fprint(writer, string(canvas.board[i][j]))
			}
		}
		fmt.Fprintln(writer)
	}
}
