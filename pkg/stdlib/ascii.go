// pkg/stdlib/ascii.go
// ASCII plotting module for the Xxlang standard library.
// This module provides ASCII chart plotting functionality without external dependencies.
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
			// plotDataToStr renders numeric series as ASCII plot string.
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
		captionColor: 0, // 0 means no color
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

// asciiPlotDataToStr implements the plotDataToStr function
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

	// Create plot buffer
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

	// Create plot grid
	height := cfg.height
	width := cfg.width
	offset := cfg.offset

	// Create the plot canvas
	canvas := make([][]rune, height)
	for i := range canvas {
		canvas[i] = make([]rune, width)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	// Plot each series
	plotChars := []rune{'*', '+', 'x', 'o', '#', '@', '&', '%'}
	colorCodes := cfg.seriesColors

	for seriesIdx, s := range series {
		if len(s) == 0 {
			continue
		}
		char := plotChars[seriesIdx%len(plotChars)]

		for i, v := range s {
			// Calculate x position
			x := int(float64(i) * float64(width-1) / float64(len(s)-1))
			if len(s) == 1 {
				x = width / 2
			}

			// Calculate y position
			y := int((cfg.maxVal - v) / (cfg.maxVal - cfg.minVal) * float64(height-1))
			if y < 0 {
				y = 0
			}
			if y >= height {
				y = height - 1
			}

			canvas[y][x] = char
		}
	}

	// Render Y axis and labels
	if cfg.axisColor > 0 {
		result.WriteString(ansiColor(cfg.axisColor))
	}

	for row := 0; row < height; row++ {
		// Y-axis label
		val := cfg.maxVal - float64(row)*(cfg.maxVal-cfg.minVal)/float64(height-1)
		if cfg.labelColor > 0 {
			result.WriteString(ansiColor(cfg.labelColor))
		}
		result.WriteString(fmt.Sprintf("%*.*f ", offset, cfg.precision, val))
		if cfg.labelColor > 0 {
			result.WriteString(ansiReset())
		}

		// Y-axis line
		if cfg.axisColor > 0 {
			result.WriteString(ansiColor(cfg.axisColor))
		}
		result.WriteString("|")
		if cfg.axisColor > 0 {
			result.WriteString(ansiReset())
		}

		// Plot content with colors
		for col := 0; col < width; col++ {
			ch := canvas[row][col]
			if ch != ' ' {
				// Find which series this point belongs to and apply color
				colorApplied := false
				for seriesIdx, s := range series {
					for i, v := range s {
						x := int(float64(i) * float64(width-1) / float64(len(s)-1))
						if len(s) == 1 {
							x = width / 2
						}
						y := int((cfg.maxVal - v) / (cfg.maxVal - cfg.minVal) * float64(height-1))
						if y < 0 {
							y = 0
						}
						if y >= height {
							y = height - 1
						}
						if y == row && x == col {
							if len(colorCodes) > seriesIdx && colorCodes[seriesIdx] > 0 {
								result.WriteString(ansiColor(colorCodes[seriesIdx]))
								result.WriteRune(ch)
								result.WriteString(ansiReset())
								colorApplied = true
							}
							break
						}
					}
					if colorApplied {
						break
					}
				}
				if !colorApplied {
					result.WriteRune(ch)
				}
			} else {
				result.WriteRune(ch)
			}
		}
		result.WriteString("\n")
	}

	// X-axis
	if cfg.axisColor > 0 {
		result.WriteString(ansiColor(cfg.axisColor))
	}
	result.WriteString(strings.Repeat(" ", offset+1))
	result.WriteString("+")
	result.WriteString(strings.Repeat("-", width-1))
	result.WriteString("\n")

	// X-axis labels (0 to n-1)
	result.WriteString(strings.Repeat(" ", offset+1))
	result.WriteString("0")
	if width > 2 {
		// Add labels at intervals
		labelInterval := (width - 1) / 5
		if labelInterval < 1 {
			labelInterval = 1
		}
		for i := labelInterval; i < width; i += labelInterval {
			result.WriteString(strings.Repeat(" ", labelInterval-1))
			if i < width {
				result.WriteString(fmt.Sprintf("%d", i))
			}
		}
	}
	result.WriteString("\n")

	if cfg.axisColor > 0 {
		result.WriteString(ansiReset())
	}

	return String(result.String())
}