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

// direction constants for line connections
const (
	dirUp    uint8 = 1
	dirDown  uint8 = 2
	dirLeft  uint8 = 4
	dirRight uint8 = 8
)

// plotCell represents a cell in the plot canvas
// Uses direction flags to track which lines pass through this cell
type plotCell struct {
	connections uint8 // bitmask of directions that have lines
	color       int
}

// connectionsToChar converts connection directions to the appropriate character
// Arc characters (based on Unicode box drawing):
//   ╭ : lines at right and down - curves from right-to-down or down-to-right
//   ╮ : lines at left and down - curves from left-to-down or down-to-left
//   ╰ : lines at right and up - curves from right-to-up or up-to-right
//   ╯ : lines at left and up - curves from left-to-up or up-to-left
func connectionsToChar(conn uint8) rune {
	switch conn {
	case dirLeft | dirRight:
		return '─'
	case dirUp | dirDown:
		return '│'
	case dirLeft | dirDown:
		return '╮' // from left, curves down
	case dirLeft | dirUp:
		return '╯' // from left, curves up
	case dirRight | dirDown:
		return '╭' // from right, curves down
	case dirRight | dirUp:
		return '╰' // from right, curves up
	case dirUp | dirDown | dirLeft | dirRight:
		return '┼'
	case dirUp | dirDown | dirLeft:
		return '├'
	case dirUp | dirDown | dirRight:
		return '┤'
	case dirUp | dirLeft | dirRight:
		return '┬'
	case dirDown | dirLeft | dirRight:
		return '┴'
	default:
		return ' ' // no connections or invalid combination
	}
}

// addConnection adds a connection direction to a cell
func addConnection(canvas [][]plotCell, x, y int, dir uint8, color int, height, width int) {
	if x < 0 || x >= width || y < 0 || y >= height {
		return
	}
	cell := canvas[y][x]
	isNew := cell.connections == 0
	cell.connections |= dir
	if isNew {
		cell.color = color
	}
	canvas[y][x] = cell
}

// drawHorizontalLine draws a horizontal line from x0 to x1 at row y
func drawHorizontalLine(canvas [][]plotCell, x0, x1, y int, color int, height, width int) {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	for x := x0; x <= x1; x++ {
		if x == x0 {
			addConnection(canvas, x, y, dirRight, color, height, width)
		} else if x == x1 {
			addConnection(canvas, x, y, dirLeft, color, height, width)
		} else {
			addConnection(canvas, x, y, dirLeft|dirRight, color, height, width)
		}
	}
}

// drawVerticalLine draws a vertical line from y0 to y1 at column x
func drawVerticalLine(canvas [][]plotCell, x, y0, y1 int, color int, height, width int) {
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	for y := y0; y <= y1; y++ {
		if y == y0 {
			addConnection(canvas, x, y, dirDown, color, height, width)
		} else if y == y1 {
			addConnection(canvas, x, y, dirUp, color, height, width)
		} else {
			addConnection(canvas, x, y, dirUp|dirDown, color, height, width)
		}
	}
}

