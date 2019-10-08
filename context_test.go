package portal

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDumpDepthContext(t *testing.T) {
	ctx := context.Background()
	depth := dumpDepthFromContext(ctx)
	assert.Equal(t, 0, depth)

	ctx = incrDumpDepthContext(ctx)
	depth = dumpDepthFromContext(ctx)
	assert.Equal(t, 1, depth)
}
