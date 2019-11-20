package must

import (
	"encoding/binary"
	"fmt"
	"math"
	"testing"
)

func TestFloatBits(t *testing.T) {
	bits := binary.LittleEndian.Uint32([]byte{0x44, 0xED, 0x80, 0x7A})
	v := math.Float32frombits(bits)
	fmt.Println(v)

	bits = binary.BigEndian.Uint32([]byte{0x44, 0xED, 0x80, 0x7A})
	v = math.Float32frombits(bits)
	fmt.Println(v)
}
