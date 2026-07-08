package lib

import "testing"

func TestShellQuote(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain", in: "hello", want: "'hello'"},
		{name: "space", in: "hello world", want: "'hello world'"},
		{name: "single quote", in: "can't", want: "'can'\\''t'"},
		{name: "mixed", in: "a b'c$d", want: "'a b'\\''c$d'"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := ShellQuote(tc.in); got != tc.want {
				t.Fatalf("ShellQuote(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
