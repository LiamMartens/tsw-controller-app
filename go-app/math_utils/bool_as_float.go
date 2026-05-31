package math_utils

func BoolAsFloat(v bool) float64 {
	if v {
		return 1.0
	}

	return 0.0
}
