package web

type integer interface {
	~int | ~int64
}

func resolveMax[T integer](requested, defaultValue, maximum T) T {
	if requested <= 0 {
		requested = defaultValue
	}

	return min(requested, maximum)
}

func resolveMinMax[T integer](requested, defaultValue, minimum, maximum T) T {
	if requested <= 0 {
		requested = defaultValue
	}

	return min(max(requested, minimum), maximum)
}
