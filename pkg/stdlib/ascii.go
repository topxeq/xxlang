// pkg/stdlib/ascii.go
// ASCII plotting module for the Xxlang standard library.
package stdlib

import (
	"fmt"
	"math"
	"strings"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "ascii",
		Exports: map[string]objects.Object{
			"plotDataToStr":    BuiltinFunc(asciiPlotDataToStr),
			"plotClearConsole": BuiltinFunc(asciiPlotClearConsole),
			"plotMoveCursor":   BuiltinFunc(asciiPlotMoveCursor),
			"plotConsoleSize":  BuiltinFunc(asciiPlotConsoleSize),
		},
	})
}

type plotConfig struct {
	width        int
	height       int
	minVal       float64
	maxVal       float64
	offset       int
	precision    int
	caption      string
	captionColor int
	axisColor    int
	labelColor   int
	seriesColors []int
}

func defaultConfig() plotConfig {
	return plotConfig{
		width: 0, height: 7, minVal: math.NaN(), maxVal: math.NaN(),
		offset: 5, precision: 2,
	}
}

// asciiPlotClearConsole - Clear the console screen using ANSI escape codes
// Usage: plotClearConsole()
// Returns null
func asciiPlotClearConsole(args ...objects.Object) objects.Object {
	if len(args) != 0 {
		return &objects.Error{Message: "plotClearConsole takes no arguments"}
	}
	// ANSI escape codes: clear screen and move cursor to top-left
	fmt.Print("\x1b[2J\x1b[H")
	return objects.NULL
}

// asciiPlotMoveCursor - Move cursor to specified position
// Usage: plotMoveCursor(row, col)
// Returns null
func asciiPlotMoveCursor(args ...objects.Object) objects.Object {
	if len(args) != 2 {
		return &objects.Error{Message: "plotMoveCursor takes 2 arguments: row, col"}
	}

	row, ok := args[0].(*objects.Int)
	if !ok {
		return &objects.Error{Message: "first argument to 'plotMoveCursor' must be INT"}
	}

	col, ok := args[1].(*objects.Int)
	if !ok {
		return &objects.Error{Message: "second argument to 'plotMoveCursor' must be INT"}
	}

	// ANSI escape code: move cursor to row, col (1-indexed)
	fmt.Printf("\x1b[%d;%dH", row.Value+1, col.Value+1)
	return objects.NULL
}

// asciiPlotConsoleSize - Get console size
// Usage: size = plotConsoleSize()
// Returns [width, height]
func asciiPlotConsoleSize(args ...objects.Object) objects.Object {
	if len(args) != 0 {
		return &objects.Error{Message: "plotConsoleSize takes no arguments"}
	}

	// Try to get console size using ANSI escape codes
	// This is a simplified implementation that returns default values
	// For a full implementation, platform-specific code would be needed
	width, height := 80, 24 // Default terminal size

	result := &objects.Array{Elements: []objects.Object{
		&objects.Int{Value: int64(width)},
		&objects.Int{Value: int64(height)},
	}}
	return result
}

func parsePlotOptions(args []objects.Object) plotConfig {
	cfg := defaultConfig()
	for i := 1; i < len(args); i++ {
		arg, ok := args[i].(*objects.String)
		if !ok {
			continue
		}
		opt := arg.Value
		if strings.HasPrefix(opt, "-caption=") {
			cfg.caption = strings.TrimPrefix(opt, "-caption=")
		} else if strings.HasPrefix(opt, "-width=") {
			fmt.Sscanf(opt, "-width=%d", &cfg.width)
		} else if strings.HasPrefix(opt, "-height=") {
			fmt.Sscanf(opt, "-height=%d", &cfg.height)
		} else if strings.HasPrefix(opt, "-min=") {
			fmt.Sscanf(opt, "-min=%f", &cfg.minVal)
		} else if strings.HasPrefix(opt, "-max=") {
			fmt.Sscanf(opt, "-max=%f", &cfg.maxVal)
		} else if strings.HasPrefix(opt, "-offset=") {
			fmt.Sscanf(opt, "-offset=%d", &cfg.offset)
		} else if strings.HasPrefix(opt, "-precision=") {
			fmt.Sscanf(opt, "-precision=%d", &cfg.precision)
		} else if strings.HasPrefix(opt, "-captionColor=") {
			fmt.Sscanf(opt, "-captionColor=%d", &cfg.captionColor)
		} else if strings.HasPrefix(opt, "-axisColor=") {
			fmt.Sscanf(opt, "-axisColor=%d", &cfg.axisColor)
		} else if strings.HasPrefix(opt, "-labelColor=") {
			fmt.Sscanf(opt, "-labelColor=%d", &cfg.labelColor)
		} else if strings.HasPrefix(opt, "-seriesColor=") {
			colorStr := strings.TrimPrefix(opt, "-seriesColor=")
			for _, c := range strings.Split(colorStr, ",") {
				c = strings.TrimSpace(c)
				if c != "" {
					var color int
					fmt.Sscanf(c, "%d", &color)
					cfg.seriesColors = append(cfg.seriesColors, color)
				}
			}
		}
	}
	return cfg
}

