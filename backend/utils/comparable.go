package utils

type Comparable interface {
	uint | uint8 | uint16 | uint32 | uint64 |
		int | int8 | int16 | int32 | int64 |
		float32 | float64 | string
}

func Min[C Comparable](c1 C, cs ...C) C {
	if len(cs) == 0 {
		return c1
	}

	min := c1
	for _, c := range cs {
		if c < min {
			min = c
		}
	}

	return min
}

func Max[C Comparable](c1 C, cs ...C) C {
	if len(cs) == 0 {
		return c1
	}

	max := c1
	for _, c := range cs {
		if c > max {
			max = c
		}
	}

	return max
}
