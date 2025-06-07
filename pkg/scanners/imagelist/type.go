package main

type (
	Config struct {
		uncompliantImageList []string `yaml:"uncompliantImageList,omitempty"`
	}
)

func DefaultConfig() *Config {
	return &Config{
		uncompliantImageList: []string{},
	}
}
