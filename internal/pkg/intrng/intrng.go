package intrng

func Ranges(b []int) (xs [][2]int) {
	for i := 0; i < len(b); i++ {
		if i < len(b)-1 && (b[i] == b[i+1] || b[i]+1 == b[i+1]) {
			continue
		}

		x, y := b[i], b[i]
		for j := i; j > -1; j-- {
			y = b[j]
			if j > 0 && !(b[j-1] == b[j] || b[j-1]+1 == b[j]) {
				break
			}
		}
		xs = append(xs, [2]int{y, x})
	}
	return
}
