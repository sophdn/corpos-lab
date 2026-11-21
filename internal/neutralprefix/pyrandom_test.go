package neutralprefix

import (
	"reflect"
	"testing"
)

// Ground truth captured from CPython:
//
//	r=random.Random(11);  [r.getrandbits(5) for _ in range(8)]
//	r=random.Random(11);  lst=list(range(20)); r.shuffle(lst)
//	r=random.Random(12345); lst=list(range(30)); r.shuffle(lst)
//	r=random.Random(11);  [r._randbelow(6) for _ in range(8)]
func TestPyRandomGetrandbits(t *testing.T) {
	r := NewPyRandom(11)
	got := make([]uint64, 8)
	for i := range got {
		got[i] = r.getrandbits(5)
	}
	want := []uint64{14, 27, 17, 27, 29, 24, 14, 14}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("getrandbits(5) = %v, want %v", got, want)
	}
}

func TestPyRandomRandbelow(t *testing.T) {
	r := NewPyRandom(11)
	got := make([]int, 8)
	for i := range got {
		got[i] = r.randbelow(6)
	}
	want := []int{3, 4, 3, 3, 4, 4, 1, 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("randbelow(6) = %v, want %v", got, want)
	}
}

func TestPyRandomShuffleSeed11(t *testing.T) {
	x := make([]int, 20)
	for i := range x {
		x[i] = i
	}
	NewPyRandom(11).ShuffleInts(x)
	want := []int{4, 15, 5, 0, 11, 13, 3, 1, 16, 9, 10, 7, 8, 12, 2, 6, 18, 19, 17, 14}
	if !reflect.DeepEqual(x, want) {
		t.Fatalf("shuffle(20) seed 11 = %v, want %v", x, want)
	}
}

func TestPyRandomShuffleSeed12345(t *testing.T) {
	x := make([]int, 30)
	for i := range x {
		x[i] = i
	}
	NewPyRandom(12345).ShuffleInts(x)
	want := []int{1, 14, 12, 16, 27, 22, 7, 20, 17, 15, 24, 2, 10, 21, 4, 19, 3, 28, 5, 29, 18, 8, 6, 11, 9, 25, 26, 0, 23, 13}
	if !reflect.DeepEqual(x, want) {
		t.Fatalf("shuffle(30) seed 12345 = %v, want %v", x, want)
	}
}

func TestPyRandomGetrandbitsWide(t *testing.T) {
	// k>32 exercises the multi-word path; compared against CPython:
	// r=random.Random(11); [r.getrandbits(40) for _ in range(3)]
	r := NewPyRandom(11)
	got := make([]uint64, 3)
	for i := range got {
		got[i] = r.getrandbits(40)
	}
	want := []uint64{951130727789, 943002041895, 858667946125}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("getrandbits(40) = %v, want %v", got, want)
	}
}

func TestPyRandomSeedZero(t *testing.T) {
	r := NewPyRandom(0)
	got := []uint64{r.getrandbits(32), r.getrandbits(32), r.getrandbits(32)}
	want := []uint64{3626764237, 1654615998, 3255389356}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("seed 0 getrandbits(32) = %v, want %v", got, want)
	}
}

func TestPyRandomRandbelowEdges(t *testing.T) {
	r := NewPyRandom(11)
	if r.randbelow(0) != 0 {
		t.Fatal("randbelow(0) should be 0")
	}
	if r.randbelow(1) != 0 {
		t.Fatal("randbelow(1) should be 0")
	}
}
