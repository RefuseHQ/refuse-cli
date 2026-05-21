package parsers

import (
	"reflect"
	"testing"
)

func TestGemParser(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want ParseResult
	}{
		{
			name: "gem install unpinned",
			args: []string{"install", "rails"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "RubyGems", Name: "rails"}}, Reason: "gem install"},
		},
		{
			name: "gem install -v",
			args: []string{"install", "rails", "-v", "6.1.0"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "RubyGems", Name: "rails", Version: "6.1.0"}}, Reason: "gem install"},
		},
		{
			name: "gem install --version=",
			args: []string{"install", "rails", "--version=6.1.0"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "RubyGems", Name: "rails", Version: "6.1.0"}}, Reason: "gem install"},
		},
		{
			name: "gem install colon form",
			args: []string{"install", "rails:6.1.0"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "RubyGems", Name: "rails", Version: "6.1.0"}}, Reason: "gem install"},
		},
		{
			name: "gem i alias",
			args: []string{"i", "rack"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "RubyGems", Name: "rack"}}, Reason: "gem install"},
		},
		{
			name: "gem install operator version trimmed",
			args: []string{"install", "rails", "-v", ">=6.0"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "RubyGems", Name: "rails", Version: "6.0"}}, Reason: "gem install"},
		},
		{
			name: "gem install local .gem skipped",
			args: []string{"install", "./mygem-1.0.gem"},
			want: ParseResult{},
		},
		{
			name: "gem update → passthrough",
			args: []string{"update"},
			want: ParseResult{},
		},
		{
			name: "gem list → passthrough",
			args: []string{"list"},
			want: ParseResult{},
		},
	}

	p := Gem()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := p.Parse(tc.args)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parse(%v):\n  got:  %+v\n  want: %+v", tc.args, got, tc.want)
			}
		})
	}
}
