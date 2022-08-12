package cache

import (
	"crypto/rand"
	"math"
	"math/big"
	insecurerand "math/rand"
)

type DJB33Hasher struct {
	Seed uint32
}

func (hasher *DJB33Hasher) Sum32(k string) uint32 {
	var (
		l = uint32(len(k))
		d = 5381 + hasher.Seed + l
		i = uint32(0)
	)
	// Why is all this 5x faster than a for loop?
	if l >= 4 {
		for i < l-4 {
			d = (d * 33) ^ uint32(k[i])
			d = (d * 33) ^ uint32(k[i+1])
			d = (d * 33) ^ uint32(k[i+2])
			d = (d * 33) ^ uint32(k[i+3])
			i += 4
		}
	}
	switch l - i {
	case 1:
	case 2:
		d = (d * 33) ^ uint32(k[i])
	case 3:
		d = (d * 33) ^ uint32(k[i])
		d = (d * 33) ^ uint32(k[i+1])
	case 4:
		d = (d * 33) ^ uint32(k[i])
		d = (d * 33) ^ uint32(k[i+1])
		d = (d * 33) ^ uint32(k[i+2])
	}
	return d ^ (d >> 16)
}

func Seed() uint32 {
	rnd, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return insecurerand.Uint32()
	}
	return uint32(rnd.Uint64())
}
