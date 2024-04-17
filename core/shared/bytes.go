package shared

import (
	"math"
	"strconv"

	"golang.org/x/exp/constraints"
)

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func ToKilo[T constraints.Integer](bytes T) float64 {
	return ToFixed(float64(bytes)/1024, 2)
}

func ToMega[T constraints.Integer](bytes T) float64 {
	return ToFixed(float64(bytes)/1024/1024, 2)
}

func IntToSignedString[T constraints.Integer](number T) string {
	if number == 0 {
		return ""
	}

	str := strconv.Itoa(int(number))
	if number >= 0 {
		return "+" + str
	}

	return "-" + str
}
