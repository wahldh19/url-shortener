package randstr

import "testing"

func benchmarkNew(n int, b *testing.B) {
	for i := 0; i < b.N; i++ {
		New(n)
	}
}

func BenchmarkNew2(b *testing.B)  { benchmarkNew(2, b) }
func BenchmarkNew4(b *testing.B)  { benchmarkNew(4, b) }
func BenchmarkNew8(b *testing.B)  { benchmarkNew(8, b) }
func BenchmarkNew16(b *testing.B) { benchmarkNew(16, b) }
func BenchmarkNew32(b *testing.B) { benchmarkNew(32, b) }
