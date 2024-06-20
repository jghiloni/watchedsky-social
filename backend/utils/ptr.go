package utils

func Ptr[T any](v T) *T {
	p := new(T)
	*p = v

	return p
}
