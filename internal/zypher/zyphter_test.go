package zypher

import "testing"

func TestAsciZyphNoIgnoreSpace(t *testing.T) {
	v := "Abc Z19"
	zy := NewZypher(WithShiftIterCount(1))
	reslt, _ := zy.AsciZyph(v)
	want := "DefxC42"

	if reslt != want {
		t.Errorf("got %q, wanted %q", reslt, want)
	}
}

func TestAsciZyphIgnoreSpace(t *testing.T) {
	v := "Abc Z19"
	zy := NewZypher(WithShiftIterCount(1), WithIgnoreSpace(true))
	reslt, _ := zy.AsciZyph(v)
	want := "Def C42"

	if reslt != want {
		t.Errorf("got %q, wanted %q", reslt, want)
	}
}

func TestDefaultZops(t *testing.T) {
	zy := DefaultZops()
	want := wantZyph(3, 3, 3, false, false)

	if zy.Shift != want.shift {
		t.Errorf("got shift %q, wanted shift %q", zy.Shift, want.shift)
	}

	if zy.ShiftIterCount != want.shfIterCount {
		t.Errorf("got shfitercount %q, wanted shfitercount %q", zy.ShiftIterCount, want.shfIterCount)
	}

	if zy.HashIterCount != want.hshIterCount {
		t.Errorf("got hshitercount %v, wanted hshitercount %v", zy.HashIterCount, want.hshIterCount)
	}

	if zy.Alternate != want.alternate {
		t.Errorf("got alternate %v, wanted alternate %v", zy.Alternate, want.alternate)
	}

	if zy.IgnoreSpace != want.ignoreSpace {
		t.Errorf("got ignorespace %v, wanted ignorespace %v", zy.IgnoreSpace, want.ignoreSpace)
	}

}

func TestNewZyph(t *testing.T) {
	zy := NewZypher()
	want := wantZyph(3, 3, 3, false, false)
	if zy.Shift != want.shift {
		t.Errorf("got shift %q, wanted shift %q", zy.Shift, want.shift)
	}

	if zy.ShiftIterCount != want.shfIterCount {
		t.Errorf("got shfitercount %q, wanted shfitercount %q", zy.ShiftIterCount, want.shfIterCount)
	}

	if zy.HashIterCount != want.hshIterCount {
		t.Errorf("got hshitercount %v, wanted hshitercount %v", zy.HashIterCount, want.hshIterCount)
	}

	if zy.Alternate != want.alternate {
		t.Errorf("got alternate %v, wanted alternate %v", zy.Alternate, want.alternate)
	}

	if zy.IgnoreSpace != want.ignoreSpace {
		t.Errorf("got ignorespace %v, wanted ignorespace %v", zy.IgnoreSpace, want.ignoreSpace)
	}
}

func TestCustomZyph(t *testing.T) {
	shf := 4
	shfiter := 4
	dfltHshIter := 3
	alt := true
	ign := true

	zy := NewZypher(WithShift(shf), WithShiftIterCount(shfiter), WithAlternate(alt), WithIgnoreSpace(ign))
	want := wantZyph(shf, shfiter, dfltHshIter, alt, ign)

	if zy.Shift != want.shift {
		t.Errorf("got shift %q, wanted shift %q", zy.Shift, want.shift)
	}

	if zy.ShiftIterCount != want.shfIterCount {
		t.Errorf("got itercount %q, wanted itercount %q", zy.ShiftIterCount, want.shfIterCount)
	}

	if zy.Alternate != want.alternate {
		t.Errorf("got alternate %v, wanted alternate %v", zy.Alternate, want.alternate)
	}

	if zy.IgnoreSpace != want.ignoreSpace {
		t.Errorf("got ignorespace %v, wanted ignorespace %v", zy.IgnoreSpace, want.ignoreSpace)
	}
}

func wantZyph(shf, shfItr, hshItr int, alt, ign bool) struct {
	shift        int
	shfIterCount int
	hshIterCount int
	alternate    bool
	ignoreSpace  bool
} {
	return struct {
		shift        int
		shfIterCount int
		hshIterCount int
		alternate    bool
		ignoreSpace  bool
	}{
		shift:        shf,
		shfIterCount: shfItr,
		hshIterCount: hshItr,
		alternate:    alt,
		ignoreSpace:  ign,
	}
}
