# Hord

[![Build 
Status](https://travis-ci.org/madflojo/hord.svg)](
https://travis-ci.org/madflojo/hord) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/madflojo/hord) [![Coverage Status](https://coveralls.io/repos/github/madflojo/hord/badge.svg?branch=master)](https://coveralls.io/github/madflojo/hord?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/madflojo/hord)](https://goreportcard.com/report/github.com/madflojo/hord) [![Documentation](https://godoc.org/github.com/madflojo/hord?status.svg)](http://godoc.org/github.com/madflojo/hord)


Hord provides a modular key-value interface for interacting with databases. The goal is to provide a consistent interface regardless, of the underlying database.

With this package, users can switch out the underlying database without major refactoring.

## Installation

To use Hord within your project you must first import the Hord interface itself.

```go
import "github.com/madflojo/hord"
```

Then import the database driver you wish to use

```go
import "github.com/madflojo/hord/driver/cassandra"
```

Available [drivers](drivers) are a follows:

* Cassandra
* Redis 

Our TODO list:

* Couchbase
* CockRoachDB
* MySQL
* TiKV
* PostgreSQL


## Usage

The below example shows using Hord to connect and interact with Cassandra.

```go
import "github.com/madflojo/hord"
import "github.com/madflojo/hord/driver/cassandra"

func main() {
  // Define our DB Interface
  var db hord.Database

  // Connect to a Cassandra Cluster
  db, err := cassandra.Dial(&cassandra.Config{})
  if err != nil {
    // do stuff
  }

  // Setup and Initialize the Keyspace if necessary
  err = db.Setup()
  if err != nil {
    // do stuff
  }

  // Write data to the cluster
  err = db.Set("mykey", []byte("My Data"))
  if err != nil {
    // do stuff
  }

  // Fetch the same data
  d, err := db.Get("mykey")
  if err != nil {
    // do stuff
  }
}
```

## Contributing
Thank you for your interest in helping develop Hord. The time, skills, and perspectives you contribute to this project are valued.

Please reference our [Contributing Guide](CONTRIBUTING.md) for details.

## License
[Apache License 2.0](https://choosealicense.com/licenses/apache-2.0/)
