package math_utils

/** Inverts an input power value; so a [0:1] range becomes [1:0] and a [-1:0] range becomes [0:-1] */
func InvertInputValue(value float64) float64 {
	if value < 0.0 {
		return -1.0 - value
	}
	return 1.0 - value
}
