package config

import(
	"os"
)

type ConfigAerospike struct {
	Aerospike_Host string
}

func LoadConfigAerospike() *ConfigAerospike {
	return &ConfigAerospike{
		Aerospike_Host: os.Getenv("AEROSPIKE_HOST"),
	}
}