// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Config struct {
	//Period time.Duration `config:"period"`
	Host           string                 `config:"host" validate:"required"`
	LineDelimiter  string                 `config:"line_delimiter" validate:"nonzero"`
	Timeout        time.Duration          `config:"timeout" validate:"nonzero,positive"`
	MaxMessageSize uint64                 `config:"max_message_size" validate:"nonzero,positive"`

	Labels         map[string]interface{} `config:"labels"`
	LabelKey       string                 `config:"label_key"`
}


var defaultLogTypes = map[string]int32{}
var defaultLabels = map[string]interface{}{}

var DefaultConfig = Config{
	LineDelimiter:  "\n",
	Timeout:        time.Minute * 5,
	MaxMessageSize: 20 * 1024 * 1024,
	Labels:         defaultLabels,
	LabelKey:       "labels",
}

