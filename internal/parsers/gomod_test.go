package parsers

import (
	"reflect"
	"testing"
)

func TestGoParser(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want ParseResult
	}{
		{
			name: "go get pinned",
			args: []string{"get", "github.com/dgrijalva/jwt-go@v3.2.0"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "Go", Name: "github.com/dgrijalva/jwt-go", Version: "v3.2.0"}}, Reason: "go get"},
		},
		{
			name: "go install @latest → unpinned",
			args: []string{"install", "github.com/foo/cmd@latest"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "Go", Name: "github.com/foo/cmd"}}, Reason: "go install"},
		},
		{
			name: "go get unpinned",
			args: []string{"get", "golang.org/x/text"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "Go", Name: "golang.org/x/text"}}, Reason: "go get"},
		},
		{
			name: "go get ./... → passthrough (no registry pkgs)",
			args: []string{"get", "./..."},
			want: ParseResult{},
		},
		{
			name: "go get all → passthrough",
			args: []string{"get", "all"},
			want: ParseResult{},
		},
		{
			name: "go build → passthrough",
			args: []string{"build", "./..."},
			want: ParseResult{},
		},
		{
			name: "go test → passthrough",
			args: []string{"test", "./..."},
			want: ParseResult{},
		},
		{
			name: "go bare → passthrough",
			args: []string{},
			want: ParseResult{},
		},
	}

	p := Go()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := p.Parse(tc.args)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parse(%v):\n  got:  %+v\n  want: %+v", tc.args, got, tc.want)
			}
		})
	}
}
