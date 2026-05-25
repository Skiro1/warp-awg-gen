package awg

import (
	"crypto/rand"
)

type AWGParams struct {
	Jc   int    `yaml:"jc"`
	Jmin int    `yaml:"jmin"`
	Jmax int    `yaml:"jmax"`
	S1   int    `yaml:"s1"`
	S2   int    `yaml:"s2"`
	S3   int    `yaml:"s3"`
	S4   int    `yaml:"s4"`
	H1   string `yaml:"h1"`
	H2   string `yaml:"h2"`
	H3   string `yaml:"h3"`
	H4   string `yaml:"h4"`
}

func GenerateDefaultParams() *AWGParams {
	return &AWGParams{
		Jc:   4,
		Jmin: 40,
		Jmax: 70,
		S1:   0,
		S2:   0,
		S3:   0,
		S4:   0,
		H1:   "1",
		H2:   "2",
		H3:   "3",
		H4:   "4",
	}
}

func cryptoRandInt64(max int64) int64 {
	if max <= 0 {
		return 0
	}
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return 0
	}
	val := int64(b[0]) | int64(b[1])<<8 | int64(b[2])<<16 | int64(b[3])<<24 |
		int64(b[4])<<32 | int64(b[5])<<40 | int64(b[6])<<48 | int64(b[7])<<56
	if val < 0 {
		val = -val
	}
	return val % max
}
