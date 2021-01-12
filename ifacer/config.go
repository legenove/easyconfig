package ifacer

import (
	"github.com/magiconair/properties"
	"time"
)

type ChangeFunc func(iv Configer)
type OptionFunc func(iv Configer)

type ConfigManager interface {
	Instance(name, parserType string, val interface{}, opts ...OptionFunc) (Configer, error)
}

type ConfigChanger interface {
	SetOnChangeFunc(onChangeFunc ChangeFunc)
	SetOnRemoveFunc(onRemoveFunc ChangeFunc)
	OnChangeChan() <-chan struct{}
}

type ConfigForPropertier interface {
	GetProperties() *properties.Properties
	SetProperties(p *properties.Properties)
}

type ConfigReader interface {
	AllKeys() []string
	GetValue() interface{}
	Get(key string) (interface{}, error)
	GetString(key string) (string, error)
	GetBool(key string) (bool, error)
	GetInt(key string) (int, error)
	GetInt32(key string) (int32, error)
	GetInt64(key string) (int64, error)
	GetFloat64(key string) (float64, error)
	GetTime(key string) (time.Time, error)
	GetDuration(key string) (time.Duration, error)
	GetStringSlice(key string) ([]string, error)
	GetStringMap(key string) (map[string]interface{}, error)
	GetStringMapString(key string) (map[string]string, error)
	GetStringMapStringSlice(key string) (map[string][]string, error)
	GetSizeInBytes(key string) (uint, error)

	UnmarshalKey(key string, rawVal interface{}) error
}

type ConfigWriter interface {
	// todo
}

type Configer interface {
	ConfigReader
	ConfigWriter
	ConfigChanger
	GetName() string
	GetFullName() string
	GetConfType() string
	SetConfType(t string)
}
