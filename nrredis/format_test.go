package nrredis

import "testing"

func TestFormatCommand(t *testing.T) {
	cases := []struct {
		cmd  string
		args []interface{}
		out  string
	}{
		{cmd: "FLUSHALL", out: "flushall"},
		{cmd: "flushall", out: "flushall"},
		{
			cmd:  "ZADD",
			args: []interface{}{"test:set", 100.555, "a", 10, "B"},
			out:  `zadd "test:set" 100.555 "a" 10 "B"`,
		},
	}

	for _, c := range cases {
		result := formatCommand(c.cmd, c.args)
		if got, want := result, c.out; got != want {
			t.Errorf("formatCommand(%s, %v) returned %q, %q", c.cmd, c.args, got, want)
		}
	}
}
