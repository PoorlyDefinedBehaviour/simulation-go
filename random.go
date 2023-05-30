package main

type Rand struct{}

func NewRand(seed uint64) Rand {
	return Rand{}
}

// Generates a boolean with probability `p` of it being true.
func (rand *Rand) genBool(p float64) bool {
	return false
}

func (rand *Rand) genBetween(min, max uint64) uint64 {
	return 0
}
