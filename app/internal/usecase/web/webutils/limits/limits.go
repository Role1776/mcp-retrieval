package limits

type integer interface {
	~int | ~int64
}

func ResolveMax[T integer](requested, defaultValue, maximum T) T {
	if requested <= 0 {
		requested = defaultValue
	}

	return min(requested, maximum)
}

func ResolveMinMax[T integer](requested, defaultValue, minimum, maximum T) T {
	if requested <= 0 {
		requested = defaultValue
	}

	return min(max(requested, minimum), maximum)
}
