// pkg/stdlib/ascii.go
// ASCII plotting module for the Xxlang standard library.
// This module provides ASCII chart plotting functionality without external dependencies.
// Implements continuous curve plotting with Unicode box-drawing characters and ANSI colors.
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
			// plotDataToStr renders numeric series as ASCII plot string with colored curves.
			// Usage: plotDataToStr(data, options...)
			// data: array of arrays containing numeric values (multiple series)
			// options:
			//   -caption=string      : chart title
			//   -width=int          : plot width (default: 60)
			//   -height=int         : plot height (default: 10)
			//   -min=float          : minimum value on Y axis
			//   -max=float          : maximum value on Y axis
			//   -offset=int         : left offset for labels (default: 5)
			//   -precision=int      : decimal precision for labels (default: 2)
			//   -captionColor=int   : ANSI color code for caption
			//   -axisColor=int      : ANSI color code for axis
			//   -labelColor=int     : ANSI color code for labels
			//   -seriesColor=string : comma-separated ANSI color codes for series
			"plotDataToStr": BuiltinFunc(asciiPlotDataToStr),
		},
	})
}

// plotConfig holds the configuration for ASCII plotting
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

// defaultConfig returns default plot configuration
func defaultConfig() plotConfig {
	return plotConfig{
		width:        60,
		height:       10,
		minVal:       math.NaN(),
		maxVal:       math.NaN(),
		offset:       5,
		precision:    2,
		caption:      "",
		captionColor: 0,
		axisColor:    0,
		labelColor:   0,
		seriesColors: nil,
	}
}

