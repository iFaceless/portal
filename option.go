package portal

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
