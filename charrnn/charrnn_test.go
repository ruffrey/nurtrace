package charrnn

import (
	"bleh/potential"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SetupVocab(t *testing.T) {
	t.Run("making new vocab adds all unique characters to a map", func(t *testing.T) {
		n := potential.NewNetwork()
		network := &n
		network.Grow(10, 10, 0)

		text := "kfk9\nI.2"
		c := Charrnn{
			chars:    strings.Split(text, ""),
			Settings: potential.NewTrainingSettings(),
		}
		c.PrepareData(network)

		if _, ok := c.Settings.Data.KeyToItem[""]; ok {
			t.Error("should not have added empty string as key")
		}
		if _, ok := c.Settings.Data.KeyToItem["k"]; !ok {
			t.Error("did not add k")
		}
		if _, ok := c.Settings.Data.KeyToItem["f"]; !ok {
			t.Error("did not add f")
		}
		if _, ok := c.Settings.Data.KeyToItem["9"]; !ok {
			t.Error("did not add 9")
		}
		if _, ok := c.Settings.Data.KeyToItem["\n"]; !ok {
			t.Error("did not add \n")
		}
		if _, ok := c.Settings.Data.KeyToItem["I"]; !ok {
			t.Error("did not add I")
		}
		if _, ok := c.Settings.Data.KeyToItem["."]; !ok {
			t.Error("did not add .")
		}
		if _, ok := c.Settings.Data.KeyToItem["2"]; !ok {
			t.Error("did not add 2")
		}
		if _, ok := c.Settings.Data.KeyToItem["START"]; !ok {
			t.Error("did not add START")
		}
		if _, ok := c.Settings.Data.KeyToItem["END"]; !ok {
			t.Error("did not add END")
		}

		assert.Equal(t, 9, len(c.Settings.Data.KeyToItem), "incorrect number of things in vocab")
	})
}

func Test_Training(t *testing.T) {
	t.Run("it works", func(t *testing.T) {
		settings := potential.NewTrainingSettings()
		bytes, err := ioutil.ReadFile("text.txt")
		assert.NoError(t, err)
		c := Charrnn{
			chars:    strings.Split(string(bytes), ""),
			Settings: settings,
		}
		n := potential.NewNetwork()
		network := &n
		network.Grow(10, 5, 10)
		potential.Train(c, settings, network)
	})
}
