package bit

func Band(a, b uint64) uint64 { return a & b }
func Bor(a, b uint64) uint64  { return a | b }
func Bxor(a, b uint64) uint64 { return a ^ b }
func Bnot(a uint64) uint64    { return ^a }

func Lshift(a uint64, n int) uint64 {
	if n >= 64 || n <= -64 {
		return 0
	}
	if n >= 0 {
		return a << n
	}
	return a >> -n
}

func Rshift(a uint64, n int) uint64 {
	if n >= 64 || n <= -64 {
		return 0
	}
	if n >= 0 {
		return a >> n
	}
	return a << -n
}

func Arshift(a int64, n int) int64 {
	if n >= 64 {
		if a >= 0 {
			return 0
		}
		return -1
	}
	if n <= -64 {
		return 0
	}
	if n >= 0 {
		return a >> n
	}
	return a << -n
}
