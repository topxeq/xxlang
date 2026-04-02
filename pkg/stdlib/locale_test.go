package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callLocaleFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("locale")
	if mod == nil {
		panic("locale module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestLocaleToPinYin(t *testing.T) {
	tests := []struct {
		name    string
		args    []objects.Object
		wantErr bool
	}{
		{
			name:    "basic chinese",
			args:    []objects.Object{String("中国")},
			wantErr: false,
		},
		{
			name:    "with tone option",
			args:    []objects.Object{String("中国"), String("-tone")},
			wantErr: false,
		},
		{
			name:    "with tone number option",
			args:    []objects.Object{String("中国"), String("-toneNum")},
			wantErr: false,
		},
		{
			name:    "with initials option",
			args:    []objects.Object{String("中国"), String("-initials")},
			wantErr: false,
		},
		{
			name:    "with separator option",
			args:    []objects.Object{String("中国"), String("-sep=-")},
			wantErr: false,
		},
		{
			name:    "empty string",
			args:    []objects.Object{String("")},
			wantErr: false,
		},
		{
			name:    "no arguments",
			args:    []objects.Object{},
			wantErr: true,
		},
		{
			name:    "wrong type argument",
			args:    []objects.Object{objects.NewInt(123)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := callLocaleFunc("toPinYin", tt.args...)
			if tt.wantErr {
				if _, ok := result.(*objects.Error); !ok {
					t.Errorf("expected error, got %T", result)
				}
			} else {
				if _, ok := result.(*objects.String); !ok {
					t.Errorf("expected string, got %T", result)
				}
			}
		})
	}
}

func TestLocaleKanaToRomaji(t *testing.T) {
	tests := []struct {
		name    string
		args    []objects.Object
		wantErr bool
	}{
		{
			name:    "basic hiragana",
			args:    []objects.Object{String("こんにちは")},
			wantErr: false,
		},
		{
			name:    "empty string",
			args:    []objects.Object{String("")},
			wantErr: false,
		},
		{
			name:    "no arguments",
			args:    []objects.Object{},
			wantErr: true,
		},
		{
			name:    "wrong type argument",
			args:    []objects.Object{objects.NewInt(123)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := callLocaleFunc("kanaToRomaji", tt.args...)
			if tt.wantErr {
				if _, ok := result.(*objects.Error); !ok {
					t.Errorf("expected error, got %T", result)
				}
			} else {
				if _, ok := result.(*objects.String); !ok {
					t.Errorf("expected string, got %T", result)
				}
			}
		})
	}
}

func TestLocaleKanjiToKana(t *testing.T) {
	tests := []struct {
		name    string
		args    []objects.Object
		wantErr bool
	}{
		{
			name:    "basic kanji",
			args:    []objects.Object{String("日本語")},
			wantErr: false,
		},
		{
			name:    "empty string",
			args:    []objects.Object{String("")},
			wantErr: false,
		},
		{
			name:    "no arguments",
			args:    []objects.Object{},
			wantErr: true,
		},
		{
			name:    "wrong type argument",
			args:    []objects.Object{objects.NewInt(123)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := callLocaleFunc("kanjiToKana", tt.args...)
			if tt.wantErr {
				if _, ok := result.(*objects.Error); !ok {
					t.Errorf("expected error, got %T", result)
				}
			} else {
				if _, ok := result.(*objects.String); !ok {
					t.Errorf("expected string, got %T", result)
				}
			}
		})
	}
}

func TestLocaleKanjiToRomaji(t *testing.T) {
	tests := []struct {
		name    string
		args    []objects.Object
		wantErr bool
	}{
		{
			name:    "basic kanji",
			args:    []objects.Object{String("日本語")},
			wantErr: false,
		},
		{
			name:    "empty string",
			args:    []objects.Object{String("")},
			wantErr: false,
		},
		{
			name:    "no arguments",
			args:    []objects.Object{},
			wantErr: true,
		},
		{
			name:    "wrong type argument",
			args:    []objects.Object{objects.NewInt(123)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := callLocaleFunc("kanjiToRomaji", tt.args...)
			if tt.wantErr {
				if _, ok := result.(*objects.Error); !ok {
					t.Errorf("expected error, got %T", result)
				}
			} else {
				if _, ok := result.(*objects.String); !ok {
					t.Errorf("expected string, got %T", result)
				}
			}
		})
	}
}
