package utils

import "fmt"

func AnySlice[T any](src []T) []any {
	if src == nil {
		return nil
	}

	dest := make([]any, len(src))
	for i := range src {
		dest[i] = src[i]
	}

	return dest
}

func FromAnySlice[T any](src []any) []T {
	if src == nil {
		return nil
	}

	dest := make([]T, len(src))
	for i := range src {
		d, ok := src[i].(T)
		if !ok {
			panic(fmt.Errorf("%T is not T", d))
		}

		dest[i] = d
	}

	return dest
}

func Reverse[T any](src []T) []T {
	if src == nil {
		return nil
	}

	dst := make([]T, len(src))
	last := len(src) - 1
	for i := range src {
		dst[last-i] = src[i]
	}

	return dst
}

func SubsliceUntil[T any](src []T, filter func(item T) bool) []T {
	dst := make([]T, 0, len(src))
	for _, item := range src {
		if filter(item) {
			return dst
		}

		dst = append(dst, item)
	}

	return dst
}

func Map[R any, T any](src []T, mapper func(T) R) []R {
	rs := make([]R, 0, len(src))
	for _, t := range src {
		rs = append(rs, mapper(t))
	}

	return rs
}

func Filter[T any](src []T, filterer func(T) bool) []T {
	res := make([]T, 0, len(src))
	for _, t := range src {
		if filterer(t) {
			res = append(res, t)
		}
	}
	return res
}

func Reduce[R any, T any](src []T, reducer func(R, T) R, initialValue R) R {
	value := initialValue
	for _, t := range src {
		value = reducer(value, t)
	}

	return value
}
