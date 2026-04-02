package objects

import (
	"testing"
)

func TestConvertToPinyin(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		withTone     bool
		initialsOnly bool
		want         string
	}{
		{
			name:         "simple chinese no tone",
			input:        "中国",
			withTone:     false,
			initialsOnly: false,
			want:         "zhōngguó",
		},
		{
			name:         "chinese with tone",
			input:        "中国",
			withTone:     true,
			initialsOnly: false,
			want:         "zhōngguó",
		},
		{
			name:         "chinese initials only",
			input:        "中国",
			withTone:     false,
			initialsOnly: true,
			want:         "zg",
		},
		{
			name:         "mixed chinese and english",
			input:        "hello世界",
			withTone:     false,
			initialsOnly: false,
			want:         "helloshìjiè",
		},
		{
			name:         "empty string",
			input:        "",
			withTone:     false,
			initialsOnly: false,
			want:         "",
		},
		{
			name:         "only ascii",
			input:        "hello world",
			withTone:     false,
			initialsOnly: false,
			want:         "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertToPinyin(tt.input, tt.withTone, tt.initialsOnly)
			if got != tt.want {
				t.Errorf("ConvertToPinyin(%q, %v, %v) = %q, want %q", tt.input, tt.withTone, tt.initialsOnly, got, tt.want)
			}
		})
	}
}

func TestConvertToPinyinEx(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		toneStyle    ToneStyle
		initialsOnly bool
		want         string
	}{
		{
			name:         "tone none",
			input:        "中国",
			toneStyle:    ToneNone,
			initialsOnly: false,
			want:         "zhongguo",
		},
		{
			name:         "tone mark",
			input:        "中国",
			toneStyle:    ToneMark,
			initialsOnly: false,
			want:         "zhōngguó",
		},
		{
			name:         "tone number",
			input:        "中国",
			toneStyle:    ToneNumber,
			initialsOnly: false,
			want:         "zhong1guo2",
		},
		{
			name:         "initials only with tone none",
			input:        "北京",
			toneStyle:    ToneNone,
			initialsOnly: true,
			want:         "bj",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertToPinyinEx(tt.input, tt.toneStyle, tt.initialsOnly)
			if got != tt.want {
				t.Errorf("ConvertToPinyinEx(%q, %v, %v) = %q, want %q", tt.input, tt.toneStyle, tt.initialsOnly, got, tt.want)
			}
		})
	}
}

func TestConvertToPinyinWithSeparator(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		withTone     bool
		initialsOnly bool
		separator    string
		want         string
	}{
		{
			name:         "space separator",
			input:        "中国",
			withTone:     false,
			initialsOnly: false,
			separator:    " ",
			want:         "zhōng guó",
		},
		{
			name:         "dash separator",
			input:        "北京",
			withTone:     false,
			initialsOnly: false,
			separator:    "-",
			want:         "běi-jīng",
		},
		{
			name:         "underscore separator initials",
			input:        "上海",
			withTone:     false,
			initialsOnly: true,
			separator:    "_",
			want:         "s_h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertToPinyinWithSeparator(tt.input, tt.withTone, tt.initialsOnly, tt.separator)
			if got != tt.want {
				t.Errorf("ConvertToPinyinWithSeparator(%q, %v, %v, %q) = %q, want %q",
					tt.input, tt.withTone, tt.initialsOnly, tt.separator, got, tt.want)
			}
		})
	}
}

func TestConvertToPinyinWithSeparatorEx(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		toneStyle    ToneStyle
		initialsOnly bool
		separator    string
		want         string
	}{
		{
			name:         "tone mark with separator",
			input:        "中国",
			toneStyle:    ToneMark,
			initialsOnly: false,
			separator:    " ",
			want:         "zhōng guó",
		},
		{
			name:         "tone number with separator",
			input:        "中国",
			toneStyle:    ToneNumber,
			initialsOnly: false,
			separator:    "-",
			want:         "zhong1-guo2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertToPinyinWithSeparatorEx(tt.input, tt.toneStyle, tt.initialsOnly, tt.separator)
			if got != tt.want {
				t.Errorf("ConvertToPinyinWithSeparatorEx(%q, %v, %v, %q) = %q, want %q",
					tt.input, tt.toneStyle, tt.initialsOnly, tt.separator, got, tt.want)
			}
		})
	}
}

func TestGetPinyinInitials(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"中国", "zg"},
		{"北京", "bj"},
		{"上海", "sh"},
		{"hello世界", "hellosj"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := GetPinyinInitials(tt.input)
			if got != tt.want {
				t.Errorf("GetPinyinInitials(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetPinyinWithTone(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"中国", "zhōngguó"},
		{"北京", "běijīng"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := GetPinyinWithTone(tt.input)
			if got != tt.want {
				t.Errorf("GetPinyinWithTone(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetPinyinWithToneNumber(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"中国", "zhong1guo2"},
		{"北京", "bei3jing1"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := GetPinyinWithToneNumber(tt.input)
			if got != tt.want {
				t.Errorf("GetPinyinWithToneNumber(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
