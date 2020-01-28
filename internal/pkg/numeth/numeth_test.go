package numeth

import (
	"fmt"
	"testing"
)

func TestInterpolate(t *testing.T) {
	xs := []Coordinate{
		{22, -11},
		{44, -13},
		{66, -7},
	}
	coefficients, _ := InterpolationCoefficients(xs)
	fmt.Println(coefficients)
	for _, v := range xs {
		fmt.Println(v.X, v.Y, PolynomialValue(v.X, coefficients))
	}
	fmt.Println(PolynomialValue(44, coefficients))
}
