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
			"plotDataToStr": BuiltinFunc(asciiPlotDataToStr),
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

func setCell(canvas [][]plotCell, x, y int, dir uint8, color int, h, w int) {
	if x < 0 || x >= w || y < 0 || y >= h {
		return
	}
	c := &canvas[y][x]
	if c.connections == 0 {
		c.color = color
	}
	c.connections |= dir
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
			drawCompactSeries(canvas, rows, color, h, w)
		} else {
			cols := make([]int, n)
			for i := 0; i < n; i++ {
				cols[i] = int(float64(i) * float64(w-1) / float64(n-1))
				if cols[i] >= w {
					cols[i] = w - 1
				}
			}
			drawDistributedSeries(canvas, cols, rows, color, h, w)
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
		result.WriteString("│")
		if cfg.axisColor > 0 {
			result.WriteString(ansiReset())
		}
		for col := 0; col < w; col++ {
			cell := canvas[row][col]
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
		li := w / 5
		if li < 1 {
			li = 1
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
func drawCompactSeries(canvas [][]plotCell, rows []int, color, h, w int) {
	n := len(rows)

	// Draw vertical lines between endpoints
	for i := 0; i < n-1; i++ {
		row1 := rows[i]
		row2 := rows[i+1]
		col := i

		if row1 == row2 {
			setCell(canvas, col, row1, dirLeft|dirRight, color, h, w)
			continue
		}

		minR := min(row1, row2)
		maxR := max(row1, row2)
		for r := minR + 1; r < maxR; r++ {
			setCell(canvas, col, r, dirUp|dirDown, color, h, w)
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
				setCell(canvas, col, r, dirDown|dirRight, color, h, w) // ╭
			} else if prevRow < r {
				setCell(canvas, col, r, dirUp|dirRight, color, h, w) // ╰
			}
		}

		// Outgoing corner (col i)
		if i < n-1 {
			col := i
			nextRow := rows[i+1]
			if r < nextRow {
				setCell(canvas, col, r, dirLeft|dirDown, color, h, w) // ╮
			} else if r > nextRow {
				setCell(canvas, col, r, dirLeft|dirUp, color, h, w) // ╯
			}
		}
	}
}

// drawDistributedSeries draws series in distributed mode (Charlang style)
// For diagonal segments: draw vertical line first, then curve, then horizontal line
// Curve position: at (x0, y1) - where the segment ends vertically
func drawDistributedSeries(canvas [][]plotCell, cols, rows []int, color, h, w int) {
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
						setCell(canvas, x0, y, dirDown, color, h, w)
					} else {
						setCell(canvas, x0, y, dirUp, color, h, w)
					}
				} else if y == maxY {
					if y0 < y1 {
						setCell(canvas, x0, y, dirUp, color, h, w)
					} else {
						setCell(canvas, x0, y, dirDown, color, h, w)
					}
				} else {
					setCell(canvas, x0, y, dirUp|dirDown, color, h, w)
				}
			}
		} else if y0 == y1 {
			// Horizontal segment only
			minX, maxX := min(x0, x1), max(x0, x1)
			for x := minX; x <= maxX; x++ {
				if x == minX {
					if x0 < x1 {
						setCell(canvas, x, y0, dirRight, color, h, w)
					} else {
						setCell(canvas, x, y0, dirLeft, color, h, w)
					}
				} else if x == maxX {
					if x0 < x1 {
						setCell(canvas, x, y0, dirLeft, color, h, w)
					} else {
						setCell(canvas, x, y0, dirRight, color, h, w)
					}
				} else {
					setCell(canvas, x, y0, dirLeft|dirRight, color, h, w)
				}
			}
		} else {
			// Diagonal: Charlang style
			// At start point (x0, y0): draw curve that shows the initial direction
			// Vertical line from y0 to y1 (exclusive of endpoints that will be curves)
			// At (x0, y1): draw curve connecting vertical to horizontal
			// Horizontal line from x0 to x1 (exclusive of endpoints)

			// Start point - curve showing initial direction
			// When x0 < x1 (going right), the curve at start shows right+vertical direction
			if y0 > y1 {
				// Going up (visually)
				if x0 < x1 {
					setCell(canvas, x0, y0, dirLeft|dirUp, color, h, w) // ╯ (right+up curve)
				} else {
					setCell(canvas, x0, y0, dirRight|dirUp, color, h, w) // ╰ (left+up curve)
				}
			} else {
				// Going down (visually)
				if x0 < x1 {
					setCell(canvas, x0, y0, dirLeft|dirDown, color, h, w) // ╮ (right+down curve)
				} else {
					setCell(canvas, x0, y0, dirRight|dirDown, color, h, w) // ╭ (left+down curve)
				}
			}

			// Vertical line (interior only, curve handles the endpoint)
			minY, maxY := min(y0, y1), max(y0, y1)
			for y := minY + 1; y < maxY; y++ {
				setCell(canvas, x0, y, dirUp|dirDown, color, h, w)
			}

			// Curve at (x0, y1) - connects vertical to horizontal
			if y0 > y1 {
				// Going up
				if x0 < x1 {
					setCell(canvas, x0, y1, dirRight|dirDown, color, h, w) // ╭
				} else {
					setCell(canvas, x0, y1, dirLeft|dirUp, color, h, w) // ╯
				}
			} else {
				// Going down
				if x0 < x1 {
					setCell(canvas, x0, y1, dirRight|dirUp, color, h, w) // ╰
				} else {
					setCell(canvas, x0, y1, dirLeft|dirDown, color, h, w) // ╮
				}
			}

			// Horizontal line (interior)
			minX, maxX := min(x0, x1), max(x0, x1)
			for x := minX + 1; x < maxX; x++ {
				setCell(canvas, x, y1, dirLeft|dirRight, color, h, w)
			}

			// End point will be handled as start of next segment or as final endpoint
		}
	}

	// Handle junctions at middle points (where two segments meet)
	for i := 1; i < n-1; i++ {
		x, y := cols[i], rows[i]
		xPrev, yPrev := cols[i-1], rows[i-1]
		xNext, yNext := cols[i+1], rows[i+1]

		// Clear and redraw the junction point
		// Incoming segment direction
		inDir := uint8(0)
		if xPrev == x {
			// Incoming vertical
			if yPrev > y {
				inDir = dirDown // came from above
			} else {
				inDir = dirUp // came from below
			}
		} else {
			// Incoming horizontal
			if xPrev < x {
				inDir = dirLeft // came from left
			} else {
				inDir = dirRight // came from right
			}
		}

		// Outgoing segment direction
		outDir := uint8(0)
		if x == xNext {
			// Outgoing vertical
			if y > yNext {
				outDir = dirUp // going up
			} else {
				outDir = dirDown // going down
			}
		} else {
			// Outgoing horizontal
			if x < xNext {
				outDir = dirRight // going right
			} else {
				outDir = dirLeft // going left
			}
		}

		setCell(canvas, x, y, inDir|outDir, color, h, w)
	}
}