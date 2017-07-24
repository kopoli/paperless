package util

import "sync"

// Options is an interface to get and set string-like options for components
type Options interface {
	Set(key string, value string)
	Get(key string, fallback string) string
	IsSet(key string) bool
}

// GetOptions gets the static options structure
func GetOptions() Options {
	return options
}

// NewOptions returns a new Options structure
func NewOptions() Options {
	return &optionMap{
		values: make(map[string]string),
	}
}

var options = NewOptions()

// optionMap implements the Options interface with a map
type optionMap struct {
	values map[string]string
	mutex  sync.Mutex
}

// Set sets the option key with value
func (o *optionMap) Set(key string, value string) {
	o.mutex.Lock()
	o.values[key] = value
	o.mutex.Unlock()
}

// Get gets the value of a key or if not available, returns the fallback
func (o *optionMap) Get(key string, fallback string) string {
	o.mutex.Lock()
	ret, ok := o.values[key]
	if !ok {
		ret = fallback
	}
	o.mutex.Unlock()

	return ret
}

// IsSet returns true if the key has been set
func (o *optionMap) IsSet(key string) (ret bool) {
	o.mutex.Lock()
	_, ret = o.values[key]
	o.mutex.Unlock()
	return
}
