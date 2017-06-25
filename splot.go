package main

import "fmt"
import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

type Point struct {
	x, y int
}

type DataSet struct {
	data  []Point
	xName string
	yName string
}

func (ds DataSet) sort() {
	sort.Slice(ds.data, func(i, j int) bool { return ds.data[i].x < ds.data[j].x })
}

func (ds DataSet) XRange() (int, int) {
	return ds.data[0].x, ds.data[len(ds.data)-1].x
}

func (ds DataSet) YRange() (int, int) {
	return ds.data[0].y, ds.data[len(ds.data)-1].y
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(os.Stderr, "error: ", r)
			os.Exit(1)
		}
	}()
	if len(os.Args) < 2 {
		panic("not enough arguments")
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer file.Close()

	ds, err := loadData(bufio.NewReader(file))
	if err != nil {
		panic(err)
	}

	ds.XRange()

	tWidth, tHeight, err := terminalDimensions()
	if err != nil {
		panic("Could not read terminal dimensions")
	}

	fmt.Printf("Terminal (w, h): (%d, %d)", tWidth, tHeight)
}

func loadData(input io.Reader) (*DataSet, error) {
	scanner := bufio.NewScanner(input)

	if !scanner.Scan() {
		return nil, errors.New("No data found")
	}
	header := strings.Split(scanner.Text(), ",")
	if len(header) < 2 || header[0] == "" || header[1] == "" {
		return nil, errors.New("Header with 2 elements required")
	}
	xAxis := header[0]
	yAxis := header[1]

	fmt.Println(xAxis, yAxis)

	dataSet := DataSet{
		xName: xAxis,
		yName: yAxis,
		data:  []Point{},
	}
	rowNum := 1
	for scanner.Scan() {
		row := scanner.Text()
		p, err := parseRow(row)
		if err != nil {
			return nil, fmt.Errorf("Could not parse row %s: \"%s\", reason: %q", strconv.Itoa(rowNum), row, err)
		}
		fmt.Println(p.x, p.y)

		dataSet.data = append(dataSet.data, p)
		rowNum++
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	dataSet.sort()

	return &dataSet, nil
}

func parseRow(row string) (Point, error) {
	r := strings.Split(row, ",")
	if len(r) < 2 {
		return Point{}, errors.New("coordinates require 2 values")
	}

	x, err := strconv.Atoi(r[0])
	if err != nil {
		return Point{}, fmt.Errorf("%v is not a number", r[0])
	}

	y, err := strconv.Atoi(r[1])
	if err != nil {
		return Point{}, fmt.Errorf("%v is not a number", r[1])
	}

	return Point{x, y}, nil
}

func terminalDimensions() (width, height int, e error) {
	stty := exec.Command("stty", "size")
	stty.Stdin = os.Stdin
	if r, err := stty.Output(); err != nil {
		return 0, 0, err
	} else {
		//h w\n
		p := strings.Split(strings.TrimSpace(string(r)), " ")

		if height, err = strconv.Atoi(p[0]); err != nil {
			return 0, 0, err
		}
		if width, err = strconv.Atoi(p[1]); err != nil {
			return 0, 0, err
		}
		return width, height, nil
	}
}
