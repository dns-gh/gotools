package algorithms

import (
	"testing"
	"time"
)

const (
	alphabetical = "abcdefghijklmnopqrstuvwxyz"
)

func launchTestBruteForce(toFind string, cpu int) error {
	length := len(toFind)
	return BruteForce(length, cpu, alphabetical, func(candidate string) bool {
		time.Sleep(20 * time.Millisecond)
		return candidate == toFind
	})
}

func TestBruteForce(t *testing.T) {
	err := launchTestBruteForce("test", 1)
	if err != nil {
		t.Error("Test failed:", err.Error())
	}
}

func benchBruteForce(b *testing.B, cpu int) {
	for n := 0; n < b.N; n++ {
		launchTestBruteForce("zz", cpu)
	}
}

func BenchmarkBruteForceCPU1(b *testing.B) { benchBruteForce(b, 1) }
func BenchmarkBruteForceCPU2(b *testing.B) { benchBruteForce(b, 2) }
func BenchmarkBruteForceCPU4(b *testing.B) { benchBruteForce(b, 4) }
func BenchmarkBruteForceCPU8(b *testing.B) { benchBruteForce(b, 8) }
