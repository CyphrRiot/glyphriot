package internal

import (
	"crypto/sha256"
	"encoding/binary"
	"strings"
)

// drbg implements a deterministic CSPRNG using SHA-256 in counter mode:
// buf = SHA256(seed || counter); counter++
// nextUint64() draws 8 bytes from the buffer, refilling as needed.
type drbg struct {
	seed    [32]byte
	counter uint64
	buf     [32]byte
	off     int
}

func newDRBG(seedMaterial []byte) *drbg {
	d := &drbg{
		seed: sha256.Sum256(seedMaterial),
	}
	d.refill()
	return d
}

func (d *drbg) refill() {
	var ctr [8]byte
	binary.BigEndian.PutUint64(ctr[:], d.counter)
	h := sha256.New()
	h.Write(d.seed[:])
	h.Write(ctr[:])
	sum := h.Sum(nil)
	copy(d.buf[:], sum)
	d.off = 0
	d.counter++
}

func (d *drbg) nextUint64() uint64 {
	var out uint64
	for i := 0; i < 8; i++ {
		if d.off >= len(d.buf) {
			d.refill()
		}
		out = (out << 8) | uint64(d.buf[d.off])
		d.off++
	}
	return out
}

// randInt returns a uniform integer in [0, n) using rejection sampling.
func (d *drbg) randInt(n int) int {
	if n <= 0 {
		return 0
	}
	N := uint64(n)
	max := ^uint64(0)
	limit := (max / N) * N // largest multiple of N <= max
	var r uint64
	for {
		r = d.nextUint64()
		if r < limit {
			return int(r % N)
		}
	}
}

// Derive returns a deterministic permutation of [0..n-1] and its inverse,
// seeded from SHA-256(key). If key is empty or whitespace, a fixed default
// seed is used for stable behavior.
func Derive(n int, key string) ([]int, []int) {
	if n <= 0 {
		return []int{}, []int{}
	}

	// Identity when no key is provided (explicitly avoid shuffling)
	if strings.TrimSpace(key) == "" {
		p := make([]int, n)
		for i := 0; i < n; i++ {
			p[i] = i
		}
		return p, Inv(p)
	}

	// Deterministic CSPRNG based on SHA-256(key || counter) in counter mode.
	// Used with unbiased Fisher–Yates to guarantee a uniform permutation.
	drbg := newDRBG([]byte(key))

	p := make([]int, n)
	for i := 0; i < n; i++ {
		p[i] = i
	}
	for i := n - 1; i > 0; i-- {
		j := drbg.randInt(i + 1) // j ∈ [0, i], unbiased
		p[i], p[j] = p[j], p[i]
	}

	return p, Inv(p)
}

// Inv computes the inverse mapping of a permutation p where p[i] is the value
// at position i. The returned slice inv has inv[p[i]] = i for all i.
func Inv(p []int) []int {
	inv := make([]int, len(p))
	for i, v := range p {
		inv[v] = i
	}
	return inv
}
