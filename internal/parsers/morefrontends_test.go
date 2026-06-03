package parsers

import (
	"reflect"
	"testing"
)

type frontendCase struct {
	name string
	args []string
	want ParseResult
}

func runCases(t *testing.T, p Parser, cases []frontendCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := p.Parse(tc.args)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parse(%v):\n  got:  %+v\n  want: %+v", tc.args, got, tc.want)
			}
		})
	}
}

func TestUVParser(t *testing.T) {
	runCases(t, UV(), []frontendCase{
		{"uv pip install",
			[]string{"pip", "install", "requests==2.32.3"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "PyPI", Name: "requests", Version: "2.32.3"}}, Reason: "uv pip install"}},
		{"uv pip install -r",
			[]string{"pip", "install", "-r", "requirements.txt"},
			ParseResult{IsInstall: true, Mode: ModeLockfile, LockfileHint: "requirements.txt", Reason: "uv pip install -r requirements.txt"}},
		{"uv add",
			[]string{"add", "django"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "PyPI", Name: "django"}}, Reason: "uv add"}},
		{"uv add pinned",
			[]string{"add", "django==5.0"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "PyPI", Name: "django", Version: "5.0"}}, Reason: "uv add"}},
		{"uv tool install",
			[]string{"tool", "install", "ruff"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "PyPI", Name: "ruff"}}, Reason: "uv tool install"}},
		{"uv run → passthrough",
			[]string{"run", "python", "-V"},
			ParseResult{}},
		{"uv sync → passthrough",
			[]string{"sync"},
			ParseResult{}},
	})
}

func TestPoetryParser(t *testing.T) {
	runCases(t, Poetry(), []frontendCase{
		{"poetry add",
			[]string{"add", "requests"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "PyPI", Name: "requests"}}, Reason: "poetry add"}},
		{"poetry add @ form",
			[]string{"add", "requests@2.32.3"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "PyPI", Name: "requests", Version: "2.32.3"}}, Reason: "poetry add"}},
		{"poetry add == form",
			[]string{"add", "requests==2.32.3"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "PyPI", Name: "requests", Version: "2.32.3"}}, Reason: "poetry add"}},
		{"poetry install → lockfile",
			[]string{"install"},
			ParseResult{IsInstall: true, Mode: ModeLockfile, LockfileHint: "poetry.lock", Reason: "poetry install → lockfile"}},
		{"poetry run → passthrough",
			[]string{"run", "pytest"},
			ParseResult{}},
	})
}

func TestPipenvParser(t *testing.T) {
	runCases(t, Pipenv(), []frontendCase{
		{"pipenv install pkg",
			[]string{"install", "requests==2.32.3"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "PyPI", Name: "requests", Version: "2.32.3"}}, Reason: "pipenv install"}},
		{"pipenv install (no args) → lockfile",
			[]string{"install"},
			ParseResult{IsInstall: true, Mode: ModeLockfile, LockfileHint: "Pipfile.lock", Reason: "pipenv install → lockfile"}},
		{"pipenv update → passthrough",
			[]string{"update"},
			ParseResult{}},
	})
}

func TestPDMParser(t *testing.T) {
	runCases(t, PDM(), []frontendCase{
		{"pdm add",
			[]string{"add", "requests"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "PyPI", Name: "requests"}}, Reason: "pdm add"}},
		{"pdm add @ form",
			[]string{"add", "requests@2.32.3"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "PyPI", Name: "requests", Version: "2.32.3"}}, Reason: "pdm add"}},
		{"pdm install → lockfile",
			[]string{"install"},
			ParseResult{IsInstall: true, Mode: ModeLockfile, LockfileHint: "pdm.lock", Reason: "pdm install → lockfile"}},
		{"pdm sync → lockfile",
			[]string{"sync"},
			ParseResult{IsInstall: true, Mode: ModeLockfile, LockfileHint: "pdm.lock", Reason: "pdm sync → lockfile"}},
		{"pdm run → passthrough",
			[]string{"run", "python"},
			ParseResult{}},
	})
}

