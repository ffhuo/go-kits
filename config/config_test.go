package config

import (
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestConfig_LoadFile(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		setup    func() *Config
		validate func(*testing.T, *Config)
		wantErr  bool
	}{
		{
			name:  "load single yaml file",
			files: []string{"example/configs/config.yaml"},
			setup: func() *Config {
				return New().AddPath(".")
			},
			validate: func(t *testing.T, cfg *Config) {
				// Test MySQL config
				var mysqlConfig MySQLConfig
				if err := cfg.Unmarshal("mysql", &mysqlConfig); err != nil {
					t.Errorf("Failed to unmarshal MySQL config: %v", err)
				}
				expectedMySQL := MySQLConfig{
					DBConfig: DBConfig{
						Debug:       true,
						MaxIdle:     10,
						MaxOpen:     100,
						MaxLifetime: 3600,
					},
					DSN: "root:password@tcp(localhost:3306)/myapp?charset=utf8mb4&parseTime=True&loc=Local",
				}
				if mysqlConfig != expectedMySQL {
					t.Errorf("MySQL config = %+v, want %+v", mysqlConfig, expectedMySQL)
				}

				// Test Redis config
				var redisConfig RedisConfig
				if err := cfg.Unmarshal("redis", &redisConfig); err != nil {
					t.Errorf("Failed to unmarshal Redis config: %v", err)
				}
				expectedRedis := RedisConfig{
					Addresses:  []string{"localhost:6379"},
					Database:   0,
					Password:   "redis-password",
					MaxRetries: 3,
				}
				if redisConfig.Password != expectedRedis.Password ||
					redisConfig.Database != expectedRedis.Database ||
					len(redisConfig.Addresses) != len(expectedRedis.Addresses) {
					t.Errorf("Redis config = %+v, want %+v", redisConfig, expectedRedis)
				}

				// Test MongoDB config
				var mongoConfig MongoDBConfig
				if err := cfg.Unmarshal("mongodb", &mongoConfig); err != nil {
					t.Errorf("Failed to unmarshal MongoDB config: %v", err)
				}
				expectedMongo := MongoDBConfig{
					URI:      "mongodb://localhost:27017",
					Database: "myapp",
					Timeout:  10,
				}
				if mongoConfig != expectedMongo {
					t.Errorf("MongoDB config = %+v, want %+v", mongoConfig, expectedMongo)
				}
			},
			wantErr: false,
		},
		{
			name:  "load and merge two files",
			files: []string{"example/configs/config.yaml", "example/configs/config.local.yaml"},
			setup: func() *Config {
				return New().AddPath(".")
			},
			validate: func(t *testing.T, cfg *Config) {
				// Test MySQL config (should be merged)
				var mysqlConfig MySQLConfig
				if err := cfg.Unmarshal("mysql", &mysqlConfig); err != nil {
					t.Errorf("Failed to unmarshal MySQL config: %v", err)
				}
				expectedMySQL := MySQLConfig{
					DBConfig: DBConfig{
						Debug:       false,
						MaxIdle:     10,
						MaxOpen:     100,
						MaxLifetime: 3600,
					},
					DSN: "root:password@tcp(localhost:3306)/myapp_local?charset=utf8mb4&parseTime=True&loc=Local",
				}
				if mysqlConfig != expectedMySQL {
					t.Errorf("MySQL config = %+v, want %+v", mysqlConfig, expectedMySQL)
				}

				// Test Redis config (should be merged)
				var redisConfig RedisConfig
				if err := cfg.Unmarshal("redis", &redisConfig); err != nil {
					t.Errorf("Failed to unmarshal Redis config: %v", err)
				}
				expectedRedis := RedisConfig{
					Addresses:  []string{"localhost:6379", "localhost:6380"},
					Database:   0,
					Password:   "local-password",
					MaxRetries: 3,
				}
				if redisConfig.Password != expectedRedis.Password ||
					redisConfig.Database != expectedRedis.Database ||
					len(redisConfig.Addresses) != len(expectedRedis.Addresses) {
					t.Errorf("Redis config = %+v, want %+v", redisConfig, expectedRedis)
				}

				// Test MongoDB config (should be merged)
				var mongoConfig MongoDBConfig
				if err := cfg.Unmarshal("mongodb", &mongoConfig); err != nil {
					t.Errorf("Failed to unmarshal MongoDB config: %v", err)
				}
				expectedMongo := MongoDBConfig{
					URI:      "mongodb://localhost:27017",
					Database: "myapp_local",
					Timeout:  10,
				}
				if mongoConfig != expectedMongo {
					t.Errorf("MongoDB config = %+v, want %+v", mongoConfig, expectedMongo)
				}
			},
			wantErr: false,
		},
		{
			name:  "non-existent file",
			files: []string{"non_existent.yaml"},
			setup: func() *Config {
				return New().AddPath("example/configs")
			},
			validate: func(t *testing.T, cfg *Config) {},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setup()
			var err error

			// Load first file
			err = cfg.LoadFile(tt.files[0])
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If there's a second file, merge it
			if len(tt.files) > 1 && err == nil {
				err = cfg.MergeFile(tt.files[1])
				if (err != nil) != tt.wantErr {
					t.Errorf("MergeFile() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}

			if err == nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestConfig_GetValues(t *testing.T) {
	cfg := New().AddPath("example/configs")
	if err := cfg.LoadFile("config.yaml"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		key      string
		want     interface{}
		validate func(interface{}) bool
	}{
		{
			name: "get string value",
			key:  "mysql.dsn",
			want: "root:password@tcp(localhost:3306)/myapp?charset=utf8mb4&parseTime=True&loc=Local",
			validate: func(v interface{}) bool {
				s, ok := v.(string)
				return ok && s == "root:password@tcp(localhost:3306)/myapp?charset=utf8mb4&parseTime=True&loc=Local"
			},
		},
		{
			name: "get bool value",
			key:  "mysql.debug",
			want: true,
			validate: func(v interface{}) bool {
				b, ok := v.(bool)
				return ok && b == true
			},
		},
		{
			name: "get int value",
			key:  "mysql.max_idle",
			want: 10,
			validate: func(v interface{}) bool {
				i, ok := v.(int)
				return ok && i == 10
			},
		},
		{
			name: "get string slice",
			key:  "redis.addresses",
			want: []string{"localhost:6379"},
			validate: func(v interface{}) bool {
				s, ok := v.([]string)
				return ok && len(s) == 1 && s[0] == "localhost:6379"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.Get(tt.key)
			if !tt.validate(got) {
				t.Errorf("Get(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestConfig_Watch(t *testing.T) {
	cfg := New().AddPath("example/configs")
	if err := cfg.LoadFile("config.yaml"); err != nil {
		t.Fatal(err)
	}

	changed := make(chan bool)
	cfg.Watch(func(event fsnotify.Event) {
		if event.Op&fsnotify.Write == fsnotify.Write {
			changed <- true
		}
	})

	// 等待文件系统事件准备就绪
	time.Sleep(time.Millisecond * 100)

	// 更新配置文件
	if err := cfg.MergeFile("config.local.yaml"); err != nil {
		t.Fatal(err)
	}

	select {
	case <-changed:
		var mysqlConfig MySQLConfig
		if err := cfg.Unmarshal("mysql", &mysqlConfig); err != nil {
			t.Fatal(err)
		}
		if mysqlConfig.Debug {
			t.Error("Watch() failed to update MySQL debug setting")
		}
		if mysqlConfig.DSN != "root:password@tcp(localhost:3306)/myapp_local?charset=utf8mb4&parseTime=True&loc=Local" {
			t.Error("Watch() failed to update MySQL DSN")
		}
	case <-time.After(time.Second):
		t.Error("Watch() failed to detect file change")
	}
}

func TestConfig_EnvOverride(t *testing.T) {
	t.Setenv("APP_MYSQL_DSN", "root:envpass@tcp(localhost:3306)/envdb")
	t.Setenv("APP_REDIS_PASSWORD", "env-redis-pass")
	t.Setenv("APP_MONGODB_URI", "mongodb://localhost:27018")

	cfg := New().
		AddPath("example/configs").
		SetEnvPrefix("APP").
		AutomaticEnv()

	if err := cfg.LoadFile("config.yaml"); err != nil {
		t.Fatal(err)
	}

	// Test MySQL DSN override
	var mysqlConfig MySQLConfig
	if err := cfg.Unmarshal("mysql", &mysqlConfig); err != nil {
		t.Fatal(err)
	}
	if mysqlConfig.DSN != "root:envpass@tcp(localhost:3306)/envdb" {
		t.Errorf("Environment variable override failed for MySQL DSN, got %s", mysqlConfig.DSN)
	}

	// Test Redis password override
	var redisConfig RedisConfig
	if err := cfg.Unmarshal("redis", &redisConfig); err != nil {
		t.Fatal(err)
	}
	if redisConfig.Password != "env-redis-pass" {
		t.Errorf("Environment variable override failed for Redis password, got %s", redisConfig.Password)
	}

	// Test MongoDB URI override
	var mongoConfig MongoDBConfig
	if err := cfg.Unmarshal("mongodb", &mongoConfig); err != nil {
		t.Fatal(err)
	}
	if mongoConfig.URI != "mongodb://localhost:27018" {
		t.Errorf("Environment variable override failed for MongoDB URI, got %s", mongoConfig.URI)
	}
}
