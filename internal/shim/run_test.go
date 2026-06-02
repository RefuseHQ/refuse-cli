package shim

import (
	"reflect"
	"testing"
)

func TestStripFlag(t *testing.T) {
	cases := []struct {
		name      string
		args      []string
		flag      string
		want      []string
		wantFound bool
	}{
		{
			name:      "flag at end",
			args:      []string{"install", "pyyaml==5.3", "--no-refuse"},
			flag:      "--no-refuse",
			want:      []string{"install", "pyyaml==5.3"},
			wantFound: true,
		},
		{
			name:      "flag in middle",
			args:      []string{"install", "--no-refuse", "pyyaml==5.3"},
			flag:      "--no-refuse",
			want:      []string{"install", "pyyaml==5.3"},
			wantFound: true,
		},
		{
			name:      "flag absent",
			args:      []string{"install", "pyyaml==5.3"},
			flag:      "--no-refuse",
			want:      []string{"install", "pyyaml==5.3"},
			wantFound: false,
		},
		{
			name:      "flag repeated",
			args:      []string{"--no-refuse", "add", "lodash", "--no-refuse"},
			flag:      "--no-refuse",
			want:      []string{"add", "lodash"},
			wantFound: true,
		},
		{
			name:      "empty args",
			args:      []string{},
			flag:      "--no-refuse",
			want:      []string{},
			wantFound: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, found := stripFlag(tc.args, tc.flag)
			if found != tc.wantFound {
				t.Errorf("found = %v, want %v", found, tc.wantFound)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
