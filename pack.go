package zero

import (
	"encoding/binary"
	"fmt"
)

// PackInt64 will write
func PackInt64(val int64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, val)
	fmt.Println("encoded", buf[:n])
	return buf[:n]
}

// UnPackInt64 will return
func UnPackInt64(data []byte) (res int64) {
	fmt.Println("decoding", data)
	res, _ = binary.Varint(data)
	return
}
