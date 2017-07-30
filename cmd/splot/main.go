package main

import "fmt"
import (
	"bufio"
	"github.com/vikram-r/splot"
	"os"
	"os/exec"
	"runtime/debug"
	"strconv"
	"strings"
)

func main() {
	// TODO error handling should not be done using panics
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(os.Stderr, "error: ", r)
			fmt.Printf("trace: %s", debug.Stack())
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

	tWidth, tHeight, err := terminalDimensions()
	if err != nil {
		panic("Could not read terminal dimensions")
	}

	plot, err := splot.NewPlot(bufio.NewReader(file))
	if err != nil {
		panic(err)
	}

	// TODO make this configurable via flag
	plot.SetColors(splot.ColorConfig{
		Point:      "\033[93m",
		Line:       "\033[92m",
		XAxis:      "\033[34m",
		YAxis:      "\033[34m",
		XAxisTitle: "\033[96m",
		YAxisTitle: "\033[96m",
		Tick:       "",
		TickLabel:  "\033[95m",
	})

	// TODO this blows up if the terminal width or height is <=10
	plot.SetNumXTicks(10)
	plot.SetNumYTicks(10)

	if err := plot.Render(os.Stdout, tWidth, tHeight); err != nil {
		panic(err)
	}
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
