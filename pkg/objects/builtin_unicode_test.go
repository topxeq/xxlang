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

func TestKanaToRomajiExported(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"あ", "a"},
		{"い", "i"},
		{"う", "u"},
		{"え", "e"},
		{"お", "o"},
		{"か", "ka"},
		{"き", "ki"},
		{"さ", "sa"},
		{"し", "shi"},
		{"た", "ta"},
		{"ち", "chi"},
		{"ん", "n"},
		{"ア", "a"},
		{"イ", "i"},
		{"カ", "ka"},
		{"キ", "ki"},
		{"ガ", "ga"},
		{"パ", "pa"},
		{"きゃ", "kya"},
		{"しゃ", "sha"},
		{"ちゃ", "cha"},
		{"hello", "hello"},
		{"", ""},
	}

	for _, tt := range tests {
		result := KanaToRomaji(tt.input)
		if result != tt.expected {
			t.Errorf("KanaToRomaji(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestKanjiToKanaExported(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"一", "いち"},
		{"二", "に"},
		{"三", "さん"},
		{"日本", "にほん"},
		{"東京", "とうきょう"},
		{"人", "じん"},
		{"日", "にち"},
		{"hello", "hello"},
		{"", ""},
	}

	for _, tt := range tests {
		result := KanjiToKana(tt.input)
		if result != tt.expected {
			t.Errorf("KanjiToKana(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestKanjiToRomajiExported(t *testing.T) {
	tests := []struct {
		input    string
		contains string
	}{
		{"一", "ichi"},
		{"二", "ni"},
		{"三", "san"},
		{"日本", "nihon"},
		{"東京", "toukyou"},
	}

	for _, tt := range tests {
		result := KanjiToRomaji(tt.input)
		if result == "" {
			t.Errorf("KanjiToRomaji(%q) returned empty string", tt.input)
		}
	}
}

func TestBuiltinKanaToRomajiErrors(t *testing.T) {
	result := builtinKanaToRomaji()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = builtinKanaToRomaji(NewInt(1))
	if !isError(result) {
		t.Error("expected error for wrong type")
	}
}

func TestBuiltinKanjiToKanaErrors(t *testing.T) {
	result := builtinKanjiToKana()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = builtinKanjiToKana(NewInt(1))
	if !isError(result) {
		t.Error("expected error for wrong type")
	}
}

func TestBuiltinKanjiToRomajiErrors(t *testing.T) {
	result := builtinKanjiToRomaji()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = builtinKanjiToRomaji(NewInt(1))
	if !isError(result) {
		t.Error("expected error for wrong type")
	}
}

func TestBuiltinToPinYinErrors(t *testing.T) {
	result := builtinToPinYin()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	result = builtinToPinYin(NewInt(1))
	if !isError(result) {
		t.Error("expected error for wrong type")
	}
}
