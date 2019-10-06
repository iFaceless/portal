package portal

import "time"

type Option func(c *Chell)

func Only(fields ...string) Option {
	return func(c *Chell) {
		filters, err := ParseFilters(fields)
		if err != nil {
			c.err = err
		} else {
			c.onlyFieldFilters = filters
		}
	}
}

func Exclude(fields ...string) Option {
	return func(c *Chell) {
		filters, err := ParseFilters(fields)
		if err != nil {
			c.err = err
		} else {
			c.excludeFieldFilters = filters
		}
	}
}

func WorkerPoolSize(size int) Option {
	return func(c *Chell) {
		if size > 0 {
			c.workerPoolSize = size
		}
	}
}

func WorkerTimeout(d time.Duration) Option {
	return func(c *Chell) {
		c.workerTimeout = d
	}
}
