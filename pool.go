package portal

import (
	"context"
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
	levelWorkerPoolMap     = make(map[int]*ants.PoolWithFunc)
	lockLevelWorkerPoolMap sync.Mutex
)

var (
	ErrFailedToInitWorkerPool = errors.New("failed to init portal worker pool")
)

type (
	ProcessFunc   func(payload interface{}) (interface{}, error)
	WorkerRequest struct {
		ctx        context.Context
		wg         *sync.WaitGroup
		payload    interface{}
		pf         ProcessFunc
		resultChan chan *WorkerResult
	}

	WorkerResult struct {
		Data interface{}
		Err  error
	}
)

func SubmitJobs(ctx context.Context, pf ProcessFunc, payloads ...interface{}) ([]*WorkerResult, error) {
	logger.Debugf("[portal.pool] submit jobs with %d payloads", len(payloads))
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	level := DumpDepthFromContext(ctx)
	workerPool, ok := levelWorkerPoolMap[level]
	if !ok {
		lockLevelWorkerPoolMap.Lock()
		logger.Debugf("[portal.pool] worker pool with level %d not found, try to create a new one", level)
		pool, err := ants.NewPoolWithFunc(1, processRequest)
		if err != nil {
			lockLevelWorkerPoolMap.Unlock()
			return nil, errors.WithStack(err)
		}

		levelWorkerPoolMap[level] = pool
		workerPool = pool
		TuneMaxPoolSize(maxWorkerPoolSize)
		lockLevelWorkerPoolMap.Unlock()
	}

	resultChan := make(chan *WorkerResult, len(payloads))
	for _, payload := range payloads {
		wg.Add(1)
		err := workerPool.Invoke(&WorkerRequest{
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

	results := make([]*WorkerResult, 0, len(payloads))
	for result := range resultChan {
		if result.Err != nil {
			cancel()
		}

		results = append(results, result)
	}

	return results, nil
}

// TuneMaxPoolSize limits the capacity of all worker pools.
func TuneMaxPoolSize(size int) {
	logger.Debugf("[portal.pool] set max worker pool size to %d", size)
	if size == 0 {
		maxWorkerPoolSize = 1
	}

	maxWorkerPoolSize = size
	if len(levelWorkerPoolMap) == 0 {
		return
	}

	// make sure capacity is valid.
	capacity := size / len(levelWorkerPoolMap)
	if capacity == 0 {
		capacity = 1
	}

	for level, pool := range levelWorkerPoolMap {
		logger.Debugf("[portal.pool] tune pool.%d capacity to %d", level, capacity)
		pool.Tune(capacity)
	}
}

// CleanUp releases the global worker pool.
// You should call this function only once before the main goroutine exits.
func CleanUp() {
	for _, pool := range levelWorkerPoolMap {
		pool.Release()
	}
}

func processRequest(request interface{}) {
	switch req := request.(type) {
	case *WorkerRequest:
		defer req.wg.Done()

		select {
		case <-req.ctx.Done():
		case req.resultChan <- func() *WorkerResult {
			data, err := req.pf(req.payload)
			return &WorkerResult{Data: data, Err: err}
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

	levelWorkerPoolMap[0] = p
}
