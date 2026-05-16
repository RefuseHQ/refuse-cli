package parsers

import (
	"reflect"
	"testing"
)

func TestYarnParser(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want ParseResult
	}{
		{
			name: "yarn bare → lockfile",
			args: []string{},
			want: ParseResult{IsInstall: true, Mode: ModeLockfile, LockfileHint: "yarn.lock", Reason: "yarn (no args) → install lockfile"},
		},
		{
			name: "yarn install",
			args: []string{"install"},
			want: ParseResult{IsInstall: true, Mode: ModeLockfile, LockfileHint: "yarn.lock", Reason: "yarn install → lockfile"},
		},
		{
			name: "yarn add",
			args: []string{"add", "lodash@^4"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "npm", Name: "lodash", Version: "^4"}}, Reason: "yarn add"},
		},
		{
			name: "yarn add -D",
			args: []string{"add", "-D", "typescript"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "npm", Name: "typescript"}}, Reason: "yarn add"},
		},
		{
			name: "yarn run → passthrough",
			args: []string{"run", "build"},
			want: ParseResult{},
		},
	}

	p := Yarn()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := p.Parse(tc.args)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parse(%v):\n  got:  %+v\n  want: %+v", tc.args, got, tc.want)
			}
		})
	}
}
