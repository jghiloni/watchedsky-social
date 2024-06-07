package utils

type WholeNumber interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64
}

func WNMin[T WholeNumber](args ...T) T {
	if len(args) == 0 {
		panic("len(args) must be > 0")
	}

	min := args[0]
	args = args[1:]
	for _, a := range args {
		if a < min {
			min = a
		}
	}

	return min
}

func WNMax[T WholeNumber](args ...T) T {
	if len(args) == 0 {
		panic("len(args) must be > 0")
	}

	max := args[0]
	args = args[1:]
	for _, a := range args {
		if a > max {
			max = a
		}
	}

	return max
}
