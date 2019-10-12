package portal

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
)

var (
	// maxWorkerPoolSize is default to 10k.
	// Since the number of incoming requests are unknown,
	// we must limit the spawned goroutines to avoid
	// consuming too many resources.
	maxWorkerPoolSize = 10 * 1000

	// levelWorkerPoolMap are global goroutine pools which are
	// responsible for processing schema fields asynchronously.
	// Note that each dumping level gets a worker pool to avoid
	// dead lock.
	levelWorkerPoolMap sync.Map
)

var (
	ErrFailedToInitWorkerPool = errors.New("failed to init portal worker pool")
)

type (
	// processFunc is a callback function to be called in a worker.
	// It accepts user defined payload and returns user expected result.
	processFunc func(payload interface{}) (interface{}, error)
	jobRequest  struct {
		ctx        context.Context
		wg         *sync.WaitGroup
		payload    interface{}
		pf         processFunc
		resultChan chan *jobResult
	}

	// jobResult contains the result data and an optional error.
	jobResult struct {
		Data interface{}
		Err  error
	}
)

// submitJobs submits jobs to the worker pool and return the collected results.
func submitJobs(ctx context.Context, pf processFunc, payloads ...interface{}) (<-chan *jobResult, error) {
	logger.Debugf("[portal.pool] submit jobs with %d payloads", len(payloads))
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	level := dumpDepthFromContext(ctx)

	var workerPool *ants.PoolWithFunc
	pool, ok := levelWorkerPoolMap.Load(level)
	if ok {
		workerPool = pool.(*ants.PoolWithFunc)
	} else {
		logger.Debugf("[portal.pool] worker pool with level %d not found, try to create a new one", level)
		pool, err := ants.NewPoolWithFunc(1, processRequest)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		levelWorkerPoolMap.Store(level, pool)
		workerPool = pool
		SetMaxPoolSize(maxWorkerPoolSize)
	}

	resultChan := make(chan *jobResult, len(payloads))
	for _, payload := range payloads {
		wg.Add(1)
		err := workerPool.Invoke(&jobRequest{
			ctx:        ctx,
			wg:         &wg,
			payload:    payload,
			pf:         pf,
			resultChan: resultChan,
		})
		if err != nil {
			cancel()
			return nil, errors.WithStack(ErrFailedToInitWorkerPool)
		}
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	results := make(chan *jobResult, len(payloads))
	for result := range resultChan {
		if result.Err != nil {
			cancel()
		}

		results <- result
	}
	close(results)

	return results, nil
}

// SetMaxPoolSize limits the capacity of all worker pools.
func SetMaxPoolSize(size int) {
	logger.Debugf("[portal.pool] set max worker pool size to %d", size)
	if size == 0 {
		maxWorkerPoolSize = 1
	}

	maxWorkerPoolSize = size

	var length int
	levelWorkerPoolMap.Range(func(key, value interface{}) bool {
		length++
		return true
	})

	if length == 0 {
		return
	}

	// make sure capacity is valid.
	capacity := size / length
	if capacity == 0 {
		capacity = 1
	}

	levelWorkerPoolMap.Range(func(level, value interface{}) bool {
		pool, ok := value.(*ants.PoolWithFunc)
		if ok {
			logger.Debugf("[portal.pool] tune pool.%d capacity to %d", level.(int), capacity)
			pool.Tune(capacity)
		}
		return true
	})
}

// CleanUp releases the global worker pool.
// You should call this function only once before the main goroutine exits.
func CleanUp() {
	levelWorkerPoolMap.Range(func(key, value interface{}) bool {
		pool, ok := value.(*ants.PoolWithFunc)
		if ok {
			pool.Release()
		}
		return true
	})
}

func processRequest(request interface{}) {
	switch req := request.(type) {
	case *jobRequest:
		defer req.wg.Done()

		select {
		case <-req.ctx.Done():
		case req.resultChan <- func() *jobResult {
			data, err := func() (data interface{}, err error) {
				defer func() {
					if p := recover(); p != nil {
						var buf [4096]byte
						n := runtime.Stack(buf[:], false)
						err = errors.New(fmt.Sprintf("%s", p))
						logger.Errorf("[portal.pool] worker crashed: %s\n%s\n", p, buf[:n])
					}
				}()

				data, err = req.pf(req.payload)
				return
			}()

			return &jobResult{Data: data, Err: err}
		}():
		}
	default:
		logger.Warnf("[portal.pool] invalid worker request: '%s'", request)
	}
}

func init() {
	p, err := ants.NewPoolWithFunc(
		maxWorkerPoolSize,
		processRequest,
	)
	if err != nil {
		panic(ErrFailedToInitWorkerPool)
	}

	levelWorkerPoolMap.Store(0, p)
}
