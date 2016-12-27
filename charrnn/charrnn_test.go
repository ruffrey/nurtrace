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
		network := potential.NewNetwork()
		network.Grow(10, 10, 0)

		text := "kfk9\nI.2"
		c := Charrnn{
			Chars:    strings.Split(text, ""),
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
			Chars:    strings.Split(string(bytes), ""),
			Settings: settings,
		}
		network := potential.NewNetwork()
		network.Grow(10, 5, 10)
		c.PrepareData(network)
		potential.Train(settings, network)
	})
}

func Test_SaveLoadVocab(t *testing.T) {
	t.Run("saves vocab and is reloadable from disk", func(t *testing.T) {
		bytes, err := ioutil.ReadFile("text.txt")
		assert.NoError(t, err)
		c := Charrnn{
			Chars:    strings.Split(string(bytes), ""),
			Settings: potential.NewTrainingSettings(),
		}
		network := potential.NewNetwork()
		network.Grow(10, 5, 10)
		c.PrepareData(network)
		assert.NotEqual(t, 0, len(c.Settings.Data.KeyToItem))

		err = c.SaveVocab("test.json")
		assert.NoError(t, err, "failed saving vocab as json")
		bytes, err = ioutil.ReadFile("test.json")
		assert.NoError(t, err, "failed reading back vocab from test.json as bytes")
		assert.NotEqual(t, 0, len(string(bytes)))

		c2 := Charrnn{
			Settings: potential.NewTrainingSettings(),
		}
		err = c2.LoadVocab("test.json")
		assert.NoError(t, err, "failed reading saved vocab")
		assert.Equal(t, len(c.Settings.Data.KeyToItem), len(c2.Settings.Data.KeyToItem))
	})
}