// parsePlotOptions parses options from arguments
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
			colors := strings.Split(colorStr, ",")
			for _, c := range colors {
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

// ansiColor returns ANSI color escape sequence
func ansiColor(code int) string {
	if code <= 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[38;5;%dm", code)
}

// ansiReset returns ANSI reset sequence
func ansiReset() string {
	return "\x1b[0m"
}

// plotCell represents a cell in the plot canvas with character and color
type plotCell struct {
	char  rune
	color int // ANSI color code, 0 means default
}

// asciiPlotDataToStr implements the plotDataToStr function with continuous curves
func asciiPlotDataToStr(args ...objects.Object) objects.Object {
	if len(args) < 1 {
		return Error("plotDataToStr requires at least 1 argument")
	}

	// Parse data series
	dataArr, ok := args[0].(*objects.Array)
	if !ok {
		return Error("first argument must be an array of arrays")
	}

	// Parse options
	cfg := parsePlotOptions(args)

	// Convert data to float slices
	var series [][]float64
	for _, elem := range dataArr.Elements {
		seriesArr, ok := elem.(*objects.Array)
		if !ok {
			continue
		}
		var floats []float64
		for _, val := range seriesArr.Elements {
			switch v := val.(type) {
			case *objects.Int:
				floats = append(floats, float64(v.Value))
			case *objects.Float:
				floats = append(floats, v.Value)
			}
		}
		if len(floats) > 0 {
			series = append(series, floats)
		}
	}

	if len(series) == 0 {
		return Error("no valid data series")
	}

	// Calculate min/max if not specified
	if math.IsNaN(cfg.minVal) || math.IsNaN(cfg.maxVal) {
		for _, s := range series {
			for _, v := range s {
				if math.IsNaN(cfg.minVal) || v < cfg.minVal {
					cfg.minVal = v
				}
				if math.IsNaN(cfg.maxVal) || v > cfg.maxVal {
					cfg.maxVal = v
				}
			}
		}
	}

	// Prevent division by zero
	if cfg.maxVal == cfg.minVal {
		cfg.maxVal = cfg.minVal + 1
	}

	height := cfg.height
	width := cfg.width

	// Create the plot canvas with cells
	canvas := make([][]plotCell, height)
	for i := range canvas {
		canvas[i] = make([]plotCell, width)
		for j := range canvas[i] {
			canvas[i][j] = plotCell{char: ' ', color: 0}
		}
	}

	// Plot each series as continuous curves
	for seriesIdx, s := range series {
		if len(s) < 2 {
			continue
		}

		// Get color for this series
		seriesColor := 0
		if len(cfg.seriesColors) > seriesIdx {
			seriesColor = cfg.seriesColors[seriesIdx]
		}

		// Calculate positions for all points
		points := make([]struct{ x, y int }, len(s))
		for i, v := range s {
			// X position: spread data points across the width
			points[i].x = int(float64(i) * float64(width-1) / float64(len(s)-1))

			// Y position: map value to row (inverted because row 0 is top)
			yFloat := (cfg.maxVal - v) / (cfg.maxVal - cfg.minVal) * float64(height-1)
			points[i].y = int(math.Round(yFloat))
			if points[i].y < 0 {
				points[i].y = 0
			}
			if points[i].y >= height {
				points[i].y = height - 1
			}
		}

		// Draw curves between consecutive points
		for i := 0; i < len(points)-1; i++ {
			drawCurve(canvas, points[i].x, points[i].y, points[i+1].x, points[i+1].y, seriesColor, height, width)
		}
	}

	// Build the output string
	var result strings.Builder

	// Add caption
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

	// Render Y axis and plot content
	for row := 0; row < height; row++ {
		// Y-axis label
		val := cfg.maxVal - float64(row)*(cfg.maxVal-cfg.minVal)/float64(height-1)
		if cfg.labelColor > 0 {
			result.WriteString(ansiColor(cfg.labelColor))
		}
		result.WriteString(fmt.Sprintf("%*.*f ", cfg.offset, cfg.precision, val))
		if cfg.labelColor > 0 {
			result.WriteString(ansiReset())
		}

		// Y-axis line
		if cfg.axisColor > 0 {
			result.WriteString(ansiColor(cfg.axisColor))
		}
		result.WriteString("│")
		if cfg.axisColor > 0 {
			result.WriteString(ansiReset())
		}

		// Plot content with colors
		for col := 0; col < width; col++ {
			cell := canvas[row][col]
			if cell.char != ' ' && cell.color > 0 {
				result.WriteString(ansiColor(cell.color))
				result.WriteRune(cell.char)
				result.WriteString(ansiReset())
			} else {
				result.WriteRune(cell.char)
			}
		}
		result.WriteString("\n")
	}

	// X-axis
	if cfg.axisColor > 0 {
		result.WriteString(ansiColor(cfg.axisColor))
	}
	result.WriteString(strings.Repeat(" ", cfg.offset+1))
	result.WriteString("└")
	result.WriteString(strings.Repeat("─", width))
	if cfg.axisColor > 0 {
		result.WriteString(ansiReset())
	}
	result.WriteString("\n")

	// X-axis labels
	result.WriteString(strings.Repeat(" ", cfg.offset+1))
	result.WriteString("0")
	if width > 2 {
		// Add labels at intervals
		labelInterval := (width) / 5
		if labelInterval < 1 {
			labelInterval = 1
		}
		for i := labelInterval; i <= width; i += labelInterval {
			result.WriteString(strings.Repeat(" ", labelInterval-1))
			if i <= width {
				result.WriteString(fmt.Sprintf("%d", i-1))
			}
		}
	}
	result.WriteString("\n")

	return String(result.String())
}

// drawCurve draws a curve between two points using Unicode box-drawing characters
func drawCurve(canvas [][]plotCell, x1, y1, x2, y2, color int, height, width int) {
	// Simple case: same position
	if x1 == x2 && y1 == y2 {
		setCell(canvas, x1, y1, '─', color, height, width)
		return
	}

	// Horizontal line
	if y1 == y2 {
		for x := x1; x <= x2; x++ {
			setCell(canvas, x, y1, '─', color, height, width)
		}
		return
	}

	// Vertical line
	if x1 == x2 {
		for y := y1; y <= y2; y++ {
			setCell(canvas, x1, y, '│', color, height, width)
		}
		return
	}

	// Diagonal or complex path: usebresenham-like line algorithm
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)

	// Determine direction
	xStep := 1
	if x2 < x1 {
		xStep = -1
	}
	yStep := 1
	if y2 < y1 {
		yStep = -1
	}

	x, y := x1, y1

	if dx > dy {
		// More horizontal movement
		err := dy / 2
		for x != x2 {
			// Choose character based on movement direction
			char := getLineChar(x, y, x+xStep, y, canvas, height, width)
			setCell(canvas, x, y, char, color, height, width)

			x += xStep
			err += dy
			if err >= dx {
				y += yStep
				err -= dx
				// Draw corner character
				char = getCornerChar(x-xStep, y-yStep, x, y, x+xStep, y)
				setCell(canvas, x-xStep, y-yStep, char, color, height, width)
			}
		}
		setCell(canvas, x2, y2, '─', color, height, width)
	} else {
		// More vertical movement
		err := dx / 2
		prevY := y
		for y != y2 {
			// Choose character based on movement direction
			if y != prevY && x != x2 {
				// Moving diagonally - draw corner
				char := getCornerChar(x, prevY, x, y, x+xStep, y)
				setCell(canvas, x, prevY, char, color, height, width)
			} else {
				char := getLineChar(x, y, x, y+yStep, canvas, height, width)
				setCell(canvas, x, y, char, color, height, width)
			}

			prevY = y
			y += yStep
			err += dx
			if err >= dy {
				x += xStep
				err -= dy
			}
		}
		setCell(canvas, x2, y2, '│', color, height, width)
	}
}

// getLineChar returns appropriate line character for direction
func getLineChar(x1, y1, x2, y2 int, canvas [][]plotCell, height, width int) rune {
	if x1 == x2 {
		// Vertical movement
		return '│'
	}
	if y1 == y2 {
		// Horizontal movement
		return '─'
	}
	// Diagonal - use appropriate corner based on direction
	if x2 > x1 && y2 > y1 {
		return '╯' // going down-right
	}
	if x2 > x1 && y2 < y1 {
		return '╮' // going up-right
	}
	if x2 < x1 && y2 > y1 {
		return '╰' // going down-left
	}
	if x2 < x1 && y2 < y1 {
		return '╭' // going up-left
	}
	return '─'
}

// getCornerChar returns appropriate corner character
func getCornerChar(prevX, prevY, curX, curY, nextX, nextY int) rune {
	// Determine incoming and outgoing directions
	fromLeft := prevX < curX
	fromRight := prevX > curX
	fromUp := prevY < curY
	fromDown := prevY > curY

	toLeft := nextX < curX
	toRight := nextX > curX
	toUp := nextY < curY
	toDown := nextY > curY

	// Combine to choose corner
	if fromLeft && toDown {
		return '╮'
	}
	if fromLeft && toUp {
		return '╯'
	}
	if fromRight && toDown {
		return '╭'
	}
	if fromRight && toUp {
		return '╰'
	}
	if fromUp && toRight {
		return '╰'
	}
	if fromUp && toLeft {
		return '╯'
	}
	if fromDown && toRight {
		return '╭'
	}
	if fromDown && toLeft {
		return '╮'
	}

	// Simple connections
	if fromLeft || fromRight || toLeft || toRight {
		return '─'
	}
	return '│'
}

// setCell sets a cell in the canvas with bounds checking
func setCell(canvas [][]plotCell, x int, y int, char rune, color int, height int, width int) {
	if x < 0 || x >= width || y < 0 || y >= height {
		return
	}
	// Merge with existing cell - prefer more complex characters
	existing := canvas[y][x]
	if existing.char == ' ' || existing.char == '─' || existing.char == '│' {
		// Override simple characters with corners
		if isCornerChar(char) && !isCornerChar(existing.char) {
			canvas[y][x] = plotCell{char: char, color: color}
		} else if existing.char == ' ' {
			canvas[y][x] = plotCell{char: char, color: color}
		}
	}
}

// isCornerChar checks if a character is a corner
func isCornerChar(c rune) bool {
	return c == '╭' || c == '╮' || c == '╯' || c == '╰' ||
		c == '┌' || c == '┐' || c == '└' || c == '┘'
}

// abs returns absolute value of an integer
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}