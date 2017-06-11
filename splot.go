package main

import "fmt"
import (
	"bufio"
	"os"
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

	ds := loadData()

	fmt.Println(ds.XRange())
	fmt.Println(ds.YRange())

	fmt.Println("done!")
}

func loadData() DataSet {
	scanner := bufio.NewScanner(os.Stdin)

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
			panic("Could not parse row " + strconv.Itoa(rowNum))
		}
		fmt.Println(p.x, p.y)

		data = append(data, p)
		rowNum++
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	data.sort()
	return data
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