func ansiColor(code int) string {
	if code <= 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[38;5;%dm", code)
}

func ansiReset() string { return "\x1b[0m" }

const (
	dirUp uint8 = 1 << iota
	dirDown
	dirLeft
	dirRight
)

type plotCell struct {
	connections uint8
	color       int
	// For overlapping series: map of series index to its connections
	seriesConnections map[int]uint8
}

func connectionsToChar(conn uint8) rune {
	switch conn {
	case dirLeft | dirRight:
		return '─'
	case dirUp | dirDown:
		return '│'
	case dirLeft | dirDown:
		return '╮'
	case dirLeft | dirUp:
		return '╯'
	case dirRight | dirDown:
		return '╭'
	case dirRight | dirUp:
		return '╰'
	case dirUp | dirDown | dirLeft | dirRight:
		return '┼'
	case dirUp | dirDown | dirLeft:
		return '┤'
	case dirUp | dirDown | dirRight:
		return '├'
	case dirUp | dirLeft | dirRight:
		return '┴'
	case dirDown | dirLeft | dirRight:
		return '┬'
	case dirDown:
		return '╷'
	case dirUp:
		return '╵'
	case dirLeft:
		return '╶'
	case dirRight:
		return '╴'
	default:
		return ' '
	}
}

func setCell(canvas [][]plotCell, x, y int, dir uint8, color int, seriesIdx int, h, w int) {
	if x < 0 || x >= w || y < 0 || y >= h {
		return
	}
	c := &canvas[y][x]
	// Store per-series connections
	if c.seriesConnections == nil {
		c.seriesConnections = make(map[int]uint8)
	}
	c.seriesConnections[seriesIdx] |= dir
	// Also store combined for backward compatibility
	c.connections |= dir
	if c.color == 0 {
		c.color = color
	}
}

func asciiPlotDataToStr(args ...objects.Object) objects.Object {
	if len(args) < 1 {
		return Error("plotDataToStr requires at least 1 argument")
	}
	dataArr, ok := args[0].(*objects.Array)
	if !ok {
		return Error("first argument must be an array of arrays")
	}

	cfg := parsePlotOptions(args)

	var series [][]float64
	for _, elem := range dataArr.Elements {
		arr, ok := elem.(*objects.Array)
		if !ok {
			continue
		}
		var floats []float64
		for _, v := range arr.Elements {
			switch x := v.(type) {
			case *objects.Int:
				floats = append(floats, float64(x.Value))
			case *objects.Float:
				floats = append(floats, x.Value)
			}
		}
		if len(floats) > 0 {
			series = append(series, floats)
		}
	}
	if len(series) == 0 {
		return Error("no valid data series")
	}

	// Calculate min/max
	if math.IsNaN(cfg.minVal) || math.IsNaN(cfg.maxVal) {
		cfg.minVal = series[0][0]
		cfg.maxVal = series[0][0]
		for _, s := range series {
			for _, v := range s {
				if v < cfg.minVal {
					cfg.minVal = v
				}
				if v > cfg.maxVal {
					cfg.maxVal = v
				}
			}
		}
	}
	if cfg.maxVal == cfg.minVal {
		cfg.maxVal = cfg.minVal + 1
	}

	// Determine plot mode and canvas width
	maxSegs := 0
	for _, s := range series {
		if len(s)-1 > maxSegs {
			maxSegs = len(s) - 1
		}
	}

	compactMode := cfg.width <= 0
	w := maxSegs
	if !compactMode {
		w = cfg.width
	}

	h := cfg.height
	canvas := make([][]plotCell, h)
	for i := range canvas {
		canvas[i] = make([]plotCell, w)
	}

	for seriesIdx, s := range series {
		if len(s) < 2 {
			continue
		}
		color := 0
		if len(cfg.seriesColors) > seriesIdx {
			color = cfg.seriesColors[seriesIdx]
		}

		n := len(s)
		rows := make([]int, n)
		for i, v := range s {
			yn := (v - cfg.minVal) / (cfg.maxVal - cfg.minVal)
			yp := int(yn * float64(h-1))
			yp = max(0, min(h-1, yp))
			rows[i] = h - 1 - yp
		}

		if compactMode {
			drawCompactSeries(canvas, rows, color, seriesIdx, h, w)
		} else {
			cols := make([]int, n)
			for i := 0; i < n; i++ {
				cols[i] = int(float64(i) * float64(w-1) / float64(n-1))
				if cols[i] >= w {
					cols[i] = w - 1
				}
			}
			drawDistributedSeries(canvas, cols, rows, color, seriesIdx, h, w)
		}
	}

	// Build output
	var result strings.Builder
	if cfg.caption != "" {
		if cfg.captionColor > 0 {
			result.WriteString(ansiColor(cfg.captionColor))
		}
		result.WriteString(cfg.caption)
		if cfg.captionColor > 0 {
			result.WriteString(ansiReset())
		}
		result.WriteString("\n")
	}

	for row := 0; row < h; row++ {
		val := cfg.maxVal - float64(row)*(cfg.maxVal-cfg.minVal)/float64(h-1)
		if cfg.labelColor > 0 {
			result.WriteString(ansiColor(cfg.labelColor))
		}
		result.WriteString(fmt.Sprintf("%*.*f ", cfg.offset, cfg.precision, val))
		if cfg.labelColor > 0 {
			result.WriteString(ansiReset())
		}
		if cfg.axisColor > 0 {
			result.WriteString(ansiColor(cfg.axisColor))
		}
		// Determine axis character: use ┼ if there's a line at column 0
		axisChar := "┤"
		if w > 0 {
			conn := canvas[row][0].connections
			if conn != 0 {
				axisChar = "┼"
			}
		}
		result.WriteString(axisChar)
		if cfg.axisColor > 0 {
			result.WriteString(ansiReset())
		}


		for col := 0; col < w; col++ {
			cell := canvas[row][col]

			// Render each series separately to avoid branch characters
			if len(cell.seriesConnections) > 0 {
				// Pick the first series (by index order) to display at this position
				minIdx := -1
				for seriesIdx := range cell.seriesConnections {
					if minIdx == -1 || seriesIdx < minIdx {
						minIdx = seriesIdx
					}
				}
				if minIdx >= 0 {
					conn := cell.seriesConnections[minIdx]
					ch := connectionsToChar(conn)
					if ch != ' ' {
						color := 0
						if len(cfg.seriesColors) > minIdx {
							color = cfg.seriesColors[minIdx]
						}
						if color > 0 {
							result.WriteString(ansiColor(color))
						}
						result.WriteRune(ch)
						if color > 0 {
							result.WriteString(ansiReset())
						}
					} else {
						result.WriteRune(' ')
					}
				} else {
					result.WriteRune(' ')
				}
			} else {
				// Fallback for no seriesConnections
				ch := connectionsToChar(cell.connections)
				if ch != ' ' && cell.color > 0 {
					result.WriteString(ansiColor(cell.color))
					result.WriteRune(ch)
					result.WriteString(ansiReset())
				} else if ch != ' ' {
					result.WriteRune(ch)
				} else {
					result.WriteRune(' ')
				}
			}
		}
		result.WriteString("\n")
	}

	if cfg.axisColor > 0 {
		result.WriteString(ansiColor(cfg.axisColor))
	}
	result.WriteString(strings.Repeat(" ", cfg.offset+1) + "└" + strings.Repeat("─", w))
	if cfg.axisColor > 0 {
		result.WriteString(ansiReset())
	}
	result.WriteString("\n")

	result.WriteString(strings.Repeat(" ", cfg.offset+1) + "0")
	if w > 2 {
		// Determine label interval based on width
		li := w / 5
		if li < 2 {
			li = 2 // minimum interval of 2 to avoid crowding
		}
		if w <= 10 {
			// For small widths, show fewer labels
			li = w / 2
			if li < 2 {
				li = 2
			}
		}
		for i := li; i <= w; i += li {
			result.WriteString(strings.Repeat(" ", li-1))
			if i <= w {
				result.WriteString(fmt.Sprintf("%d", i-1))
			}
		}
	}
	result.WriteString("\n")

	return String(result.String())
}

// drawCompactSeries draws series in compact mode (one column per segment)
func drawCompactSeries(canvas [][]plotCell, rows []int, color, seriesIdx, h, w int) {
	n := len(rows)

	// Draw vertical lines between endpoints
	for i := 0; i < n-1; i++ {
		row1 := rows[i]
		row2 := rows[i+1]
		col := i

		if row1 == row2 {
			setCell(canvas, col, row1, dirLeft|dirRight, color, seriesIdx, h, w)
			continue
		}

		minR := min(row1, row2)
		maxR := max(row1, row2)
		for r := minR + 1; r < maxR; r++ {
			setCell(canvas, col, r, dirUp|dirDown, color, seriesIdx, h, w)
		}
	}

	// Draw corners at each point
	for i := 0; i < n; i++ {
		r := rows[i]

		// Incoming corner (col i-1)
		if i > 0 {
			col := i - 1
			prevRow := rows[i-1]
			if prevRow > r {
				setCell(canvas, col, r, dirDown|dirRight, color, seriesIdx, h, w) // ╭
			} else if prevRow < r {
				setCell(canvas, col, r, dirUp|dirRight, color, seriesIdx, h, w) // ╰
			}
		}

		// Outgoing corner (col i)
		if i < n-1 {
			col := i
			nextRow := rows[i+1]
			if r < nextRow {
				setCell(canvas, col, r, dirLeft|dirDown, color, seriesIdx, h, w) // ╮
			} else if r > nextRow {
				setCell(canvas, col, r, dirLeft|dirUp, color, seriesIdx, h, w) // ╯
			}
		}
	}
}

// drawDistributedSeries draws series in distributed mode
func drawDistributedSeries(canvas [][]plotCell, cols, rows []int, color, seriesIdx, h, w int) {
	n := len(cols)

	// Draw all segments
	for i := 0; i < n-1; i++ {
		x0, y0 := cols[i], rows[i]
		x1, y1 := cols[i+1], rows[i+1]

		if x0 == x1 {
			// Vertical segment only
			minY, maxY := min(y0, y1), max(y0, y1)
			for y := minY; y <= maxY; y++ {
				if y == minY {
					if y0 < y1 {
						setCell(canvas, x0, y, dirDown, color, seriesIdx, h, w)
					} else {
						setCell(canvas, x0, y, dirUp, color, seriesIdx, h, w)
					}
				} else if y == maxY {
					if y0 < y1 {
						setCell(canvas, x0, y, dirUp, color, seriesIdx, h, w)
					} else {
						setCell(canvas, x0, y, dirDown, color, seriesIdx, h, w)
					}
				} else {
					setCell(canvas, x0, y, dirUp|dirDown, color, seriesIdx, h, w)
				}
			}
		} else if y0 == y1 {
			// Horizontal segment only
			minX, maxX := min(x0, x1), max(x0, x1)
			for x := minX; x <= maxX; x++ {
				if x == minX {
					if x0 < x1 {
						setCell(canvas, x, y0, dirRight, color, seriesIdx, h, w)
					} else {
						setCell(canvas, x, y0, dirLeft, color, seriesIdx, h, w)
					}
				} else if x == maxX {
					if x0 < x1 {
						setCell(canvas, x, y0, dirLeft, color, seriesIdx, h, w)
					} else {
						setCell(canvas, x, y0, dirRight, color, seriesIdx, h, w)
					}
				} else {
					setCell(canvas, x, y0, dirLeft|dirRight, color, seriesIdx, h, w)
				}
			}
		} else {
			// Diagonal: use staircase algorithm for true diagonal lines
			dx := x1 - x0
			dy := y1 - y0
			xStep := 1
			if dx < 0 {
				xStep = -1
				dx = -dx
			}
			yStep := 1
			if dy < 0 {
				yStep = -1
				dy = -dy
			}

			// Draw staircase path from (x0, y0) to (x1, y1)
			// Record path as sequence of (x, y) points
			type pt struct{ x, y int }
			var path []pt
			path = append(path, pt{x0, y0})

			cx, cy := x0, y0
			for cx != x1 || cy != y1 {
				remDx := x1 - cx
				if remDx < 0 {
					remDx = -remDx
				}
				remDy := y1 - cy
				if remDy < 0 {
					remDy = -remDy
				}

				if remDx > remDy {
					cx += xStep
				} else if remDy > 0 {
					cy += yStep
				} else if remDx > 0 {
					cx += xStep
				}
				path = append(path, pt{cx, cy})
			}

			// Determine direction at each point based on neighbors
			for i := 0; i < len(path); i++ {
				p := path[i]
				dir := uint8(0)

				// If this is the start point at col 0, add dirLeft to connect to axis
				if i == 0 && p.x == 0 {
					dir |= dirLeft
				}

				// Check previous point
				if i > 0 {
					prev := path[i-1]
					if prev.x < p.x {
						dir |= dirLeft
					} else if prev.x > p.x {
						dir |= dirRight
					} else if prev.y < p.y {
						dir |= dirUp
					} else if prev.y > p.y {
						dir |= dirDown
					}
				}

				// Check next point
				if i < len(path)-1 {
					next := path[i+1]
					if next.x > p.x {
						dir |= dirRight
					} else if next.x < p.x {
						dir |= dirLeft
					} else if next.y > p.y {
						dir |= dirDown
					} else if next.y < p.y {
						dir |= dirUp
					}
				}

				if dir != 0 {
					setCell(canvas, p.x, p.y, dir, color, seriesIdx, h, w)
				}
			}
		}
	}

}
