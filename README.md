# splot

splot (simple plot) allows you to render ascii graphs from a list of coordinates. This project comes with a fully configurable package that can be added to your own projects, and a CLI that can be used out of the box.

<img src="http://i.imgur.com/TsI1jOz.jpg" width="60%">

## Usage (CLI)

Build the executable
```
go build ./cmd/splot/
```

The executable takes a file path to a list of coordinates as a parameter
```
./splot <file_path>
```

The file should be a csv containing x,y coordinates, with the first row defining the axis titles
```csv
x axis,y axis
0,0
1,25
2,50
3,25
4,0
```

Result:

<img src="http://i.imgur.com/GBlQ5Or.jpg" width="60%">

## Usage (Package)

```
go get github.com/vikram-r/splot
```

Provide an `io.Reader` to your data source:
```go
plot, err := splot.NewPlot(bytes.NewReader([]byte(
`x axis,y axis
0,0
1,25
2,50
3,25
4,0`)))
```
Then `Render` the `Plot` using your provided `io.Writer`:
```go
err := plot.Render(os.Stdout, tWidth, tHeight)
```

Optionally configure colors by providing ANSI color codes:
```go
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
```
You can also change the number of ticks per axis:
```go
plot.SetNumXTicks(10)
plot.SetNumYTicks(10)
```

## TODO
- The line drawing between points is finicky.
- Provide a way to pipe data in (will need to provide terminal dimensions manually in this case)
- Add splot binary release
- Improve splot package interface, and update README
- splot package should support configurable output destination (not just stdout)
