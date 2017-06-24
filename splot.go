package main

import "fmt"
import (
	"bufio"
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

type DataSet []Point

func (ds DataSet) sort() {
	sort.Slice(ds, func(i, j int) bool { return ds[i].x < ds[j].x })
}

func (ds DataSet) XRange() (int, int) {
	return ds[0].x, ds[len(ds)-1].x
}

func (ds DataSet) YRange() (int, int) {
	return ds[0].y, ds[len(ds)-1].y
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

func loadData(input io.Reader) (DataSet, error) {
	scanner := bufio.NewScanner(input)

	scanner.Scan()
	header := strings.Split(scanner.Text(), ",")
	xAxis := header[0]
	yAxis := header[1]

	fmt.Println(xAxis, yAxis)

	var data DataSet
	rowNum := 1
	for scanner.Scan() {
		p, err := parseRow(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("Could not parse row: %s, reason: %q", strconv.Itoa(rowNum), err)
		}
		fmt.Println(p.x, p.y)

		data = append(data, p)
		rowNum++
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	data.sort()
	return data, nil
}

func parseRow(row string) (Point, error) {
	r := strings.Split(row, ",")

	x, err := strconv.Atoi(r[0])
	if err != nil {
		return Point{}, err
	}

	y, err := strconv.Atoi(r[1])
	if err != nil {
		return Point{}, err
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
