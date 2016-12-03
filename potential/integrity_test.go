package potential

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Integrity(t *testing.T) {
	t.Run("isOK()", func(t *testing.T) {
		report := newIntegrityReport()
		assert.Equal(t, true, report.isOK())
	})
}
