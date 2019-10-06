package portal

import "github.com/pkg/errors"

type Option func(c *Chell) error

func Only(fields ...string) Option {
	return func(c *Chell) error {
		filters, err := ParseFilters(fields)
		if err != nil {
			return errors.WithStack(err)
		}
		c.onlyFieldFilters = filters
		return nil
	}
}

func Exclude(fields ...string) Option {
	return func(c *Chell) error {
		filters, err := ParseFilters(fields)
		if err != nil {
			return errors.WithStack(err)
		}
		c.excludeFieldFilters = filters
		return nil
	}
}

func WorkerPoolSize(size int) Option {
	return func(c *Chell) error {
		if size > 0 {
			c.workerPoolSize = size
		}
		return nil
	}
}
