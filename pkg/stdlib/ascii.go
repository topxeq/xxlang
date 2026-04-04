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
		width: 60, height: 7, minVal: math.NaN(), maxVal: math.NaN(),
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
		return '╵'
	case dirUp:
		return '╷'
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

	// Calculate width based on max number of segments across all series
	maxSegs := 0
	for _, s := range series {
		if len(s)-1 > maxSegs {
			maxSegs = len(s) - 1
		}
	}
	if maxSegs < cfg.width {
		maxSegs = cfg.width
	}

	h, w := cfg.height, maxSegs
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
			rows[i] = h - 1 - yp // invert: max value at row 0 (top), min at row h-1 (bottom)
		}

		// Charlang-style compact algorithm:
		// Each segment i connects point i to point i+1, drawn in column i
		// At each point, the incoming segment (col i-1) and outgoing segment (col i) both draw corners
		// Between points, segments draw vertical lines

		// First, draw all vertical lines for segments between their endpoints
		for i := 0; i < n-1; i++ {
			row1 := rows[i]
			row2 := rows[i+1]
			col := i

			if row1 == row2 {
				// Horizontal segment
				setCell(canvas, col, row1, dirLeft|dirRight, color, h, w)
				continue
			}

			minR := min(row1, row2)
			maxR := max(row1, row2)

			// Draw vertical lines for rows strictly between the endpoints
			for r := minR + 1; r < maxR; r++ {
				setCell(canvas, col, r, dirUp|dirDown, color, h, w)
			}
		}

		// Then, draw corners at each point
		for i := 0; i < n; i++ {
			r := rows[i]

			// Draw corner for incoming segment (col i-1) if not the first point
			// The corner direction depends on where the segment CAME FROM (rows[i-1])
			if i > 0 {
				col := i - 1
				prevRow := rows[i-1]
				if prevRow > r {
					// Segment went UP (from lower row to higher row numerically, but visually from bottom to top)
					// At the endpoint, connection comes from below, draw ╭ (dirDown|dirRight)
					setCell(canvas, col, r, dirDown|dirRight, color, h, w)
				} else if prevRow < r {
					// Segment went DOWN (from higher row to lower row numerically, but visually from top to bottom)
					// At the endpoint, connection comes from above, draw ╰ (dirUp|dirRight)
					setCell(canvas, col, r, dirUp|dirRight, color, h, w)
				}
			}

			// Draw corner for outgoing segment (col i) if not the last point
			if i < n-1 {
				col := i
				nextRow := rows[i+1]
				if r < nextRow {
					// Going DOWN (to lower value visually): draw ╮
					setCell(canvas, col, r, dirLeft|dirDown, color, h, w)
				} else if r > nextRow {
					// Going UP (to higher value visually): draw ╯
					setCell(canvas, col, r, dirLeft|dirUp, color, h, w)
				}
			}
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