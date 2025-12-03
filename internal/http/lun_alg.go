package http

//алгоритм луна
func isLuhnValid(number string) bool {
	if len(number) == 0 {
		return false
	}

	sum := 0
	double := false

	for i := len(number) - 1; i >= 0; i-- {
		ch := number[i]
		if ch < '0' || ch > '9' {
			return false
		}
		d := int(ch - '0')

		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}

	return sum%10 == 0
}
