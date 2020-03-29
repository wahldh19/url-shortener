package randstr

import (
	"strings"
	"testing"
)

// TestNew tests the length and composition of strings returned by New.
func TestNew(t *testing.T) {
	t.Run("length", testLength)
	t.Run("composition", testComposition)
}

// testLength checks that the string outputted by New is:
// 0 length in the case of a negative length input.
// 0 length in the case of a zero length input.
// In all other cases, a length equal to the input length.
func testLength(t *testing.T) {
	var tests = []struct {
		Input    int
		Expected int
	}{
		{Input: -1, Expected: 0},
		{Input: 0, Expected: 0},
		{Input: 1, Expected: 1},
		{Input: 8, Expected: 8},
		{Input: 1024, Expected: 1024},
	}

	for _, test := range tests {
		str := New(test.Input)

		if len(str) != test.Expected {
			t.Error("string length does not equal requested length")
		}
	}
}

// testComposition checks that the string outputted by New:
// Is empty if the requested length is negative.
// Is empty if the requested length is zero.
// In all other cases, is composed solely by characters in the charset.
func testComposition(t *testing.T) {
	if str := New(-1); str != "" {
		t.Error("negative length returned a non-empty string")
	}

	if str := New(0); str != "" {
		t.Error("zero length returned a non-empty string")
	}

	var lengths = []int{1, 8, 1024}
	for _, l := range lengths {
		for _, r := range New(l) {
			if !strings.ContainsRune(charset, r) {
				t.Error("string has rune not in charset")
				break
			}
		}
	}
}
