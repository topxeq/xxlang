package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callLeFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("le")
	if mod == nil {
		panic("le module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestLe_Create(t *testing.T) {
	result := callLeFunc("create")
	if _, ok := result.(*objects.LineEditor); !ok {
		t.Errorf("expected *objects.LineEditor, got %T", result)
	}
}

func TestLe_FromText(t *testing.T) {
	tests := []struct {
		name    string
		args    []objects.Object
		wantErr bool
	}{
		{
			name:    "basic text",
			args:    []objects.Object{String("line1\nline2\nline3")},
			wantErr: false,
		},
		{
			name:    "empty text",
			args:    []objects.Object{String("")},
			wantErr: false,
		},
		{
			name:    "no arguments",
			args:    []objects.Object{},
			wantErr: true,
		},
		{
			name:    "wrong type",
			args:    []objects.Object{objects.NewInt(123)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := callLeFunc("fromText", tt.args...)
			if tt.wantErr {
				if _, ok := result.(*objects.Error); !ok {
					t.Errorf("expected error, got %T", result)
				}
			} else {
				if _, ok := result.(*objects.LineEditor); !ok {
					t.Errorf("expected LineEditor, got %T", result)
				}
			}
		})
	}
}

func TestLe_FromLines(t *testing.T) {
	tests := []struct {
		name    string
		args    []objects.Object
		wantErr bool
	}{
		{
			name: "string array",
			args: []objects.Object{&objects.Array{Elements: []objects.Object{
				String("line1"),
				String("line2"),
			}}},
			wantErr: false,
		},
		{
			name:    "empty array",
			args:    []objects.Object{&objects.Array{Elements: []objects.Object{}}},
			wantErr: false,
		},
		{
			name:    "no arguments",
			args:    []objects.Object{},
			wantErr: true,
		},
		{
			name:    "wrong type",
			args:    []objects.Object{objects.NewInt(123)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := callLeFunc("fromLines", tt.args...)
			if tt.wantErr {
				if _, ok := result.(*objects.Error); !ok {
					t.Errorf("expected error, got %T", result)
				}
			} else {
				if _, ok := result.(*objects.LineEditor); !ok {
					t.Errorf("expected LineEditor, got %T", result)
				}
			}
		})
	}
}

func TestLe_IsLineEditor(t *testing.T) {
	le := callLeFunc("create")

	result := callLeFunc("isLineEditor", le)
	if b, ok := result.(*objects.Bool); !ok || !b.Value {
		t.Error("isLineEditor should return true for LineEditor")
	}

	result = callLeFunc("isLineEditor", objects.NewInt(123))
	if b, ok := result.(*objects.Bool); !ok || b.Value {
		t.Error("isLineEditor should return false for non-LineEditor")
	}

	result = callLeFunc("isLineEditor")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for missing argument")
	}
}

func TestLe_OpenNonexistent(t *testing.T) {
	result := callLeFunc("open", String("/nonexistent/file.txt"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for nonexistent file")
	}
}

func TestLe_OpenWrongArgs(t *testing.T) {
	result := callLeFunc("open")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for missing argument")
	}

	result = callLeFunc("open", objects.NewInt(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for wrong type")
	}
}

func TestLe_ReplaceInFileWrongArgs(t *testing.T) {
	result := callLeFunc("replaceInFile")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for missing arguments")
	}

	result = callLeFunc("replaceInFile", String("path"), String("old"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for missing third argument")
	}

	result = callLeFunc("replaceInFile", objects.NewInt(123), String("old"), String("new"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for wrong type")
	}
}

func TestLe_SortFileWrongArgs(t *testing.T) {
	result := callLeFunc("sortFile")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for missing argument")
	}

	result = callLeFunc("sortFile", objects.NewInt(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for wrong type")
	}
}

func TestLe_UniqueFileWrongArgs(t *testing.T) {
	result := callLeFunc("uniqueFile")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for missing argument")
	}

	result = callLeFunc("uniqueFile", objects.NewInt(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for wrong type")
	}
}

func TestLe_CountLinesWrongArgs(t *testing.T) {
	result := callLeFunc("countLines")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for missing argument")
	}

	result = callLeFunc("countLines", objects.NewInt(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for wrong type")
	}
}

func TestLe_HeadWrongArgs(t *testing.T) {
	result := callLeFunc("head")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for missing arguments")
	}

	result = callLeFunc("head", String("path"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for missing second argument")
	}
}

func TestLe_TailWrongArgs(t *testing.T) {
	result := callLeFunc("tail")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for missing arguments")
	}

	result = callLeFunc("tail", String("path"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for missing second argument")
	}
}

func TestLe_GrepFileWrongArgs(t *testing.T) {
	result := callLeFunc("grepFile")
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for missing arguments")
	}

	result = callLeFunc("grepFile", String("path"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for missing pattern")
	}
}
