package neutralprefix

import "math/bits"

// PyRandom reproduces CPython's random.Random for an integer seed: MT19937 with
// the init_by_array seeding CPython uses for int seeds, plus getrandbits,
// _randbelow_with_getrandbits, and shuffle. build_anchor.py (seed 11) and
// build_slices.py (seed 12345) both draw their ordering from this generator, so
// reproducing it exactly is what makes the Go ports' output byte-identical to the
// committed artifacts.
type PyRandom struct {
	mt  [mtN]uint32
	idx int
}

const (
	mtN       = 624
	mtM       = 397
	matrixA   = 0x9908b0df
	upperMask = 0x80000000
	lowerMask = 0x7fffffff
)

// NewPyRandom returns a generator seeded like Python's random.Random(seed) for a
// non-negative integer seed.
func NewPyRandom(seed uint64) *PyRandom {
	r := &PyRandom{}
	r.seed(seed)
	return r
}

func (r *PyRandom) initGenrand(s uint32) {
	r.mt[0] = s
	for i := 1; i < mtN; i++ {
		r.mt[i] = 1812433253*(r.mt[i-1]^(r.mt[i-1]>>30)) + uint32(i)
	}
	r.idx = mtN
}

func (r *PyRandom) initByArray(key []uint32) {
	r.initGenrand(19650218)
	i, j := 1, 0
	k := mtN
	if len(key) > k {
		k = len(key)
	}
	for ; k > 0; k-- {
		r.mt[i] = (r.mt[i] ^ ((r.mt[i-1] ^ (r.mt[i-1] >> 30)) * 1664525)) + key[j] + uint32(j)
		i++
		j++
		if i >= mtN {
			r.mt[0] = r.mt[mtN-1]
			i = 1
		}
		if j >= len(key) {
			j = 0
		}
	}
	for k = mtN - 1; k > 0; k-- {
		r.mt[i] = (r.mt[i] ^ ((r.mt[i-1] ^ (r.mt[i-1] >> 30)) * 1566083941)) - uint32(i)
		i++
		if i >= mtN {
			r.mt[0] = r.mt[mtN-1]
			i = 1
		}
	}
	r.mt[0] = 0x80000000
}

// seed reproduces CPython's int-seed path: the absolute value is split into
// 32-bit little-endian words and fed to init_by_array (a zero seed uses [0]).
func (r *PyRandom) seed(s uint64) {
	if s == 0 {
		r.initByArray([]uint32{0})
		return
	}
	var key []uint32
	for s > 0 {
		key = append(key, uint32(s&0xffffffff))
		s >>= 32
	}
	r.initByArray(key)
}

func (r *PyRandom) genrandUint32() uint32 {
	if r.idx >= mtN {
		var y uint32
		for kk := 0; kk < mtN-mtM; kk++ {
			y = (r.mt[kk] & upperMask) | (r.mt[kk+1] & lowerMask)
			r.mt[kk] = r.mt[kk+mtM] ^ (y >> 1) ^ ((y & 1) * matrixA)
		}
		for kk := mtN - mtM; kk < mtN-1; kk++ {
			y = (r.mt[kk] & upperMask) | (r.mt[kk+1] & lowerMask)
			r.mt[kk] = r.mt[kk+(mtM-mtN)] ^ (y >> 1) ^ ((y & 1) * matrixA)
		}
		y = (r.mt[mtN-1] & upperMask) | (r.mt[0] & lowerMask)
		r.mt[mtN-1] = r.mt[mtM-1] ^ (y >> 1) ^ ((y & 1) * matrixA)
		r.idx = 0
	}
	y := r.mt[r.idx]
	r.idx++
	y ^= y >> 11
	y ^= (y << 7) & 0x9d2c5680
	y ^= (y << 15) & 0xefc60000
	y ^= y >> 18
	return y
}

// getrandbits returns k random bits, matching CPython's _random_getrandbits:
// one 32-bit word for k<=32 (high bits dropped), or little-endian words for
// larger k.
func (r *PyRandom) getrandbits(k int) uint64 {
	if k <= 32 {
		return uint64(r.genrandUint32() >> uint(32-k))
	}
	var result uint64
	shift := uint(0)
	for k > 0 {
		bitsThis := k
		if bitsThis > 32 {
			bitsThis = 32
		}
		v := r.genrandUint32()
		if bitsThis < 32 {
			v >>= uint(32 - bitsThis)
		}
		result |= uint64(v) << shift
		shift += 32
		k -= 32
	}
	return result
}

// randbelow returns a uniform int in [0, n), matching CPython's
// _randbelow_with_getrandbits (rejection sampling on n.bit_length() bits).
func (r *PyRandom) randbelow(n int) int {
	if n <= 0 {
		return 0
	}
	k := bits.Len(uint(n))
	v := r.getrandbits(k)
	for v >= uint64(n) {
		v = r.getrandbits(k)
	}
	return int(v)
}

// Shuffle performs an in-place Fisher-Yates shuffle over n elements using swap,
// consuming the generator exactly as CPython's random.shuffle does (i from n-1
// down to 1, j = randbelow(i+1)).
func (r *PyRandom) Shuffle(n int, swap func(i, j int)) {
	for i := n - 1; i >= 1; i-- {
		j := r.randbelow(i + 1)
		swap(i, j)
	}
}

// ShuffleInts shuffles x in place with Python semantics.
func (r *PyRandom) ShuffleInts(x []int) {
	r.Shuffle(len(x), func(i, j int) { x[i], x[j] = x[j], x[i] })
}
