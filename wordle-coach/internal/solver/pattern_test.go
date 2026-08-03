package solver

import "testing"

// pat builds a Pattern from the g/y/b spelling used in the tests.
func pat(t *testing.T, s string) Pattern {
	t.Helper()
	if len(s) != WordLen {
		t.Fatalf("pattern %q is %d long, want %d", s, len(s), WordLen)
	}
	code := Pattern(0)
	for i := WordLen - 1; i >= 0; i-- {
		var d Pattern
		switch s[i] {
		case 'b':
			d = Grey
		case 'y':
			d = Yellow
		case 'g':
			d = Green
		default:
			t.Fatalf("pattern %q has bad letter %q", s, s[i])
		}
		code = code*3 + d
	}
	return code
}

func TestGetPattern(t *testing.T) {
	tests := []struct {
		name   string
		guess  string
		answer string
		want   string
	}{
		{"exact match", "crane", "crane", "ggggg"},
		{"nothing shared", "crimp", "toads", "bbbbb"},

		// The duplicate-letter case. GEESE has three Es but THOSE has one, and
		// that one is already green, so both other Es must go grey.
		{"duplicates in guess, one in answer", "geese", "those", "bbbgg"},
		{"same the other way round", "those", "geese", "bbbgg"},

		// A letter that is green in one place and yellow in another. The first
		// L is yellow because ALLEY has a second L; the trailing A is grey
		// because ALLEY has only the one A, already spent on the earlier A.
		{"green and yellow of one letter", "llama", "alley", "ygybb"},

		// Two Es in the guess, two in the answer, neither aligned.
		{"duplicates on both sides", "speed", "erase", "ybyyb"},

		// Yellows are assigned left to right, so the leading S takes the
		// answer's only S and the later two get nothing.
		{"extra copies go grey", "sassy", "birds", "ybbbb"},

		// A green consumes its letter before yellows are handed out, so the
		// other two Ss find none left.
		{"green consumes the copy", "sassy", "basil", "bggbb"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			want := pat(t, tc.want)
			if got := GetPattern(tc.guess, tc.answer); got != want {
				t.Errorf("GetPattern(%q, %q) = %d, want %d (%s)", tc.guess, tc.answer, got, want, tc.want)
			}
		})
	}
}

func TestGetPatternAllGreenConstant(t *testing.T) {
	if got := GetPattern("crane", "crane"); got != AllGreen {
		t.Errorf("GetPattern on a match = %d, want AllGreen (%d)", got, AllGreen)
	}
}

func TestPatternDigit(t *testing.T) {
	p := pat(t, "bygbg")
	want := []int{Grey, Yellow, Green, Grey, Green}
	for i, w := range want {
		if got := p.Digit(i); got != w {
			t.Errorf("Digit(%d) = %d, want %d", i, got, w)
		}
	}
}
