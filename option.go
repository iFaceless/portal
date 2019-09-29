package portal

type Option func(c *Chell)

func Only(fields ...string) Option {
	return func(c *Chell) {
		c.onlyFieldNames = fields
		c.onlyFieldFilters = ParseFilters(fields)
	}
}

func Exclude(fields ...string) Option {
	return func(c *Chell) {
		c.excludedFieldNames = fields
		c.excludeFieldFilters = ParseFilters(fields)
	}
}
