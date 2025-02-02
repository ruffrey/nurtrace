package potential

import (
	"encoding/json"
	"testing"

	"github.com/ruffrey/nurtrace/laws"
	"github.com/stretchr/testify/assert"
)

func Test_FiringPattern(t *testing.T) {
	t.Run("FireNetworkUntilDone fires the seed cells and returns the fired ones", func(t *testing.T) {
		network := NewNetwork()
		a := NewCell(network)
		a.Tag = "a"
		b := NewCell(network)
		b.Tag = "b"
		c := NewCell(network)
		c.Tag = "c"
		d := NewCell(network)
		d.Tag = "d"

		// purposely in a forever firing loop to make sure it exits

		s1 := network.linkCells(a.ID, b.ID)
		s2 := network.linkCells(b.ID, c.ID)
		s3 := network.linkCells(c.ID, a.ID)
		network.linkCells(d.ID, c.ID) // will not fire
		s1.Millivolts = laws.ActualSynapseMax
		s2.Millivolts = laws.ActualSynapseMax
		s3.Millivolts = laws.ActualSynapseMax

		// never fires d
		cells := make(FiringPattern)
		cells[a.ID] = 1
		result := FireNetworkUntilDone(network, cells)
		assert.Equal(t, 3, len(result), "wrong number of cells fired")
		assert.Equal(t, uint16(0), result[d.ID], "should not have fired this cell")
		assert.Equal(t, uint16(2), result[a.ID], "did not fire cell: a-0")
		assert.Equal(t, uint16(3), result[b.ID], "did not fire cell: b-1")
		assert.Equal(t, uint16(2), result[c.ID], "did not fire cell: c-2")
	})
}

func Test_FiringDiffRatio(t *testing.T) {
	t.Run("identical firing patterns have ratio of 1", func(t *testing.T) {
		fp1 := make(FiringPattern)
		fp2 := make(FiringPattern)

		// 1 different
		fp1[CellID(0)] = 2
		fp2[CellID(0)] = 2

		// 2 different
		fp1[CellID(1)] = 7
		fp2[CellID(1)] = 7

		diff := DiffFiringPatterns(fp1, fp2)
		r, _ := diff.SimilarityRatio()
		assert.Equal(t, 1.0, r)
	})
	t.Run("totally different firing patterns have ratio of 0", func(t *testing.T) {
		fp1 := make(FiringPattern)
		fp2 := make(FiringPattern)

		fp1[CellID(0)] = 1
		fp1[CellID(10)] = 2

		fp2[CellID(77)] = 3
		fp2[CellID(99)] = 4

		diff := DiffFiringPatterns(fp1, fp2)
		r, _ := diff.SimilarityRatio()
		assert.Equal(t, 0.0, r)
	})
	t.Run("half the same firing patterns have 0.5 ratio", func(t *testing.T) {
		fp1 := make(FiringPattern)
		fp2 := make(FiringPattern)

		fp1[CellID(0)] = 1
		fp1[CellID(1)] = 1

		fp2[CellID(0)] = 1
		fp2[CellID(2)] = 0

		diff := DiffFiringPatterns(fp1, fp2)
		r, _ := diff.SimilarityRatio()
		assert.Equal(t, 0.5, r)
	})
	t.Run("radio calculates the number of unrepresented fires to represented fires", func(t *testing.T) {
		fp1 := make(FiringPattern)
		fp2 := make(FiringPattern)

		// 1 different
		fp1[CellID(0)] = 2
		fp2[CellID(0)] = 1

		// 2 different
		fp1[CellID(1)] = 14
		fp2[CellID(1)] = 16

		// 4 different / unshared
		fp1[CellID(2)] = 4

		diff := DiffFiringPatterns(fp1, fp2)
		r, _ := diff.SimilarityRatio()
		// (1 + 2 + 4) / (3 + 14 + 16 + 4)
		tot := 2 + 16 + 4.0
		assert.Equal(t, (tot-(1+2+4.0))/tot, r)
	})
}