func TestPipxParser(t *testing.T) {
	runCases(t, Pipx(), []frontendCase{
		{"pipx install",
			[]string{"install", "black==24.0.0"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "PyPI", Name: "black", Version: "24.0.0"}}, Reason: "pipx install"}},
		{"pipx run",
			[]string{"run", "cowsay==6.0"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "PyPI", Name: "cowsay", Version: "6.0"}}, Reason: "pipx run"}},
		{"pipx list → passthrough",
			[]string{"list"},
			ParseResult{}},
	})
}

func TestBundleParser(t *testing.T) {
	runCases(t, Bundle(), []frontendCase{
		{"bundle install",
			[]string{"install"},
			ParseResult{IsInstall: true, Mode: ModeLockfile, LockfileHint: "Gemfile.lock", Reason: "bundle install → lockfile"}},
		{"bundle bare → lockfile",
			[]string{},
			ParseResult{IsInstall: true, Mode: ModeLockfile, LockfileHint: "Gemfile.lock", Reason: "bundle (no args) → install Gemfile.lock"}},
		{"bundle add",
			[]string{"add", "rails"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "RubyGems", Name: "rails"}}, Reason: "bundle add"}},
		{"bundle add -v",
			[]string{"add", "rails", "-v", "7.1.0"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "RubyGems", Name: "rails", Version: "7.1.0"}}, Reason: "bundle add"}},
		{"bundle exec → passthrough",
			[]string{"exec", "rake"},
			ParseResult{}},
	})
}

func TestNPXParser(t *testing.T) {
	runCases(t, NPX(), []frontendCase{
		{"npx pkg",
			[]string{"cowsay"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "npm", Name: "cowsay"}}, Reason: "npx"}},
		{"npx pkg@ver runs only first positional",
			[]string{"cowsay@1.0", "hello"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "npm", Name: "cowsay", Version: "1.0"}}, Reason: "npx"}},
		{"npx -p pkg",
			[]string{"-p", "create-react-app@5.0.1", "create-react-app", "my-app"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "npm", Name: "create-react-app", Version: "5.0.1"}}, Reason: "npx"}},
		{"npx local path → passthrough",
			[]string{"./scripts/foo.js"},
			ParseResult{}},
	})
}

func TestComposerParser(t *testing.T) {
	runCases(t, Composer(), []frontendCase{
		{"composer require colon form",
			[]string{"require", "monolog/monolog:2.9.0"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "Packagist", Name: "monolog/monolog", Version: "2.9.0"}}, Reason: "composer require"}},
		{"composer require separate version arg",
			[]string{"require", "monolog/monolog", "^2.9"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "Packagist", Name: "monolog/monolog", Version: "2.9"}}, Reason: "composer require"}},
		{"composer install → lockfile",
			[]string{"install"},
			ParseResult{IsInstall: true, Mode: ModeLockfile, LockfileHint: "composer.lock", Reason: "composer install → lockfile"}},
		{"composer run-script → passthrough",
			[]string{"run-script", "test"},
			ParseResult{}},
	})
}

func TestDotnetParser(t *testing.T) {
	runCases(t, Dotnet(), []frontendCase{
		{"dotnet add package",
			[]string{"add", "package", "Newtonsoft.Json"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "NuGet", Name: "Newtonsoft.Json"}}, Reason: "dotnet add package"}},
		{"dotnet add package -v",
			[]string{"add", "package", "Newtonsoft.Json", "-v", "13.0.3"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "NuGet", Name: "Newtonsoft.Json", Version: "13.0.3"}}, Reason: "dotnet add package"}},
		{"dotnet add package --version=",
			[]string{"add", "package", "Newtonsoft.Json", "--version=13.0.3"},
			ParseResult{IsInstall: true, Mode: ModeDirect, Packages: []PkgRef{{Ecosystem: "NuGet", Name: "Newtonsoft.Json", Version: "13.0.3"}}, Reason: "dotnet add package"}},
		{"dotnet restore → lockfile",
			[]string{"restore"},
			ParseResult{IsInstall: true, Mode: ModeLockfile, LockfileHint: "packages.lock.json", Reason: "dotnet restore → lockfile"}},
		{"dotnet build → passthrough",
			[]string{"build", "--configuration", "Release"},
			ParseResult{}},
	})
}
