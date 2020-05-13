package intrng

import (
	"fmt"
	"sort"
	"strings"
)

func IntRanges(b []int) (xs [][2]int) {
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

type Bytes map[byte]struct{}

func (x Bytes) Push(b byte) {
	x[b] = struct{}{}
}

func (x Bytes) PopFront() byte {
	b := x.Slice()[0]
	delete(x, b)
	return b
}

func (x Bytes) Front() byte {
	return x.Slice()[0]
}

func (x Bytes) Slice() (xs []byte) {
	for v := range x {
		xs = append(xs, v)
	}
	sort.Slice(xs, func(i, j int) bool {
		return xs[i] < xs[j]
	})
	return
}

func (x Bytes) Ranges() (xs [][2]byte) {
	b := x.Slice()
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
		xs = append(xs, [2]byte{y, x})
	}
	return
}

func (x Bytes) Format() string {
	var rs []string
	for _, x := range x.Ranges() {
		var str string
		if x[0] == x[1] {
			str = fmt.Sprintf("%d", x[0])
		} else {
			str = fmt.Sprintf("%d-%d", x[0], x[1])
		}
		rs = append(rs, str)
	}
	return strings.Join(rs, " ")
}
