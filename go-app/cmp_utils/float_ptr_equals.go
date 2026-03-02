package cmp_utils

func IsSameFloatValue(a *float64, b *float64) bool {
	if a == nil || b == nil {
		/* if either is nil; they can only be the same if they are equal (meaning they are both nil) */
		return a == b
	}

	return *a == *b
}
