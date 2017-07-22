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
1,25
2,50
3,75
4,100
5,75
6,50
7,25
8,25
9,25
10,30
11,40
12,20
13,35
14,50
15,60
16,30
17,80
```

Result:

<img src="http://i.imgur.com/uN4jB4r.jpg" width="60%">

## TODO
- The line drawing between points is finicky.
- Refactor to make the asciigraph package easy to use (outside of the CLI tool).
- Provide a way to pipe data in (will need to provide terminal dimensions manually in this case)
