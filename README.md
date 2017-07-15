# splot

splot (simple plot) is a command line utility to render an ascii graph from a list of coordinates.

## Usage

Build the executable
```
go build
```

The executable takes a file path to a list of coordinates as a parameter
```
./splot <file_path>
```

The file should be a csv containing x,y coordinates, with the first row defining the axis titles
```csv
x axis,y axis
0,0
1,1
2,2
```

Result:

<img src="http://i.imgur.com/uYp8mMO.jpg" width="60%">

## TODO
- The line drawing between points is finicky.
- Refactor to make the asciigraph package easy to use (outside of the CLI tool).
- Provide a way to pipe data in (will need to provide terminal dimensions manually in this case)
