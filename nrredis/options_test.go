package nrredis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestWithHost will test the method WithHost()
func TestWithHost(t *testing.T) {
	t.Run("set host", func(t *testing.T) {
		o := WithHost("host")
		tObj := new(Option)
		assert.IsType(t, *tObj, o)

		c := createConfig([]Option{o})
		assert.NotNil(t, c)
		assert.Equal(t, "host", c.Host)
	})

	t.Run("set host (empty)", func(t *testing.T) {
		o := WithHost("")
		tObj := new(Option)
		assert.IsType(t, *tObj, o)

		c := createConfig([]Option{o})
		assert.NotNil(t, c)
		assert.Equal(t, "", c.Host)
	})
}

// TestWithDBName will test the method WithDBName()
func TestWithDBName(t *testing.T) {
	t.Run("set db name", func(t *testing.T) {
		o := WithDBName("db_name")
		tObj := new(Option)
		assert.IsType(t, *tObj, o)

		c := createConfig([]Option{o})
		assert.NotNil(t, c)
		assert.Equal(t, "db_name", c.DBName)
	})

	t.Run("set db name (empty)", func(t *testing.T) {
		o := WithDBName("")
		tObj := new(Option)
		assert.IsType(t, *tObj, o)

		c := createConfig([]Option{o})
		assert.NotNil(t, c)
		assert.Equal(t, "", c.DBName)
	})
}

// TestWithPortPathOrID will test the method WithPortPathOrID()
func TestWithPortPathOrID(t *testing.T) {
	t.Run("set port", func(t *testing.T) {
		o := WithPortPathOrID("port")
		tObj := new(Option)
		assert.IsType(t, *tObj, o)

		c := createConfig([]Option{o})
		assert.NotNil(t, c)
		assert.Equal(t, "port", c.PortPathOrID)
	})

	t.Run("set port (empty)", func(t *testing.T) {
		o := WithPortPathOrID("")
		tObj := new(Option)
		assert.IsType(t, *tObj, o)

		c := createConfig([]Option{o})
		assert.NotNil(t, c)
		assert.Equal(t, "", c.PortPathOrID)
	})
}
