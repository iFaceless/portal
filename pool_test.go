package portal

import (
	"context"
	"testing"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func Test_submitJobsOk(t *testing.T) {
	ctx := context.TODO()

	resultChan, err := submitJobs(ctx, func(payload interface{}) (i interface{}, e error) {
		return payload, nil
	}, 1)
	assert.Nil(t, err)

	for result := range resultChan {
		assert.NotNil(t, result)
		assert.Nil(t, result.Err)
		assert.Equal(t, 1, result.Data)
	}
}

func Test_submitJobsReturnErr(t *testing.T) {
	ctx := context.TODO()

	resultChan, err := submitJobs(ctx, func(payload interface{}) (i interface{}, e error) {
		return nil, errors.New("error happened")
	}, 1)
	assert.Nil(t, err)

	for result := range resultChan {
		assert.NotNil(t, result)
		assert.NotNil(t, result.Err)
		assert.Nil(t, result.Data)
	}
}

func Test_submitJobsCrashed(t *testing.T) {
	ctx := context.TODO()

	resultChan, err := submitJobs(ctx, func(payload interface{}) (i interface{}, e error) {
		panic("job crashed")
		return
	}, 1)
	assert.Nil(t, err)

	for result := range resultChan {
		assert.NotNil(t, result)
		assert.Equal(t, "job crashed", result.Err.Error())
		assert.Nil(t, result.Data)
	}
}
