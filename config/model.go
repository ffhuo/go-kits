package config

// DBConfig represents common database configuration
type DBConfig struct {
	Debug       bool `config:"debug" json:"debug" yaml:"debug"`
	MaxIdle     int  `config:"max_idle" json:"max_idle" yaml:"max_idle"`
	MaxOpen     int  `config:"max_open" json:"max_open" yaml:"max_open"`
	MaxLifetime int  `config:"max_lifetime" json:"max_lifetime" yaml:"max_lifetime"`
}

// MySQLConfig represents MySQL database configuration
type MySQLConfig struct {
	DBConfig `config:",squash"`
	DSN      string `config:"dsn" json:"dsn" yaml:"dsn"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Addresses  []string `config:"addresses" json:"addresses" yaml:"addresses"`
	Database   int      `config:"database" json:"database" yaml:"database"`
	Password   string   `config:"password" json:"password" yaml:"password"`
	MaxRetries int      `config:"max_retries" json:"max_retries" yaml:"max_retries"`
}

// MongoDBConfig represents MongoDB configuration
type MongoDBConfig struct {
	URI      string `config:"uri" json:"uri" yaml:"uri"`
	Database string `config:"database" json:"database" yaml:"database"`
	Timeout  int    `config:"timeout" json:"timeout" yaml:"timeout"`
}

// LogConfig represents logging configuration
type LogConfig struct {
	Level      string `config:"level" json:"level" yaml:"level"`
	Format     string `config:"format" json:"format" yaml:"format"`
	Output     string `config:"output" json:"output" yaml:"output"`
	MaxSize    int    `config:"max_size" json:"max_size" yaml:"max_size"`
	MaxBackups int    `config:"max_backups" json:"max_backups" yaml:"max_backups"`
	MaxAge     int    `config:"max_age" json:"max_age" yaml:"max_age"`
	Compress   bool   `config:"compress" json:"compress" yaml:"compress"`
}
