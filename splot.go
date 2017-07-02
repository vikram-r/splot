package main

import "fmt"
import (
	"bufio"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"github.com/vikram-r/splot/asciiplot"
)

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

	tWidth, tHeight, err := terminalDimensions()
	if err != nil {
		panic("Could not read terminal dimensions")
	}

	fmt.Printf("Terminal (w, h): (%d, %d)", tWidth, tHeight)

	err = asciiplot.Render(bufio.NewReader(file), os.Stdout, tWidth, tHeight, 10, 10)
	if err != nil {
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
