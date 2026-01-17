package negotiation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHeader(t *testing.T) {
	// Test basic struct creation and field assignment
	header := newHeader("value", "type", "base", "sub", 0.5, map[string]string{"param": "value"})

	assert.Equal(t, "value", header.Value)
	assert.Equal(t, "type", header.Type)
	assert.Equal(t, "base", header.BasePart)
	assert.Equal(t, "sub", header.SubPart)
	assert.Equal(t, 0.5, header.Quality)
	assert.Equal(t, map[string]string{"param": "value"}, header.Parameters)
	assert.Equal(t, "type; param=value", header.NormalizedValue)
	assert.Equal(t, 0, header.originalIndex)
}
