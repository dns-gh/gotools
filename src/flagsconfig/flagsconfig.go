// Package flagsconfig enables to save user defined key-value pairs
// and flags into a configuration file written in json format.
package flagsconfig

import (
	"flag"
	"fmt"
	"os"
	"tojson"
)

// Config represents the configuration file overloaded with
// user defined flags.
// There can be no configuration file and just flags.
type Config struct {
	path   string
	flags  map[string]string
	filter map[string]struct{}
}

func (c *Config) saveFlags() error {
	// loop through used and user defined flags
	// except for the filtered ones and update
	// the flags map from the Config
	flag.VisitAll(func(flag *flag.Flag) {
		name := flag.Name
		if _, ok := c.filter[name]; !ok {
			c.flags[name] = flag.Value.String()
		}
	})
	// save the flags map into a configuration file if any
	if c.path != "" {
		return tojson.Save(c.path, &c.flags)
	}
	return nil
}

func (c *Config) updateFlags(path string) error {
	c.path = path
	c.flags = make(map[string]string)
	// load configuration file data if any into a flags map
	if path != "" {
		err := tojson.Load(c.path, &c.flags)
		if err != nil {
			return err
		}
	}
	// loop through used flags
	// and remove them from the flags map
	flag.Visit(func(flag *flag.Flag) {
		_, ok := c.flags[flag.Name]
		if ok {
			delete(c.flags, flag.Name)
		}
	})
	// loop through used flags and user defined flags
	// and update the flags map
	flag.VisitAll(func(flag *flag.Flag) {
		val, ok := c.flags[flag.Name]
		if ok {
			flag.Value.Set(val)
		}
	})
	return nil
}

// Parse parses the given configuration file and overloads it with the
// currently used flags.
func (c *Config) Parse(path string) error {
	err := c.updateFlags(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load the config file: %v", err)
	}
	return c.saveFlags()
}

// NewConfig makes a Config given a config file path and a list
// of filtered flags not to appear in the config file.
func NewConfig(path string, filter ...string) (*Config, error) {
	c := &Config{
		filter: make(map[string]struct{}),
	}
	for _, v := range filter {
		c.filter[v] = struct{}{}
	}
	return c, c.Parse(path)
}

// Update updates a pair of key-value flags
// from the flags map of Config and update
// the configuration file if any.
func (c *Config) Update(key, value string) error {
	c.flags[key] = value
	if c.path != "" {
		return tojson.Save(c.path, &c.flags)
	}
	return nil
}

// Get gets a flag from a given key or
// return 0 value for a string if none.
func (c *Config) Get(key string) string {
	value, ok := c.flags[key]
	if ok {
		return value
	}
	return ""
}
