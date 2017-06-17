package main

import (
	"log"

	"github.com/bjwbell/gensimd/simd"
)

func main() {
	a := simd.I16x8{1, 2, 3, 4, 5, 6, 7, 8}
	b := simd.I16x8{-8, -7, -6, -5, -4, -3, -2, -1}

	result := simd.AddI16x8(a, b)
	log.Println(result)
}
