package internal

import (
	"crypto/sha256"
	"encoding/binary"
	"math/rand"
	"strings"
)

// Derive returns a deterministic permutation of [0..n-1] and its inverse,
// seeded from SHA-256(key). If key is empty or whitespace, a fixed default
// seed is used for stable behavior.
func Derive(n int, key string) ([]int, []int) {
	if n <= 0 {
		return []int{}, []int{}
	}

	p := make([]int, n)
	for i := 0; i < n; i++ {
		p[i] = i
	}

	seed := int64(0x9f4e8bfaa1) // default stable seed when no key is provided
	if strings.TrimSpace(key) != "" {
		h := sha256.Sum256([]byte(key))
		// Mix 4x uint64 chunks via XOR into a single int64 seed
		seed = int64(binary.BigEndian.Uint64(h[0:8])) ^
			int64(binary.BigEndian.Uint64(h[8:16])) ^
			int64(binary.BigEndian.Uint64(h[16:24])) ^
			int64(binary.BigEndian.Uint64(h[24:32]))
	}

	r := rand.New(rand.NewSource(seed))
	r.Shuffle(n, func(i, j int) { p[i], p[j] = p[j], p[i] })

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
