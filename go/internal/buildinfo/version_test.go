package buildinfo

import "testing"

func TestIsVersionCommand(t *testing.T) {
	testCases := []struct {
		name string
		args []string
		want bool
	}{
		{name: "no argument", args: nil, want: false},
		{name: "word", args: []string{"version"}, want: true},
		{name: "long option", args: []string{"--version"}, want: true},
		{name: "short option", args: []string{"-v"}, want: true},
		{name: "case and space", args: []string{" Version "}, want: true},
		{name: "unknown", args: []string{"serve"}, want: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := IsVersionCommand(testCase.args); got != testCase.want {
				t.Fatalf("IsVersionCommand() = %t, want %t", got, testCase.want)
			}
		})
	}
}
