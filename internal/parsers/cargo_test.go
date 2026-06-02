package parsers

import (
	"reflect"
	"testing"
)

func TestCargoParser(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want ParseResult
	}{
		{
			name: "cargo add unpinned",
			args: []string{"add", "serde"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "crates.io", Name: "serde"}}, Reason: "cargo add"},
		},
		{
			name: "cargo add @ version",
			args: []string{"add", "serde@1.0.150"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "crates.io", Name: "serde", Version: "1.0.150"}}, Reason: "cargo add"},
		},
		{
			name: "cargo add --vers",
			args: []string{"add", "serde", "--vers", "1.0.150"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "crates.io", Name: "serde", Version: "1.0.150"}}, Reason: "cargo add"},
		},
		{
			name: "cargo add --version=",
			args: []string{"add", "serde", "--version=1.0.150"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "crates.io", Name: "serde", Version: "1.0.150"}}, Reason: "cargo add"},
		},
		{
			name: "cargo add with --features (flag value not treated as crate)",
			args: []string{"add", "serde", "--features", "derive"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "crates.io", Name: "serde"}}, Reason: "cargo add"},
		},
		{
			name: "cargo install binary crate",
			args: []string{"install", "ripgrep@14.1.0"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "crates.io", Name: "ripgrep", Version: "14.1.0"}}, Reason: "cargo install"},
		},
		{
			name: "cargo install --path skips local",
			args: []string{"install", "--path", "."},
			want: ParseResult{},
		},
		{
			name: "cargo build → passthrough",
			args: []string{"build", "--release"},
			want: ParseResult{},
		},
		{
			name: "cargo test → passthrough",
			args: []string{"test"},
			want: ParseResult{},
		},
		{
			name: "cargo bare → passthrough",
			args: []string{},
			want: ParseResult{},
		},
	}

	p := Cargo()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := p.Parse(tc.args)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parse(%v):\n  got:  %+v\n  want: %+v", tc.args, got, tc.want)
			}
		})
	}
}
