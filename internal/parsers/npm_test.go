package parsers

import (
	"reflect"
	"testing"
)

func TestNPMParser(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want ParseResult
	}{
		{
			name: "install single",
			args: []string{"install", "lodash"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "npm", Name: "lodash"}}, Reason: "npm direct"},
		},
		{
			name: "install pinned",
			args: []string{"install", "lodash@4.17.10"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "npm", Name: "lodash", Version: "4.17.10"}}, Reason: "npm direct"},
		},
		{
			name: "install scoped",
			args: []string{"install", "@scope/name@^1.0.0"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "npm", Name: "@scope/name", Version: "^1.0.0"}}, Reason: "npm direct"},
		},
		{
			name: "i alias",
			args: []string{"i", "lodash"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "npm", Name: "lodash"}}, Reason: "npm direct"},
		},
		{
			name: "install with save-dev flag",
			args: []string{"install", "--save-dev", "lodash"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "npm", Name: "lodash"}}, Reason: "npm direct"},
		},
		{
			name: "install with prefix and value",
			args: []string{"install", "--prefix", "/opt/x", "lodash"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "npm", Name: "lodash"}}, Reason: "npm direct"},
		},
		{
			name: "install with prefix= eqform",
			args: []string{"install", "--prefix=/opt/x", "lodash"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "npm", Name: "lodash"}}, Reason: "npm direct"},
		},
		{
			name: "install no args → lockfile",
			args: []string{"install"},
			want: ParseResult{IsInstall: true, Mode: ModeLockfile, LockfileHint: "package-lock.json", Reason: "npm install (no args) → lockfile mode"},
		},
		{
			name: "ci → lockfile",
			args: []string{"ci"},
			want: ParseResult{IsInstall: true, Mode: ModeLockfile, LockfileHint: "package-lock.json", Reason: "npm ci → lockfile mode"},
		},
		{
			name: "test verb → passthrough",
			args: []string{"test"},
			want: ParseResult{},
		},
		{
			name: "run verb → passthrough",
			args: []string{"run", "lint"},
			want: ParseResult{},
		},
		{
			name: "multiple packages with mixed flags",
			args: []string{"install", "-D", "lodash", "@scope/x@~1.0.0", "-w", "core", "axios"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{
				{Ecosystem: "npm", Name: "lodash"},
				{Ecosystem: "npm", Name: "@scope/x", Version: "~1.0.0"},
				{Ecosystem: "npm", Name: "axios"},
			}, Reason: "npm direct"},
		},
	}

	p := NPM()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := p.Parse(tc.args)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parse(%v):\n  got:  %+v\n  want: %+v", tc.args, got, tc.want)
			}
		})
	}
}

func TestPNPMUsesPnpmLock(t *testing.T) {
	got := PNPM().Parse([]string{"install"})
	if got.LockfileHint != "pnpm-lock.yaml" {
		t.Errorf("expected pnpm-lock.yaml, got %q", got.LockfileHint)
	}
}

func TestBunUsesBunLock(t *testing.T) {
	got := Bun().Parse([]string{"install"})
	if got.LockfileHint != "bun.lock" {
		t.Errorf("expected bun.lock, got %q", got.LockfileHint)
	}
}
