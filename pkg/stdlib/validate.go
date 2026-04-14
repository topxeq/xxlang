// pkg/stdlib/validate.go
// Validation utilities for the Xxlang standard library.
package stdlib

import (
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "validate",
		Exports: map[string]objects.Object{
			// Check if string is valid email
			"isEmail": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isEmail() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isEmail() requires a string argument")
				}
				_, err := mail.ParseAddress(s.Value)
				return Bool(err == nil)
			}),

			// Check if string is valid URL
			"isURL": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isURL() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isURL() requires a string argument")
				}
				url := s.Value
				if len(url) < 3 {
					return Bool(false)
				}
				// Simple URL validation
				hasScheme := strings.HasPrefix(url, "http://") ||
					strings.HasPrefix(url, "https://") ||
					strings.HasPrefix(url, "ftp://") ||
					strings.HasPrefix(url, "ws://") ||
					strings.HasPrefix(url, "wss://")
				return Bool(hasScheme && strings.Contains(url, "."))
			}),

			// Check if string matches regex pattern
			"matches": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("matches() takes exactly 2 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("matches() requires a string as first argument")
				}
				pattern, ok := args[1].(*objects.String)
				if !ok {
					return Error("matches() requires a string pattern")
				}
				re, err := regexp.Compile(pattern.Value)
				if err != nil {
					return Error(err.Error())
				}
				return Bool(re.MatchString(s.Value))
			}),

			// Check if string length is in range
			"lengthRange": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("lengthRange() takes exactly 3 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("lengthRange() requires a string as first argument")
				}
				min, ok := args[1].(*objects.Int)
				if !ok {
					return Error("lengthRange() requires an integer min")
				}
				max, ok := args[2].(*objects.Int)
				if !ok {
					return Error("lengthRange() requires an integer max")
				}
				length := int64(utf8.RuneCountInString(s.Value))
				return Bool(length >= min.Value && length <= max.Value)
			}),

			// Check if string is not empty
			"required": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("required() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("required() requires a string argument")
				}
				return Bool(strings.TrimSpace(s.Value) != "")
			}),

			// Check if value is in array
			"inArray": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("inArray() takes exactly 2 arguments")
				}
				val := args[0]
				arr, ok := args[1].(*objects.Array)
				if !ok {
					return Error("inArray() requires an array as second argument")
				}
				for _, elem := range arr.Elements {
					if elem.Inspect() == val.Inspect() {
						return Bool(true)
					}
				}
				return Bool(false)
			}),

			// Check if value is not in array
			"notInArray": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("notInArray() takes exactly 2 arguments")
				}
				val := args[0]
				arr, ok := args[1].(*objects.Array)
				if !ok {
					return Error("notInArray() requires an array as second argument")
				}
				for _, elem := range arr.Elements {
					if elem.Inspect() == val.Inspect() {
						return Bool(false)
					}
				}
				return Bool(true)
			}),

			// Check if number is in range
			"inRange": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("inRange() takes exactly 3 arguments")
				}
				var num float64
				switch v := args[0].(type) {
				case *objects.Int:
					num = float64(v.Value)
				case *objects.Float:
					num = v.Value
				default:
					return Error("inRange() requires a number as first argument")
				}
				min, ok := args[1].(*objects.Int)
				if !ok {
					if f, ok := args[1].(*objects.Float); ok {
						return Bool(num >= f.Value)
					}
					return Error("inRange() requires a number min")
				}
				max, ok := args[2].(*objects.Int)
				if !ok {
					if f, ok := args[2].(*objects.Float); ok {
						return Bool(num >= float64(min.Value) && num <= f.Value)
					}
					return Error("inRange() requires a number max")
				}
				return Bool(num >= float64(min.Value) && num <= float64(max.Value))
			}),

			// Check if string is valid JSON
			"isJSON": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isJSON() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isJSON() requires a string argument")
				}
				str := strings.TrimSpace(s.Value)
				if len(str) == 0 {
					return Bool(false)
				}
				first := str[0]
				last := str[len(str)-1]
				return Bool((first == '{' && last == '}') || (first == '[' && last == ']'))
			}),

			// Check if string is alphanumeric
			"isAlphanumeric": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isAlphanumeric() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isAlphanumeric() requires a string argument")
				}
				if len(s.Value) == 0 {
					return Bool(false)
				}
				for _, r := range s.Value {
					if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
						return Bool(false)
					}
				}
				return Bool(true)
			}),

			// Check if string is alphabetic
			"isAlpha": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isAlpha() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isAlpha() requires a string argument")
				}
				if len(s.Value) == 0 {
					return Bool(false)
				}
				for _, r := range s.Value {
					if !unicode.IsLetter(r) {
						return Bool(false)
					}
				}
				return Bool(true)
			}),

			// Check if string is numeric
			"isNumeric": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isNumeric() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isNumeric() requires a string argument")
				}
				if len(s.Value) == 0 {
					return Bool(false)
				}
				_, err := strconv.ParseFloat(s.Value, 64)
				return Bool(err == nil)
			}),

			// Check if string is integer
			"isInteger": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isInteger() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isInteger() requires a string argument")
				}
				if len(s.Value) == 0 {
					return Bool(false)
				}
				_, err := strconv.ParseInt(s.Value, 10, 64)
				return Bool(err == nil)
			}),

			// Check if string is a valid hex color
			"isHexColor": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isHexColor() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isHexColor() requires a string argument")
				}
				str := s.Value
				if len(str) != 4 && len(str) != 7 {
					return Bool(false)
				}
				if str[0] != '#' {
					return Bool(false)
				}
				for i := 1; i < len(str); i++ {
					c := str[i]
					if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
						return Bool(false)
					}
				}
				return Bool(true)
			}),

			// Check if string is valid UUID
			"isUUID": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isUUID() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isUUID() requires a string argument")
				}
				uuid := s.Value
				if len(uuid) != 36 {
					return Bool(false)
				}
				if uuid[8] != '-' || uuid[13] != '-' || uuid[18] != '-' || uuid[23] != '-' {
					return Bool(false)
				}
				for i, c := range uuid {
					if i == 8 || i == 13 || i == 18 || i == 23 {
						continue
					}
					if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
						return Bool(false)
					}
				}
				return Bool(true)
			}),

			// Check if string is valid IP address (IPv4)
			"isIPv4": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isIPv4() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isIPv4() requires a string argument")
				}
				parts := strings.Split(s.Value, ".")
				if len(parts) != 4 {
					return Bool(false)
				}
				for _, part := range parts {
					num, err := strconv.Atoi(part)
					if err != nil || num < 0 || num > 255 {
						return Bool(false)
					}
					if len(part) > 1 && part[0] == '0' {
						return Bool(false)
					}
				}
				return Bool(true)
			}),

			// Check if string is valid phone number (simple)
			"isPhone": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isPhone() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isPhone() requires a string argument")
				}
				// Remove common separators
				phone := strings.ReplaceAll(s.Value, " ", "")
				phone = strings.ReplaceAll(phone, "-", "")
				phone = strings.ReplaceAll(phone, "(", "")
				phone = strings.ReplaceAll(phone, ")", "")
				phone = strings.ReplaceAll(phone, "+", "")
				// Check if all digits
				if len(phone) < 7 || len(phone) > 15 {
					return Bool(false)
				}
				for _, c := range phone {
					if !unicode.IsDigit(c) {
						return Bool(false)
					}
				}
				return Bool(true)
			}),

			// Check if string is valid date (YYYY-MM-DD)
			"isDate": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isDate() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isDate() requires a string argument")
				}
				date := s.Value
				if len(date) != 10 {
					return Bool(false)
				}
				if date[4] != '-' || date[7] != '-' {
					return Bool(false)
				}
				year, err1 := strconv.Atoi(date[0:4])
				month, err2 := strconv.Atoi(date[5:7])
				day, err3 := strconv.Atoi(date[8:10])
				if err1 != nil || err2 != nil || err3 != nil {
					return Bool(false)
				}
				return Bool(year >= 1000 && year <= 9999 && month >= 1 && month <= 12 && day >= 1 && day <= 31)
			}),

			// Check if string is valid time (HH:MM or HH:MM:SS)
			"isTime": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isTime() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isTime() requires a string argument")
				}
				time := s.Value
				parts := strings.Split(time, ":")
				if len(parts) != 2 && len(parts) != 3 {
					return Bool(false)
				}
				hour, err1 := strconv.Atoi(parts[0])
				min, err2 := strconv.Atoi(parts[1])
				if err1 != nil || err2 != nil {
					return Bool(false)
				}
				if hour < 0 || hour > 23 || min < 0 || min > 59 {
					return Bool(false)
				}
				if len(parts) == 3 {
					sec, err := strconv.Atoi(parts[2])
					if err != nil || sec < 0 || sec > 59 {
						return Bool(false)
					}
				}
				return Bool(true)
			}),

			// Check if string starts with
			"startsWith": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("startsWith() takes exactly 2 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("startsWith() requires a string as first argument")
				}
				prefix, ok := args[1].(*objects.String)
				if !ok {
					return Error("startsWith() requires a string prefix")
				}
				return Bool(strings.HasPrefix(s.Value, prefix.Value))
			}),

			// Check if string ends with
			"endsWith": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("endsWith() takes exactly 2 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("endsWith() requires a string as first argument")
				}
				suffix, ok := args[1].(*objects.String)
				if !ok {
					return Error("endsWith() requires a string suffix")
				}
				return Bool(strings.HasSuffix(s.Value, suffix.Value))
			}),

			// Check if string contains
			"contains": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("contains() takes exactly 2 arguments")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("contains() requires a string as first argument")
				}
				substr, ok := args[1].(*objects.String)
				if !ok {
					return Error("contains() requires a string substring")
				}
				return Bool(strings.Contains(s.Value, substr.Value))
			}),

			// Check if string is a valid credit card number (Luhn algorithm)
			"isCreditCard": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isCreditCard() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("isCreditCard() requires a string argument")
				}
				// Remove spaces and dashes
				num := strings.ReplaceAll(s.Value, " ", "")
				num = strings.ReplaceAll(num, "-", "")
				if len(num) < 13 || len(num) > 19 {
					return Bool(false)
				}
				// Check if all digits
				for _, c := range num {
					if !unicode.IsDigit(c) {
						return Bool(false)
					}
				}
				// Luhn algorithm
				sum := 0
				alt := false
				for i := len(num) - 1; i >= 0; i-- {
					digit := int(num[i] - '0')
					if alt {
						digit *= 2
						if digit > 9 {
							digit -= 9
						}
					}
					sum += digit
					alt = !alt
				}
				return Bool(sum%10 == 0)
			}),
		},
	})
}
