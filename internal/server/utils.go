package server

import (
	"io"
	"strconv"
)

func getOrderNumber(body io.ReadCloser) (int, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return -1, err
	}
	defer body.Close()
	return strconv.Atoi(string(data))
}

func isValidLuhnAlgorithm(number int) bool {
	sum := 0
	number = reverseNumber(number)

	for isOdd := 1; number > 0; isOdd ^= 1 {
		digit := number % 10
		if isOdd == 1 {
			digit = digit * 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		number /= 10
	}

	return sum%10 == 0
}

func reverseNumber(number int) int {
	reversed := 0
	for ; number > 0; number /= 10 {
		reversed = reversed*10 + number%10
	}
	return reversed
}
