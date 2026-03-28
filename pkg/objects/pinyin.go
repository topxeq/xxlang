// pkg/objects/pinyin.go
// Chinese character to Pinyin converter - pure Go implementation
package objects

import (
	"strings"
	"unicode"
)

// ToneStyle defines how to represent tones in pinyin output
type ToneStyle int

const (
	ToneNone   ToneStyle = iota // No tone (default): zhong
	ToneMark                    // Tone marks: zhōng
	ToneNumber                  // Tone numbers: zhong1
)

// ConvertToPinyin converts Chinese characters to Pinyin
// withTone: if true, returns pinyin with tone marks (deprecated, use ConvertToPinyinEx)
// initialsOnly: if true, returns only the first letter (initial)
func ConvertToPinyin(text string, withTone bool, initialsOnly bool) string {
	return ConvertToPinyinEx(text, ToneMark, initialsOnly)
}

// ConvertToPinyinEx converts Chinese characters to Pinyin with extended options
// toneStyle: how to represent tones (ToneNone, ToneMark, ToneNumber)
// initialsOnly: if true, returns only the first letter (initial)
func ConvertToPinyinEx(text string, toneStyle ToneStyle, initialsOnly bool) string {
	var result []string

	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			if pyList, ok := pinyinMap[r]; ok && len(pyList) > 0 {
				var py string
				switch toneStyle {
				case ToneMark:
					if len(pyList) > 1 {
						py = pyList[1] // with tone marks
					} else {
						py = pyList[0]
					}
				case ToneNumber:
					if len(pyList) > 2 {
						py = pyList[2] // with tone numbers
					} else if len(pyList) > 0 {
						py = pyList[0]
					}
				default: // ToneNone
					py = pyList[0] // plain
				}

				if initialsOnly && len(py) > 0 {
					py = string(py[0])
				}

				result = append(result, py)
			} else {
				result = append(result, string(r))
			}
		} else {
			result = append(result, string(r))
		}
	}

	return strings.Join(result, "")
}

// ConvertToPinyinWithSeparator converts Chinese characters to Pinyin with a separator
func ConvertToPinyinWithSeparator(text string, withTone bool, initialsOnly bool, separator string) string {
	return ConvertToPinyinWithSeparatorEx(text, ToneMark, initialsOnly, separator)
}

// ConvertToPinyinWithSeparatorEx converts Chinese characters to Pinyin with extended options
func ConvertToPinyinWithSeparatorEx(text string, toneStyle ToneStyle, initialsOnly bool, separator string) string {
	var result []string

	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			if pyList, ok := pinyinMap[r]; ok && len(pyList) > 0 {
				var py string
				switch toneStyle {
				case ToneMark:
					if len(pyList) > 1 {
						py = pyList[1]
					} else {
						py = pyList[0]
					}
				case ToneNumber:
					if len(pyList) > 2 {
						py = pyList[2]
					} else if len(pyList) > 0 {
						py = pyList[0]
					}
				default:
					py = pyList[0]
				}

				if initialsOnly && len(py) > 0 {
					py = string(py[0])
				}

				result = append(result, py)
			} else {
				result = append(result, string(r))
			}
		} else {
			result = append(result, string(r))
		}
	}

	return strings.Join(result, separator)
}

// GetPinyinInitials returns the initials (first letters) of pinyin
func GetPinyinInitials(text string) string {
	return ConvertToPinyinEx(text, ToneNone, true)
}

// GetPinyinWithTone returns pinyin with tone marks
func GetPinyinWithTone(text string) string {
	return ConvertToPinyinEx(text, ToneMark, false)
}

// GetPinyinWithToneNumber returns pinyin with tone numbers (e.g., zhong1guo2)
func GetPinyinWithToneNumber(text string) string {
	return ConvertToPinyinEx(text, ToneNumber, false)
}
