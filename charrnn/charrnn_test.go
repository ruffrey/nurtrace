package charrnn

import (
	"bleh/potential"
	"testing"
)

func Test_NewVocab(t *testing.T) {
	t.Run("making new vocab adds all unique characters to a map", func(t *testing.T) {
		text := "kfk9\nI.2"
		network := potential.NewNetwork()
		network.Grow(10, 10, 0)

		vocab := NewVocab(text, &network)

		if _, ok := vocab["k"]; !ok {
			panic("vocab did not add k")
		}
		if _, ok := vocab["f"]; !ok {
			panic("vocab did not add f")
		}
		if _, ok := vocab["9"]; !ok {
			panic("vocab did not add 9")
		}
		if _, ok := vocab["\n"]; !ok {
			panic("vocab did not add \n")
		}
		if _, ok := vocab["I"]; !ok {
			panic("vocab did not add I")
		}
		if _, ok := vocab["."]; !ok {
			panic("vocab did not add .")
		}
		if _, ok := vocab["2"]; !ok {
			panic("vocab did not add 2")
		}
		if len(vocab) != 7 {
			panic("vocab has too many things")
		}
	})
}