// asciiPlotDataToStr implements the plotDataToStr function
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

	// Plot each series
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

			// Y: map value to row (row 0 is top)
			yNorm := (v - cfg.minVal) / (cfg.maxVal - cfg.minVal)
			yPos := int(yNorm * float64(height-1))
			if yPos < 0 {
				yPos = 0
			}
			if yPos >= height {
				yPos = height - 1
			}
			points[i].y = height - 1 - yPos
		}

		// Draw lines between consecutive points
		for i := 0; i < n-1; i++ {
			p0 := points[i]
			p1 := points[i+1]

			if p0.x == p1.x {
				// Vertical line only
				drawVerticalLine(canvas, p0.x, p0.y, p1.y, seriesColor, height, width)
			} else if p0.y == p1.y {
				// Horizontal line only
				drawHorizontalLine(canvas, p0.x, p1.x, p0.y, seriesColor, height, width)
			} else {
				// Diagonal: draw horizontal then vertical with curve at junction
				dx := p1.x - p0.x
				dy := p1.y - p0.y

				// Split point - use position closer to p0 for the curve
				splitX := p0.x
				if dx > 0 {
					splitX = p0.x + 1
					if splitX > p1.x {
						splitX = p1.x
					}
				} else {
					splitX = p0.x - 1
					if splitX < p1.x {
						splitX = p1.x
					}
				}

				// Draw horizontal segment from p0 to splitX (exclusive of splitX)
				if dx > 0 {
					for x := p0.x; x < splitX; x++ {
						if x == p0.x {
							addConnection(canvas, x, p0.y, dirRight, seriesColor, height, width)
						} else {
							addConnection(canvas, x, p0.y, dirLeft|dirRight, seriesColor, height, width)
						}
					}
				} else {
					for x := p0.x; x > splitX; x-- {
						if x == p0.x {
							addConnection(canvas, x, p0.y, dirLeft, seriesColor, height, width)
						} else {
							addConnection(canvas, x, p0.y, dirLeft|dirRight, seriesColor, height, width)
						}
					}
				}

				// Draw curve at splitX
				// The curve connects horizontal direction and vertical direction
				var curveStartDir uint8
				if dx > 0 {
					curveStartDir = dirLeft // horizontal comes from left
				} else {
					curveStartDir = dirRight // horizontal comes from right
				}
				// The curve point has both horizontal and vertical connections
				// But we need to also add the continuation direction
				if dy > 0 {
					// Going down
					addConnection(canvas, splitX, p0.y, curveStartDir|dirDown, seriesColor, height, width)
					for y := p0.y + 1; y < p1.y; y++ {
						addConnection(canvas, splitX, y, dirUp|dirDown, seriesColor, height, width)
					}
					if splitX < p1.x {
						addConnection(canvas, splitX, p1.y, dirUp|dirRight, seriesColor, height, width)
					} else if splitX > p1.x {
						addConnection(canvas, splitX, p1.y, dirUp|dirLeft, seriesColor, height, width)
					} else {
						addConnection(canvas, splitX, p1.y, dirUp, seriesColor, height, width)
					}
				} else {
					// Going up
					addConnection(canvas, splitX, p0.y, curveStartDir|dirUp, seriesColor, height, width)
					for y := p0.y - 1; y > p1.y; y-- {
						addConnection(canvas, splitX, y, dirUp|dirDown, seriesColor, height, width)
					}
					if splitX < p1.x {
						addConnection(canvas, splitX, p1.y, dirDown|dirRight, seriesColor, height, width)
					} else if splitX > p1.x {
						addConnection(canvas, splitX, p1.y, dirDown|dirLeft, seriesColor, height, width)
					} else {
						addConnection(canvas, splitX, p1.y, dirDown, seriesColor, height, width)
					}
				}

				// Draw horizontal segment from splitX to p1 (exclusive of splitX)
				if dx > 0 {
					for x := splitX + 1; x <= p1.x; x++ {
						if x == p1.x {
							addConnection(canvas, x, p1.y, dirLeft, seriesColor, height, width)
						} else {
							addConnection(canvas, x, p1.y, dirLeft|dirRight, seriesColor, height, width)
						}
					}
				} else {
					for x := splitX - 1; x >= p1.x; x-- {
						if x == p1.x {
							addConnection(canvas, x, p1.y, dirRight, seriesColor, height, width)
						} else {
							addConnection(canvas, x, p1.y, dirLeft|dirRight, seriesColor, height, width)
						}
					}
				}
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
			char := connectionsToChar(cell.connections)
			if char != ' ' && cell.color > 0 {
				result.WriteString(ansiColor(cell.color))
				result.WriteRune(char)
				result.WriteString(ansiReset())
			} else if char != ' ' {
				result.WriteRune(char)
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