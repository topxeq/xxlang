// pkg/objects/builtin_unicode_test.go
package objects

import (
	"testing"
)

func TestToPinYin(t *testing.T) {
	tests := []struct {
		name     string
		args     []Object
		expected string
	}{
		{
			name:     "simple chinese",
			args:     []Object{NewString("你好")},
			expected: "nihao",
		},
		{
			name:     "chinese with number",
			args:     []Object{NewString("北京2024")},
			expected: "beijing2024",
		},
		{
			name:     "mixed string",
			args:     []Object{NewString("中国China")},
			expected: "zhongguoChina",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builtinToPinYin(tt.args...)
			if str, ok := result.(*String); ok {
				if str.Value != tt.expected {
					t.Errorf("toPinYin() = %v, want %v", str.Value, tt.expected)
				}
			} else if err, ok := result.(*Error); ok {
				t.Errorf("toPinYin() returned error: %v", err.Message)
			} else {
				t.Errorf("toPinYin() returned unexpected type: %T", result)
			}
		})
	}
}

func TestKanaToRomaji(t *testing.T) {
	tests := []struct {
		name     string
		args     []Object
		expected string
	}{
		{
			name:     "hiragana basic",
			args:     []Object{NewString("こんにちは")},
			expected: "konnichiha",
		},
		{
			name:     "katakana basic",
			args:     []Object{NewString("コンニチハ")},
			expected: "konnichiha",
		},
		{
			name:     "mixed kana",
			args:     []Object{NewString("ありがとう")},
			expected: "arigatou",
		},
		{
			name:     "with combined",
			args:     []Object{NewString("きゃっと")},
			expected: "kyatto",
		},
		{
			name:     "simple word",
			args:     []Object{NewString("さくら")},
			expected: "sakura",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builtinKanaToRomaji(tt.args...)
			if str, ok := result.(*String); ok {
				if str.Value != tt.expected {
					t.Errorf("kanaToRomaji() = %v, want %v", str.Value, tt.expected)
				}
			} else if err, ok := result.(*Error); ok {
				t.Errorf("kanaToRomaji() returned error: %v", err.Message)
			} else {
				t.Errorf("kanaToRomaji() returned unexpected type: %T", result)
			}
		})
	}
}

func TestKanjiToKana(t *testing.T) {
	tests := []struct {
		name     string
		args     []Object
		expected string
	}{
		{
			name:     "numbers",
			args:     []Object{NewString("一二三")},
			expected: "いちにさん",
		},
		{
			name:     "japan compound",
			args:     []Object{NewString("日本")},
			expected: "にほん",
		},
		{
			name:     "tokyo",
			args:     []Object{NewString("東京")},
			expected: "とうきょう",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builtinKanjiToKana(tt.args...)
			if str, ok := result.(*String); ok {
				if str.Value != tt.expected {
					t.Errorf("kanjiToKana() = %v, want %v", str.Value, tt.expected)
				}
			} else if err, ok := result.(*Error); ok {
				t.Errorf("kanjiToKana() returned error: %v", err.Message)
			} else {
				t.Errorf("kanjiToKana() returned unexpected type: %T", result)
			}
		})
	}
}

func TestKanjiToRomaji(t *testing.T) {
	tests := []struct {
		name     string
		args     []Object
		expected string
	}{
		{
			name:     "japan compound",
			args:     []Object{NewString("日本")},
			expected: "nihon",
		},
		{
			name:     "tokyo",
			args:     []Object{NewString("東京")},
			expected: "toukyou",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builtinKanjiToRomaji(tt.args...)
			if str, ok := result.(*String); ok {
				if str.Value != tt.expected {
					t.Errorf("kanjiToRomaji() = %v, want %v", str.Value, tt.expected)
				}
			} else if err, ok := result.(*Error); ok {
				t.Errorf("kanjiToRomaji() returned error: %v", err.Message)
			} else {
				t.Errorf("kanjiToRomaji() returned unexpected type: %T", result)
			}
		})
	}
}