func Test_FiringPatternMerge(t *testing.T) {
	t.Run("merging firing patterns returns a new combined pattern", func(t *testing.T) {
		fp1 := make(FiringPattern)
		fp2 := make(FiringPattern)

		// 1 different
		fp1[CellID(0)] = 2
		fp2[CellID(0)] = 1

		// 2 different
		fp1[CellID(1)] = 14
		fp2[CellID(1)] = 16

		// 4 different / unshared
		fp1[CellID(2)] = 4

		merged := mergeFiringPatterns(fp1, fp2)

		assert.Equal(t, uint16(1), merged[CellID(0)])
		assert.Equal(t, uint16(15), merged[CellID(1)])
		assert.Equal(t, uint16(4), merged[CellID(2)])
	})
}

func Test_RunFiringPatternTraining(t *testing.T) {
	t.Run("single input and output will predict correctly", func(t *testing.T) {
		// setup the network
		network := NewNetwork()
		network.GrowRandomNeurons(50, laws.ComputedSynapsesPerCell)
		vocab := NewVocabulary(network)

		// setup the training data
		unit := UnitGroup{InputText: "1+3", ExpectedOutput: "4"}
		unitArray := make([]*UnitGroup, 1)
		unitArray[0] = &unit
		unitJSON, err := json.Marshal(unitArray)
		assert.Nil(t, err)

		err = vocab.AddTrainingData(unitJSON)

		assert.Nil(t, err)
		assert.Equal(t, 1, len(vocab.Samples))
		assert.Equal(t, 3, len(vocab.Inputs))
		assert.Equal(t, 1, len(vocab.Outputs))
		assert.Equal(t, 0, len(vocab.Outputs[OutputValue("4")].FirePattern))

		in := make(chan *Vocabulary)
		out := make(chan *Vocabulary)
		go RunFiringPatternTraining(vocab, in, out, "")
		vocab = <-in
		out <- vocab

		assert.NotEqual(t, 0, len(vocab.Outputs[OutputValue("4")].FirePattern))

		network.ResetForTraining()
		assert.Equal(t, "4", Sample("1+3", vocab, 1))
	})
	t.Run("single input out-of-order will predict correctly", func(t *testing.T) {
		// setup the network
		network := NewNetwork()
		network.GrowRandomNeurons(50, laws.ComputedSynapsesPerCell)
		vocab := NewVocabulary(network)

		// setup the training data
		unit := UnitGroup{InputText: "1+3", ExpectedOutput: "4"}
		unitArray := make([]*UnitGroup, 1)
		unitArray[0] = &unit
		unitJSON, err := json.Marshal(unitArray)
		assert.Nil(t, err)

		err = vocab.AddTrainingData(unitJSON)

		assert.Nil(t, err)
		assert.Equal(t, 1, len(vocab.Samples))
		assert.Equal(t, 3, len(vocab.Inputs))
		assert.Equal(t, 1, len(vocab.Outputs))
		assert.Equal(t, 0, len(vocab.Outputs[OutputValue("4")].FirePattern))

		in := make(chan *Vocabulary)
		out := make(chan *Vocabulary)
		go RunFiringPatternTraining(vocab, in, out, "")
		vocab = <-in
		out <- vocab

		assert.NotEqual(t, 0, len(vocab.Outputs[OutputValue("4")].FirePattern))

		network.ResetForTraining()
		assert.Equal(t, "4", Sample("3+1", vocab, 1))
	})
	t.Run("two non-overlapping inputs and outputs will predict correctly", func(t *testing.T) {
		// setup the network
		network := NewNetwork()
		network.GrowRandomNeurons(200, laws.ComputedSynapsesPerCell)
		vocab := NewVocabulary(network)

		// setup the training data
		unit2 := UnitGroup{InputText: "1+3", ExpectedOutput: "4"}
		unit1 := UnitGroup{InputText: "2+5", ExpectedOutput: "7"}

		unitArray := make([]*UnitGroup, 2)
		unitArray[0] = &unit1
		unitArray[1] = &unit2
		unitJSON, err := json.Marshal(unitArray)
		assert.Nil(t, err)

		err = vocab.AddTrainingData(unitJSON)
		assert.Nil(t, err)

		// prechecks
		assert.Equal(t, 2, len(vocab.Samples))
		assert.Equal(t, 5, len(vocab.Inputs))
		assert.Equal(t, 2, len(vocab.Outputs))
		assert.Equal(t, 0, len(vocab.Outputs[OutputValue("4")].FirePattern))
		assert.Equal(t, 0, len(vocab.Outputs[OutputValue("7")].FirePattern))

		in := make(chan *Vocabulary)
		out := make(chan *Vocabulary)
		go RunFiringPatternTraining(vocab, in, out, "")
		vocab = <-in
		out <- vocab

		// tests
		assert.NotEqual(t, 0, len(vocab.Outputs[OutputValue("4")].FirePattern))
		assert.NotEqual(t, 0, len(vocab.Outputs[OutputValue("7")].FirePattern))

		// most important checks
		network.ResetForTraining()
		assert.Equal(t, "4", Sample("1+3", vocab, 1))
		network.ResetForTraining()
		// assert.Equal(t, "7", Sample("2+5", vocab, 1))
	})
	t.Run("two overlapping inputs and non-overlapping outputs will predict correctly", func(t *testing.T) {
		// setup the network
		network := NewNetwork()
		network.GrowRandomNeurons(200, laws.ComputedSynapsesPerCell)
		vocab := NewVocabulary(network)

		// setup the training data
		unit1 := UnitGroup{InputText: "2+3", ExpectedOutput: "5"}
		unit2 := UnitGroup{InputText: "3+4", ExpectedOutput: "7"}

		unitArray := make([]*UnitGroup, 2)
		unitArray[0] = &unit1
		unitArray[1] = &unit2
		unitJSON, err := json.Marshal(unitArray)
		assert.Nil(t, err)

		err = vocab.AddTrainingData(unitJSON)
		assert.Nil(t, err)

		// prechecks
		assert.Equal(t, 2, len(vocab.Samples))
		assert.Equal(t, 4, len(vocab.Inputs))
		assert.Equal(t, 2, len(vocab.Outputs))
		assert.Equal(t, 0, len(vocab.Outputs[OutputValue("5")].FirePattern))
		assert.Equal(t, 0, len(vocab.Outputs[OutputValue("7")].FirePattern))

		in := make(chan *Vocabulary)
		out := make(chan *Vocabulary)
		go RunFiringPatternTraining(vocab, in, out, "")
		vocab = <-in
		out <- vocab

		// tests
		assert.NotEqual(t, 0, len(vocab.Outputs[OutputValue("5")].FirePattern))
		assert.NotEqual(t, 0, len(vocab.Outputs[OutputValue("7")].FirePattern))

		// most important checks
		assert.Equal(t, "5", Sample("2+3", vocab, 1))
		assert.Equal(t, "7", Sample("3+4", vocab, 1))
	})
	t.Run("two overlapping inputs and overlapping outputs will predict correctly", func(t *testing.T) {
		// setup the network
		network := NewNetwork()
		network.GrowRandomNeurons(200, laws.ComputedSynapsesPerCell)
		vocab := NewVocabulary(network)

		// setup the training data
		unit1 := UnitGroup{InputText: "2+3", ExpectedOutput: "5"}
		unit2 := UnitGroup{InputText: "3+5", ExpectedOutput: "8"}

		unitArray := make([]*UnitGroup, 2)
		unitArray[0] = &unit1
		unitArray[1] = &unit2
		unitJSON, err := json.Marshal(unitArray)
		assert.Nil(t, err)

		err = vocab.AddTrainingData(unitJSON)
		assert.Nil(t, err)

		// prechecks
		assert.Equal(t, 2, len(vocab.Samples))
		assert.Equal(t, 4, len(vocab.Inputs))
		assert.Equal(t, 2, len(vocab.Outputs))
		assert.Equal(t, 0, len(vocab.Outputs[OutputValue("5")].FirePattern))
		assert.Equal(t, 0, len(vocab.Outputs[OutputValue("8")].FirePattern))

		in := make(chan *Vocabulary)
		out := make(chan *Vocabulary)
		go RunFiringPatternTraining(vocab, in, out, "")
		vocab = <-in
		out <- vocab

		// tests
		assert.NotEqual(t, 0, len(vocab.Outputs[OutputValue("5")].FirePattern))
		assert.NotEqual(t, 0, len(vocab.Outputs[OutputValue("8")].FirePattern))

		// most important checks
		assert.Equal(t, "5", Sample("2+3", vocab, 1))
		assert.Equal(t, "8", Sample("3+5", vocab, 1))
	})
	t.Run("mixed inputs and overlapping outputs will predict correctly", func(t *testing.T) {
		// setup the network
		network := NewNetwork()
		network.GrowRandomNeurons(400, laws.ComputedSynapsesPerCell)
		vocab := NewVocabulary(network)

		// setup the training data
		unit1 := UnitGroup{InputText: "2+3", ExpectedOutput: "5"}
		unit2 := UnitGroup{InputText: "3+5", ExpectedOutput: "8"}
		unit3 := UnitGroup{InputText: "1+5", ExpectedOutput: "6"}
		unit4 := UnitGroup{InputText: "4+2", ExpectedOutput: "6"}

		unitArray := make([]*UnitGroup, 4)
		unitArray[0] = &unit1
		unitArray[1] = &unit2
		unitArray[2] = &unit3
		unitArray[3] = &unit4
		unitJSON, err := json.Marshal(unitArray)
		assert.Nil(t, err)

		err = vocab.AddTrainingData(unitJSON)
		assert.Nil(t, err)

		// prechecks
		assert.Equal(t, 4, len(vocab.Samples))
		assert.Equal(t, 6, len(vocab.Inputs))
		assert.Equal(t, 3, len(vocab.Outputs))
		assert.Equal(t, 0, len(vocab.Outputs[OutputValue("5")].FirePattern))
		assert.Equal(t, 0, len(vocab.Outputs[OutputValue("8")].FirePattern))
		assert.Equal(t, 0, len(vocab.Outputs[OutputValue("6")].FirePattern))

		in := make(chan *Vocabulary)
		out := make(chan *Vocabulary)
		go RunFiringPatternTraining(vocab, in, out, "")
		vocab = <-in
		out <- vocab

		// tests
		assert.NotEqual(t, 0, len(vocab.Outputs[OutputValue("5")].FirePattern))
		assert.NotEqual(t, 0, len(vocab.Outputs[OutputValue("8")].FirePattern))
		assert.NotEqual(t, 0, len(vocab.Outputs[OutputValue("6")].FirePattern))

		// most important checks
		network.ResetForTraining()
		assert.Equal(t, "5", Sample("2+3", vocab, 1))
		network.ResetForTraining()
		assert.Equal(t, "8", Sample("3+5", vocab, 1))
		network.ResetForTraining()
		assert.Equal(t, "6", Sample("1+5", vocab, 1))
		network.ResetForTraining()
		assert.Equal(t, "6", Sample("4+2", vocab, 1))
	})
}

