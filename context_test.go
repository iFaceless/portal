package portal

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDumpDepthContext(t *testing.T) {
	ctx := context.Background()
	depth := DumpDepthFromContext(ctx)
	assert.Equal(t, 0, depth)

	ctx = IncrDumpDepthContext(ctx)
	depth = DumpDepthFromContext(ctx)
	assert.Equal(t, 1, depth)
}
