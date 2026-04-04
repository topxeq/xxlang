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

// plotCell represents a cell in the plot canvas
type plotCell struct {
	char  rune
	color int
}

// asciiPlotDataToStr implements the plotDataToStr function
// Based on asciigraph algorithm: https://github.com/guptarohit/asciigraph
func asciiPlotDataToStr(args ...objects.Object) objects.Object {
	if len(args) < 1 {
		return Error("plotDataToStr requires at least 1 argument")
	}

	dataArr, ok := args[0].(*objects.Array)
	if !ok {
		return Error("first argument must be an array of arrays")
	}

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

	if cfg.maxVal == cfg.minVal {
		cfg.maxVal = cfg.minVal + 1
	}

	height := cfg.height
	width := cfg.width

	// Create canvas
	canvas := make([][]plotCell, height)
	for i := range canvas {
		canvas[i] = make([]plotCell, width)
	}

	// Plot each series using asciigraph algorithm
	for seriesIdx, s := range series {
		if len(s) < 2 {
			continue
		}

		seriesColor := 0
		if len(cfg.seriesColors) > seriesIdx {
			seriesColor = cfg.seriesColors[seriesIdx]
		}

		n := len(s)

		// Map each data point to canvas positions
		// X: distribute across width
		// Y: 0 = top, height-1 = bottom
		type pt struct {
			x, y int
		}
		points := make([]pt, n)
		for i, v := range s {
			// X: distribute across width
			points[i].x = int(float64(i) * float64(width-1) / float64(n-1))
			if points[i].x >= width {
				points[i].x = width - 1
			}

			// Y: map value to row
			yNorm := (v - cfg.minVal) / (cfg.maxVal - cfg.minVal)
			yPos := int(yNorm * float64(height-1))
			if yPos < 0 {
				yPos = 0
			}
			if yPos >= height {
				yPos = height - 1
			}
			// Invert: canvas row 0 is top
			points[i].y = height - 1 - yPos
		}

		// Draw lines between consecutive points
		for i := 0; i < n-1; i++ {
			p0 := points[i]
			p1 := points[i+1]

			// Draw horizontal segment from p0.x to just before transition
			// Then draw vertical transition at the last column before p1.x

			if p0.x == p1.x {
				// Same column - just draw vertical transition
				drawVerticalTransition(canvas, p0.x, p0.y, p1.y, seriesColor, height, width)
			} else if p0.y == p1.y {
				// Same Y - draw horizontal line
				for x := p0.x; x <= p1.x; x++ {
					setCell(canvas, x, p0.y, '─', seriesColor, height, width)
				}
			} else {
				// Both X and Y change
				// Strategy: horizontal line then vertical transition
				// Draw horizontal from p0.x to p1.x-1
				for x := p0.x; x < p1.x; x++ {
					setCell(canvas, x, p0.y, '─', seriesColor, height, width)
				}
				// Draw vertical transition at p1.x-1 (or p1.x if adjacent)
				transitionX := p1.x - 1
				if transitionX < p0.x {
					transitionX = p0.x
				}
				drawVerticalTransition(canvas, transitionX, p0.y, p1.y, seriesColor, height, width)
			}
		}
	}

	// Build output
	var result strings.Builder

	// Caption
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

	// Y-axis and plot content
	for row := 0; row < height; row++ {
		// Label
		val := cfg.maxVal - float64(row)*(cfg.maxVal-cfg.minVal)/float64(height-1)
		if cfg.labelColor > 0 {
			result.WriteString(ansiColor(cfg.labelColor))
		}
		result.WriteString(fmt.Sprintf("%*.*f ", cfg.offset, cfg.precision, val))
		if cfg.labelColor > 0 {
			result.WriteString(ansiReset())
		}

		// Axis
		if cfg.axisColor > 0 {
			result.WriteString(ansiColor(cfg.axisColor))
		}
		result.WriteString("│")
		if cfg.axisColor > 0 {
			result.WriteString(ansiReset())
		}

		// Plot row
		for col := 0; col < width; col++ {
			cell := canvas[row][col]
			if cell.char != 0 && cell.color > 0 {
				result.WriteString(ansiColor(cell.color))
				result.WriteRune(cell.char)
				result.WriteString(ansiReset())
			} else if cell.char != 0 {
				result.WriteRune(cell.char)
			} else {
				result.WriteRune(' ')
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

// setCell sets a cell on the canvas
func setCell(canvas [][]plotCell, x, y int, char rune, color int, height, width int) {
	if x < 0 || x >= width || y < 0 || y >= height {
		return
	}
	canvas[y][x] = plotCell{char: char, color: color}
}

// drawVerticalTransition draws a vertical transition with arc characters
//
// Arc characters and their correct usage:
//
//	╯ (lines on right and bottom): from left horizontal, curve UP
//	╰ (lines on left and bottom): from left horizontal, curve DOWN
//	╭ (lines on left and top): from BELOW vertical, curve right
//	╮ (lines on right and top): from ABOVE vertical, curve right
//
// When going UP on canvas (y decreases, value increases):
//   - Start point: horizontal from left, then curve UP → ╯
//   - End point: came from BELOW (larger y), curve right → ╭
//
// When going DOWN on canvas (y increases, value decreases):
//   - Start point: horizontal from left, then curve DOWN → ╰
//   - End point: came from ABOVE (smaller y), curve right → ╮
func drawVerticalTransition(canvas [][]plotCell, x, y0, y1, color int, height, width int) {
	if y0 == y1 {
		setCell(canvas, x, y0, '─', color, height, width)
		return
	}

	if y0 > y1 {
		// Going up on canvas (y decreasing)
		setCell(canvas, x, y0, '╯', color, height, width) // horizontal→up
		setCell(canvas, x, y1, '╭', color, height, width) // from below→right
		for y := y1 + 1; y < y0; y++ {
			setCell(canvas, x, y, '│', color, height, width)
		}
	} else {
		// Going down on canvas (y increasing)
		setCell(canvas, x, y0, '╰', color, height, width) // horizontal→down
		setCell(canvas, x, y1, '╮', color, height, width) // from above→right
		for y := y0 + 1; y < y1; y++ {
			setCell(canvas, x, y, '│', color, height, width)
		}
	}
}