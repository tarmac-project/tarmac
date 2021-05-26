// Package redis is a Hord database driver for Redis. This package satisfies the Hord interface
// and can be used to interact with both Open Source Redis and Enterprise Redis.
//
//  // Connect to Redis
//  db, err := redis.Dial(&redis.Config{})
//  if err != nil {
//    // do stuff
//  }
//
//  // Setup and Initialize the Keyspace if necessary
//  err = db.Setup()
//  if err != nil {
//    // do stuff
//  }
//
//  // Write data to the cluster
//  err = db.Set("mykey", []byte("My Data"))
//  if err != nil {
//    // do stuff
//  }
//
//  // Fetch the same data
//  d, err := db.Get("mykey")
//  if err != nil {
//    // do stuff
//  }
//
package redis

import (
	"crypto/tls"
	"fmt"
	"github.com/FZambia/sentinel"
	"github.com/gomodule/redigo/redis"
	"time"
)

// Config provides configuration options for connecting to and controlling the behavior of Redis.
type Config struct {
	// ConnectTimeout is used to specify a global connection timeout value.
	ConnectTimeout time.Duration

	// Database specifies the Redis database to connect to and use. If not set, default is 0.
	Database int

	// IdleTimeout will close idle connections that have remained idle beyond the specified time duration.
	IdleTimeout time.Duration

	// KeepAlive defines the TCP Keep-Alive interval for Redis connections. By default this set to 5 minutes, this
	// setting is useful for detecting TCP sessions that are stale.
	KeepAlive time.Duration

	// MaxActive is the maximum number of connections that can be allocated and used for the Redis connection pool.
	MaxActive int

	// MaxConnLifetime will set a maximum lifespan for connections. This setting will close connections in the pool
	// after the time duration is exceeded, regardless of whether the connection is active or not.
	MaxConnLifetime time.Duration

	// MaxIdle sets the maximum number of idle connections the Redis connection pool will allow.
	MaxIdle int

	// Password specifies the AUTH token to be used for Redis Authentication.
	Password string

	// ReadTimeout is used to specify a global read timeout for each Redis command.
	ReadTimeout time.Duration

	// SentinelConfig is used to configure Sentinel connection details. If not using Redis Sentinel or Discovery
	// Service, leave this blank.
	SentinelConfig SentinelConfig

	// Server specifies the Redis Server to connect to. If using Redis Sentinel or Discovery Service leave this
	// blank.
	Server string

	// SkipTLSVerify will disable the TLS hostname checking. Warning, using this setting opens the risk of
	// man-in-the-middle attacks.
	SkipTLSVerify bool

	// TLSConfig allows users to specify TLS settings for connecting to Redis. This is a standard TLS configuration
	// and can be used to configure 2-way TLS for Redis and Redis Sentinel.
	TLSConfig *tls.Config

	// WriteTimeout is used to specify a global write timeout for each Redis command.
	WriteTimeout time.Duration
}

// SentinelConfig can be used to configure the Redis client to connect using Redis Sentinel or Enterprise Redis
// Discovery Service.
type SentinelConfig struct {
	// Servers is a list of Sentinel servers to connect to and use for master discovery.
	Servers []string

	// Master is the name of the Redis master that the Sentinel Servers monitor.
	Master string
}

// Database is used to interface with Redis. It also satisfies the Hord Database interface.
type Database struct {
	// config holds the initial Redis configuration used to create this Database instance.
	config Config

	// pool is the Redis connection pool
	pool *redis.Pool

	// sentinel holds the Redis sentinel connections
	sentinel *sentinel.Sentinel
}

