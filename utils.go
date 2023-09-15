package jsonobj

import "bytes"

func camelToSnake(camel string) string {
	var snake bytes.Buffer
	var prevLower bool
	var cur, next rune

	for i := range camel {
		var nextLower bool
		cur = rune(camel[i])
		isDigit := ('0' <= cur && cur <= '9')

		// check if next char is lower case
		if i < len(camel)-1 {
			next = rune(camel[i+1])
			if 'a' <= next && next <= 'z' {
				nextLower = true
			}
		}

		// if it is upper case or a digit
		if ('A' <= cur && cur <= 'Z') || isDigit {
			// just convert [A-Z] to _[a-z]
			if snake.Len() > 0 && (nextLower || prevLower) {
				snake.WriteRune('_')
			}

			if isDigit {
				// don't convert digit
				snake.WriteRune(cur)
			} else {
				// convert upper to lower
				snake.WriteRune(cur - 'A' + 'a')
			}

			prevLower = false
		} else {
			snake.WriteRune(cur)
			prevLower = true
		}
	}

	return snake.String()
}
