package boilermaker

import (
	"testing"

	"github.com/josharian/txtarfs"
	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/txtar"
)

func TestBoilerplate(t *testing.T) {
	tests := []struct {
		name         string
		files        string
		wantMetadata Metadata
		data         map[string]any
		want         Result
		wantErr      bool
	}{
		{
			name: "ok",
			files: `
-- example --
hello
`,
			want: Result{
				"example": "hello\n",
			},
		},
		{
			name: "with template data",
			data: map[string]any{
				"MyVariable": "something",
			},
			files: `
-- example --
{{ .MyVariable }}
`,
			want: Result{
				"example": "something\n",
			},
		},
		{
			name: "with templated file name",
			data: map[string]any{
				"MyVariable": "something",
			},
			files: `
-- {{ .MyVariable }}.go --
{{ .MyVariable }}
`,
			want: Result{
				"something.go": "something\n",
			},
		},
		{
			name: "with metadata",
			data: map[string]any{
				"MyVariable": "something",
			},
			files: `
-- _boilermaker.json --
{
	"description": "an example boilerplate"
}
`,
			wantMetadata: Metadata{
				Description: "an example boilerplate",
			},
			want: Result{},
		},
		{
			name: "empty file",
			files: `
-- __init__.py --`,
			want: Result{
				"__init__.py": "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ar := txtar.Parse([]byte(tt.files))

			arfs := txtarfs.As(ar)

			bp, err := ParseFS(arfs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.wantMetadata, bp.Metadata)

			got, err := bp.Generate(tt.data)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBoilerplateSet(t *testing.T) {
	tests := []struct {
		name    string
		files   string
		data    map[string]any
		want    map[string]Result
		wantErr bool
	}{
		{
			name: "ok",
			data: map[string]any{
				"ReplaceMe": "replaced",
			},
			files: `
-- one/example --
{{ .ReplaceMe }}
-- two/{{.ReplaceMe}} --
test
`,
			want: map[string]Result{
				"one": {
					"example": "replaced\n",
				},
				"two": {
					"replaced": "test\n",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ar := txtar.Parse([]byte(tt.files))

			arfs := txtarfs.As(ar)

			bpMap, err := ParseMapFS(arfs, ".")
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := map[string]Result{}

			for k, bp := range bpMap {
				result, err := bp.Generate(tt.data)
				if err != nil {
					t.Fatal(err)
				}
				got[k] = result
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