func Test_FindClosestOutputCollection(t *testing.T) {
	t.Run("finds an exact firing pattern", func(t *testing.T) {
		network := NewNetwork()
		vocab := NewVocabulary(network)

		outputA := make(FiringPattern)
		outputA[CellID(0)] = 2
		outputA[CellID(1)] = 2
		outputA[CellID(2)] = 2
		vocab.Outputs[OutputValue("A")] = NewOutputCollection(OutputValue("A"))
		vocab.Outputs[OutputValue("A")].FirePattern = outputA

		outputB := make(FiringPattern)
		outputB[CellID(0)] = 1
		outputB[CellID(1)] = 1
		vocab.Outputs[OutputValue("B")] = NewOutputCollection(OutputValue("B"))
		vocab.Outputs[OutputValue("B")].FirePattern = outputB

		// what we will try to find
		fp := make(FiringPattern)
		fp[CellID(0)] = 1
		fp[CellID(1)] = 1

		foundOutput := FindClosestOutputCollection(fp, vocab)
		assert.Equal(t, OutputValue("B"), foundOutput.Value)
	})
	t.Run("finds a partial match", func(t *testing.T) {
		network := NewNetwork()
		vocab := NewVocabulary(network)

		outputA := make(FiringPattern)
		outputA[CellID(0)] = 1
		outputA[CellID(1)] = 1
		outputA[CellID(2)] = 1
		vocab.Outputs[OutputValue("A")] = NewOutputCollection(OutputValue("A"))
		vocab.Outputs[OutputValue("A")].FirePattern = outputA

		outputB := make(FiringPattern)
		outputB[CellID(0)] = 3
		outputB[CellID(1)] = 3
		outputB[CellID(2)] = 3
		vocab.Outputs[OutputValue("B")] = NewOutputCollection(OutputValue("B"))
		vocab.Outputs[OutputValue("B")].FirePattern = outputB

		// what we will try to find
		fp := make(FiringPattern)
		fp[CellID(0)] = 1
		fp[CellID(1)] = 1

		foundOutput := FindClosestOutputCollection(fp, vocab)
		assert.Equal(t, OutputValue("A"), foundOutput.Value)
	})
	t.Run("finds a partial match when only slightly different", func(t *testing.T) {
		network := NewNetwork()
		vocab := NewVocabulary(network)

		outputA := make(FiringPattern)
		outputA[CellID(0)] = 1
		outputA[CellID(1)] = 1
		outputA[CellID(2)] = 1
		vocab.Outputs[OutputValue("A")] = NewOutputCollection(OutputValue("A"))
		vocab.Outputs[OutputValue("A")].FirePattern = outputA

		outputB := make(FiringPattern)
		outputB[CellID(0)] = 1
		outputB[CellID(1)] = 1
		outputB[CellID(2)] = 2
		vocab.Outputs[OutputValue("B")] = NewOutputCollection(OutputValue("B"))
		vocab.Outputs[OutputValue("B")].FirePattern = outputB

		// what we will try to find
		fp := make(FiringPattern)
		fp[CellID(0)] = 1
		fp[CellID(1)] = 1
		fp[CellID(2)] = 1

		foundOutput := FindClosestOutputCollection(fp, vocab)
		assert.Equal(t, OutputValue("A"), foundOutput.Value)
	})
	t.Run("finds a partial match with only vague overlaps", func(t *testing.T) {
		network := NewNetwork()
		vocab := NewVocabulary(network)

		outputA := make(FiringPattern)
		outputA[CellID(0)] = 1
		outputA[CellID(1)] = 1
		outputA[CellID(2)] = 1
		vocab.Outputs[OutputValue("A")] = NewOutputCollection(OutputValue("A"))
		vocab.Outputs[OutputValue("A")].FirePattern = outputA

		outputB := make(FiringPattern)
		outputB[CellID(2)] = 1
		outputB[CellID(3)] = 1
		outputB[CellID(4)] = 1
		vocab.Outputs[OutputValue("B")] = NewOutputCollection(OutputValue("B"))
		vocab.Outputs[OutputValue("B")].FirePattern = outputB

		// what we will try to find
		fp := make(FiringPattern)
		fp[CellID(4)] = 1

		foundOutput := FindClosestOutputCollection(fp, vocab)
		assert.Equal(t, OutputValue("B"), foundOutput.Value)
	})

}
