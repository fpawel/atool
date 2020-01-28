package numeth

import (
	"math"
)

type Coordinate = struct {
	X, Y float64
}

func PolynomialValue(x float64, coefficients []float64) (y float64) {
	for i, c := range coefficients {
		y += math.Pow(x, float64(i)) * c
	}
	return
}

func InterpolationCoefficients(xs []Coordinate) (coefficients []float64, ok bool) {
	matrix := make([][]float64, 0)
	for i := range xs {
		matrix = append(matrix, make([]float64, len(xs)+1))
		s, r := xs[i].X, float64(1)
		matrix[i][len(xs)] = xs[i].Y
		for j := range xs {
			matrix[i][j] = r
			r *= s
		}
	}
	coefficients = make([]float64, len(xs))
	for k := range xs {
		k1, s, j := k+1, matrix[k][k], k
		for i := k1; i < len(xs); i++ {
			r := matrix[i][k]
			if math.Abs(r) > math.Abs(s) {
				s, j = r, i
			}
		}
		if s == 0 {
			return
		}
		if j != k {
			for i := k; i < len(xs)+1; i++ {
				matrix[k][i], matrix[j][i] = matrix[j][i], matrix[k][i]
			}
		}
		for j := k1; j < len(xs)+1; j++ {
			matrix[k][j] /= s
		}
		for i := k1; i < len(xs); i++ {
			r := matrix[i][k]
			for j := k1; j < len(xs)+1; j++ {
				matrix[i][j] = matrix[i][j] - matrix[k][j]*r
			}
		}
	}
	for i := len(xs) - 1; i > -1; i-- {
		s := matrix[i][len(xs)]
		for j := i + 1; j < len(xs); j++ {
			s -= matrix[i][j] * coefficients[j]
		}
		coefficients[i] = s
	}
	ok = true
	return
}
