package config

import "testing"

/*
	{
	  "foo": {
		"ax": "aaaa",
		"az": 3,
		"bx": "example"
	  },
	  "fix": {
		"right": "right",
		"error": "error"
	  },
	  "pri": {
		"cz": "b"
	  }
}
*/

func TestMerge(t *testing.T) {
	dest := map[string]any{
		"foo": map[string]any{
			"ax": "2",
			"bx": "example",
		},
		"pri": map[string]any{
			"cz": "b",
		},
	}
	src := map[string]any{
		"foo": map[string]any{
			"az": 3,
			"ax": "aaa",
		},
		"fix": map[string]any{
			"right": "right",
			"error": "error",
		},
	}
	merge(dest, src)
	t.Logf("mp:%+v", dest)
}

func TestSpread(t *testing.T) {
	src := map[string]any{
		"foo": map[string]any{
			"bar": map[int]bool{
				1: true,
			},
			"baz": "123",
		},
	}
	mp := spread(src, "", ".")
	t.Logf("spread mp:%+v", mp)
}
