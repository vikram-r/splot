# splot

splot (simple plot) allows you to render ascii graphs from a list of coordinates. This project comes with a package that can be added to your own projects, and a CLI that can be used out of the box.

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

Then simply import
```
import "github.com/vikram-r/splot"
```

and call either `Render` or `RenderWithColor`.

## TODO
- The line drawing between points is finicky.
- Provide a way to pipe data in (will need to provide terminal dimensions manually in this case)
- Add splot binary release
- Improve splot package interface, and update README
- splot package should support configurable output destination (not just stdout)