// Dial will establish a Redis connection pool using the configuration provided. It provides back an interface that
// satisfies the hord.Database interface.
func Dial(conf Config) (*Database, error) {
	db := &Database{
		config: conf,
	}

	// Verify that Either Server or Sentinel Servers is set
	if db.config.Server == "" && len(db.config.SentinelConfig.Servers) == 0 {
		return nil, fmt.Errorf("must specify either a Redis Server or Sentinel Pool")
	}

	// Setup Redis DailOptions
	opts := []redis.DialOption{}
	opts = append(opts, redis.DialConnectTimeout(db.config.ConnectTimeout))
	opts = append(opts, redis.DialDatabase(db.config.Database))
	opts = append(opts, redis.DialKeepAlive(db.config.KeepAlive))
	opts = append(opts, redis.DialPassword(db.config.Password))
	opts = append(opts, redis.DialReadTimeout(db.config.ReadTimeout))
	opts = append(opts, redis.DialWriteTimeout(db.config.WriteTimeout))
	if db.config.TLSConfig != nil {
		opts = append(opts, redis.DialUseTLS(true), redis.DialTLSConfig(db.config.TLSConfig))
		opts = append(opts, redis.DialTLSSkipVerify(db.config.SkipTLSVerify))
	}

	// If Sentinel is set, let's connect
	if len(db.config.SentinelConfig.Servers) > 0 {
		if db.config.SentinelConfig.Master == "" {
			return nil, fmt.Errorf("if using Sentinel the Redis Master must be defined")
		}
		db.sentinel = &sentinel.Sentinel{
			Addrs:      db.config.SentinelConfig.Servers,
			MasterName: db.config.SentinelConfig.Master,
			Dial: func(addr string) (redis.Conn, error) {
				c, err := redis.Dial("tcp", addr, opts...)
				if err != nil {
					return nil, err
				}
				return c, nil
			},
		}
	}

	// Create a Redis Connection Pool
	db.pool = &redis.Pool{
		IdleTimeout:     db.config.IdleTimeout,
		MaxActive:       db.config.MaxActive,
		MaxConnLifetime: db.config.MaxConnLifetime,
		MaxIdle:         db.config.MaxIdle,
		Wait:            true,
		// Used to create new connections for the pool
		Dial: func() (redis.Conn, error) {
			var err error
			server := db.config.Server
			if db.sentinel != nil {
				server, err = db.sentinel.MasterAddr()
				if err != nil {
					return nil, err
				}
			}
			c, err := redis.Dial("tcp", server, opts...)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		// Used to Test the provided connection
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if !sentinel.TestRole(c, "master") {
				return fmt.Errorf("server is not a master")
			}
			_, err := c.Do("PING")
			if err != nil {
				return fmt.Errorf("connection is unhealthy, failed ping %s", err)
			}
			return nil
		},
	}

	return db, nil
}

// Setup does nothing with Redis, this is only here to meet interface requirements.
func (db *Database) Setup() error {
	return nil
}

// Get is called to retrieve data from the database. This function will take in a key and return
// the data or any errors received from querying the database.
func (db *Database) Get(key string) ([]byte, error) {
	if key == "" {
		return nil, fmt.Errorf("key must not be empty")
	}

	c := db.pool.Get()
	defer c.Close()

	d, err := redis.Bytes(c.Do("GET", key))
	if err != nil {
		return nil, fmt.Errorf("unable to fetch data from Redis - %s", err)
	}
	return d, nil
}

// Set is called when data within the database needs to be updated or inserted. This function will
// take the data provided and create an entry within the database using the key as a lookup value.
func (db *Database) Set(key string, data []byte) error {
	if key == "" {
		return fmt.Errorf("key must not be empty")
	}
	if len(data) == 0 {
		return fmt.Errorf("data must not be empty")
	}

	c := db.pool.Get()
	defer c.Close()

	_, err := c.Do("SET", key, data)
	if err != nil {
		return fmt.Errorf("unable to write data to Redis - %s", err)
	}

	return nil
}

// Delete is called when data within the database needs to be deleted. This function will delete
// the data stored within the database for the specified key.
func (db *Database) Delete(key string) error {
	if key == "" {
		return fmt.Errorf("key must not be empty")
	}

	c := db.pool.Get()
	defer c.Close()

	_, err := c.Do("DEL", key)
	if err != nil {
		return fmt.Errorf("unable to remove key from Redis - %s", err)
	}
	return nil
}

// Keys is called to retrieve a list of keys stored within the database. This function will query
// the database returning all keys used within the hord database.
func (db *Database) Keys() ([]string, error) {
	c := db.pool.Get()
	defer c.Close()

	keys, err := redis.Strings(c.Do("KEYS", "*"))
	if err != nil {
		return keys, fmt.Errorf("unable to fetch keys from Redis - %s", err)
	}
	return keys, nil
}

// HealthCheck is used to verify connectivity and health of the database. This function
// simply runs a generic ping against the database. If the ping errors in any fashion this
// function will return an error.
func (db *Database) HealthCheck() error {
	c := db.pool.Get()
	defer c.Close()

	_, err := c.Do("PING")
	if err != nil {
		return fmt.Errorf("unable to ping Redis - %s", err)
	}
	return nil
}

// Close will close all connections to Redis and clean up the pool.
func (db *Database) Close() {
	if db == nil {
		return
	}
	defer db.pool.Close()
	if db.sentinel != nil {
		defer db.sentinel.Close()
	}
}
