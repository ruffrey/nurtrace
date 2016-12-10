package potential

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Integrity(t *testing.T) {
	t.Run("isOK() works", func(t *testing.T) {
		report := newIntegrityReport()
		assert.Equal(t, true, report.isOK())
	})
	t.Run("when run on a network with a bad cell returns false with the bad cell", func(t *testing.T) {
	})
}
