package randstr

import "testing"

func TestRandomHex(t *testing.T) {
	s := RandomHex(8)
	if len(s) != 16 { // 8 bytes = 16 hex chars
		t.Errorf("expected length 16, got %d", len(s))
	}

	s2 := RandomHex(8)
	if s == s2 {
		t.Error("two consecutive calls should not return the same value")
	}
}

func TestRandomHex_DifferentLengths(t *testing.T) {
	for _, n := range []int{1, 4, 16, 32} {
		got := RandomHex(n)
		if len(got) != n*2 {
			t.Errorf("RandomHex(%d): expected length %d, got %d", n, n*2, len(got))
		}
	}
}
