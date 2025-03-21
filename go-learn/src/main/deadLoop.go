package main

import (
	"crypto/rand"
	"math/big"
)

func nrand() int64 {
	max := big.NewInt(int64(1) << 3)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}
func main() {
	// 当i第一次被赋的值小于5，则死循环，因为for循环中定义语句只在第一次循环开始的时候生效
	for i := nrand(); i < 5; {

	}
}
