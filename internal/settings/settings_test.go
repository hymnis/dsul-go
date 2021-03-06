// DSUL - Disturb State USB Light : Settings module tests.
package settings

import "testing"

func TestSomething(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Hello, world", "Hello, world"},
		{"", ""},
	}
	for _, c := range cases {
		got := c.in
		if got != c.want {
			t.Errorf("Wrong(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
