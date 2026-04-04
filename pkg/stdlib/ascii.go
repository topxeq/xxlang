// pkg/stdlib/ascii.go
// ASCII plotting module for the Xxlang standard library.
// This module provides ASCII chart plotting functionality without external dependencies.
// Implements continuous curve plotting based on asciigraph algorithm.
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

// charSet holds the characters used for drawing
type charSet struct {
	Horizontal    rune
	Vertical      rune
	ArcUpLeft     rune // ┘
	ArcUpRight    rune // └
	ArcDownLeft   rune // ┐
	ArcDownRight  rune // ┌
}

// defaultCharSet returns the default character set for drawing
func defaultCharSet() charSet {
	return charSet{
		Horizontal:    '─',
		Vertical:      '│',
		ArcUpLeft:     '┘',
		ArcUpRight:    '└',
		ArcDownLeft:   '┐',
		ArcDownRight:  '┌',
	}
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
	charSet := defaultCharSet()

	// Create the plot canvas with cells
	canvas := make([][]plotCell, height)
	for i := range canvas {
		canvas[i] = make([]plotCell, width)
		for j := range canvas[i] {
			canvas[i][j] = plotCell{char: ' ', color: 0}
		}
	}

	// Plot each series
	for seriesIdx, s := range series {
		if len(s) < 2 {
			continue
		}

		// Get color for this series
		seriesColor := 0
		if len(cfg.seriesColors) > seriesIdx {
			seriesColor = cfg.seriesColors[seriesIdx]
		}

		// Calculate Y positions for all data points
		n := len(s)
		yPositions := make([]int, n)
		for i, v := range s {
			// Map value to Y position (0 = bottom, height-1 = top)
			yFloat := (v - cfg.minVal) / (cfg.maxVal - cfg.minVal) * float64(height-1)
			yPos := int(math.Round(yFloat))
			if yPos < 0 {
				yPos = 0
			}
			if yPos >= height {
				yPos = height - 1
			}
			// Convert to canvas row (0 = top)
			yPositions[i] = height - 1 - yPos
		}

		// Draw the curve using linear interpolation for smooth curves
		// Map data indices to column positions across the full width
		for i := 0; i < n-1; i++ {
			y0 := yPositions[i]
			y1 := yPositions[i+1]

			// Calculate x positions spread across the width
			x0 := int(float64(i) * float64(width-1) / float64(n-1))
			x1 := int(float64(i+1) * float64(width-1) / float64(n-1))

			// Draw horizontal line if same Y level
			if y0 == y1 {
				for x := x0; x <= x1; x++ {
					setCell(canvas, x, y0, charSet.Horizontal, seriesColor)
				}
			} else {
				// Draw curve with arcs
				// Determine arc characters based on direction
				var topChar, bottomChar rune
				var topY, bottomY int

				if y0 < y1 {
					// Going down on canvas
					topChar = charSet.ArcDownLeft
					bottomChar = charSet.ArcUpRight
					topY = y0
					bottomY = y1
				} else {
					// Going up on canvas
					topChar = charSet.ArcUpLeft
					bottomChar = charSet.ArcDownRight
					topY = y1
					bottomY = y0
				}

				// Draw horizontal segment before vertical transition
				if x0 < x1-1 {
					for x := x0; x < x1-1; x++ {
						setCell(canvas, x, y0, charSet.Horizontal, seriesColor)
					}
				}

				// Draw the arc transition at the last column before x1
				transitionX := x1 - 1
				if transitionX < x0 {
					transitionX = x0
				}

				setCell(canvas, transitionX, topY, topChar, seriesColor)
				setCell(canvas, transitionX, bottomY, bottomChar, seriesColor)

				// Fill vertical space between
				for y := topY + 1; y < bottomY; y++ {
					setCell(canvas, transitionX, y, charSet.Vertical, seriesColor)
				}
			}
		}

		// Draw the last point
		if n > 0 {
			lastY := yPositions[n-1]
			lastX := int(float64(n-1) * float64(width-1) / float64(n-1))
			if lastX >= width {
				lastX = width - 1
			}
			if canvas[lastY][lastX].char == ' ' {
				setCell(canvas, lastX, lastY, charSet.Horizontal, seriesColor)
			}
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
		labelInterval := width / 5
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

// setCell sets a cell in the canvas
func setCell(canvas [][]plotCell, x, y int, char rune, color int) {
	if y < 0 || y >= len(canvas) || x < 0 || x >= len(canvas[0]) {
		return
	}
	canvas[y][x] = plotCell{char: char, color: color}
}