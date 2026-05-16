package parsers

import (
	"reflect"
	"testing"
)

func TestPipParser(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want ParseResult
	}{
		{
			name: "install simple",
			args: []string{"install", "requests"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "PyPI", Name: "requests"}}, Reason: "pip install"},
		},
		{
			name: "install pinned ==",
			args: []string{"install", "requests==2.19.0"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "PyPI", Name: "requests", Version: "2.19.0"}}, Reason: "pip install"},
		},
		{
			name: "install range ~=",
			args: []string{"install", "requests~=2.19"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "PyPI", Name: "requests", Version: "2.19"}}, Reason: "pip install"},
		},
		{
			name: "install -r requirements",
			args: []string{"install", "-r", "requirements.txt"},
			want: ParseResult{IsInstall: true, Mode: ModeLockfile, LockfileHint: "requirements.txt", Reason: "pip install -r requirements.txt"},
		},
		{
			name: "install --requirement= form",
			args: []string{"install", "--requirement=dev-requirements.txt"},
			want: ParseResult{IsInstall: true, Mode: ModeLockfile, LockfileHint: "dev-requirements.txt", Reason: "pip install -r dev-requirements.txt"},
		},
		{
			name: "install with -U flag, packages keep getting parsed",
			args: []string{"install", "-U", "requests"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "PyPI", Name: "requests"}}, Reason: "pip install"},
		},
		{
			name: "install --target consumes next token",
			args: []string{"install", "--target", "/tmp/x", "requests==2.19.0"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "PyPI", Name: "requests", Version: "2.19.0"}}, Reason: "pip install"},
		},
		{
			name: "install local wheel — skip",
			args: []string{"install", "./dist/foo-1.0.whl"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{}, Reason: "pip install"},
		},
		{
			name: "install git+ — skip",
			args: []string{"install", "git+https://github.com/u/r.git@main"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{}, Reason: "pip install"},
		},
		{
			name: "freeze → passthrough",
			args: []string{"freeze"},
			want: ParseResult{},
		},
		{
			name: "uninstall → passthrough",
			args: []string{"uninstall", "requests"},
			want: ParseResult{},
		},
		{
			name: "multiple packages",
			args: []string{"install", "requests==2.19.0", "flask", "click<9"},
			want: ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{
				{Ecosystem: "PyPI", Name: "requests", Version: "2.19.0"},
				{Ecosystem: "PyPI", Name: "flask"},
				{Ecosystem: "PyPI", Name: "click", Version: "9"},
			}, Reason: "pip install"},
		},
	}

	p := Pip()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := p.Parse(tc.args)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parse(%v):\n  got:  %+v\n  want: %+v", tc.args, got, tc.want)
			}
		})
	}
}
