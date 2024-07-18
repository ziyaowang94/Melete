package utils

import (
	"bytes"
	"emulator/proto/uranus/abci/minibank"
	"fmt"
	"log"

	"google.golang.org/protobuf/proto"
)

func CompareBytesList(a, b [][]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if !bytes.Equal(a[i], b[i]) {
			log.Println("bytes", i)
			x, y := new(minibank.BankData), new(minibank.BankData)
			fmt.Println(a[i])
			fmt.Println(b[i])
			if err := proto.Unmarshal(a[i], x); err != nil {
				fmt.Println(err)
			}
			if err := proto.Unmarshal(b[i], y); err != nil {
				fmt.Println(err)
			}
			fmt.Println(x)
			fmt.Println(y)
			return false
		}
	}
	return true
}

func SplitByteArray(data []byte, size int) [][]byte {
	var result [][]byte
	for i := 0; i < len(data); i += size {
		end := i + size
		if end > len(data) {
			end = len(data)
		}
		result = append(result, data[i:end])
	}
	return result
}
