package nrredis

// Config contains metadata to send to New Relic.
type Config struct {
	DBName       string
	Host         string
	PortPathOrID string
}

func createConfig(opts []Option) *Config {
	cfg := &Config{}
	for _, f := range opts {
		f(cfg)
	}
	return cfg
}

// Option configures a Config object.
type Option func(*Config)

// WithDBName sets a DB name.
func WithDBName(dbName string) Option {
	return func(c *Config) {
		c.DBName = dbName
	}
}

// WithHost sets a DB host.
func WithHost(host string) Option {
	return func(c *Config) {
		c.Host = host
	}
}

// WithPortPathOrID sets a DB port, path or id.
func WithPortPathOrID(v string) Option {
	return func(c *Config) {
		c.PortPathOrID = v
	}
}
