# go-cache
**go-cache** is a simple redis cache dependency system on-top of the famous [redigo](https://github.com/gomodule/redigo) package

[![Go](https://img.shields.io/github/go-mod/go-version/mrz1836/go-cache)](https://golang.org/)
[![Build Status](https://travis-ci.org/mrz1836/go-cache.svg?branch=master)](https://travis-ci.org/mrz1836/go-cache)
[![Report](https://goreportcard.com/badge/github.com/mrz1836/go-cache?style=flat)](https://goreportcard.com/report/github.com/mrz1836/go-cache)
[![Release](https://img.shields.io/github/release-pre/mrz1836/go-cache.svg?style=flat)](https://github.com/mrz1836/go-cache/releases)
[![GoDoc](https://godoc.org/github.com/mrz1836/go-cache?status.svg&style=flat)](https://pkg.go.dev/github.com/mrz1836/go-cache?tab=doc)

## Table of Contents
- [Installation](#installation)
- [Documentation](#documentation)
- [Examples & Tests](#examples--tests)
- [Benchmarks](#benchmarks)
- [Code Standards](#code-standards)
- [Usage](#usage)
- [Maintainers](#maintainers)
- [Contributing](#contributing)
- [License](#license)

## Installation

**go-cache** requires a [supported release of Go](https://golang.org/doc/devel/release.html#policy).
```bash
$ go get -u github.com/mrz1836/go-cache
```

### Features
- Cache Dependencies Between Keys (toggle functionality)
- Connect via URL
- Better Pool Management & Creation
- Register Scripts
- Helper Methods (Get, Set, HashGet, etc)
- Basic Lock/Release (from [bgentry lock.go](https://gist.github.com/bgentry/6105288))

## Documentation
You can view the generated [documentation here](https://pkg.go.dev/github.com/mrz1836/go-cache?tab=doc).

<details>
<summary><strong><code>Library Deployment</code></strong></summary>

[goreleaser](https://github.com/goreleaser/goreleaser) for easy binary or library deployment to Github and can be installed via: `brew install goreleaser`.

The [.goreleaser.yml](.goreleaser.yml) file is used to configure [goreleaser](https://github.com/goreleaser/goreleaser).

Use `make release-snap` to create a snapshot version of the release, and finally `make release` to ship to production.
</details>

<details>
<summary><strong><code>Makefile Commands</code></strong></summary>

View all `makefile` commands
```bash
$ make help
```

List of all current commands:
```text
all                            Runs test, install, clean, docs
bench                          Run all benchmarks in the Go application
clean                          Remove previous builds and any test cache data
clean-mods                     Remove all the Go mod cache
coverage                       Shows the test coverage
godocs                         Sync the latest tag with GoDocs
help                           Show all make commands available
lint                           Run the Go lint application
release                        Full production release (creates release in Github)
release-test                   Full production test release (everything except deploy)
release-snap                   Test the full release (build binaries)
tag                            Generate a new tag and push (IE: make tag version=0.0.0)
tag-remove                     Remove a tag if found (IE: make tag-remove version=0.0.0)
tag-update                     Update an existing tag to current commit (IE: make tag-update version=0.0.0)
test                           Runs vet, lint and ALL tests
test-short                     Runs vet, lint and tests (excludes integration tests)
update                         Update all project dependencies
update-releaser                Update the goreleaser application
vet                            Run the Go vet application
```
</details>

<details>
<summary><strong><code>Package Dependencies</code></strong></summary>

- Gary Burd's [Redigo](https://github.com/gomodule/redigo)
</details>

## Examples & Tests
All unit tests and [examples](examples/examples.go) run via [Travis CI](https://travis-ci.org/mrz1836/go-cache) and uses [Go version 1.14.x](https://golang.org/doc/go1.14). View the [deployment configuration file](.travis.yml).

Run all tests (including integration tests)
```bash
$ make test
```

Run tests (excluding integration tests)
```bash
$ make test-short
```

Run the [examples](examples/examples.go):
```bash
$ make run-examples
```

## Benchmarks
Run the Go benchmarks:
```bash
$ make bench
```

## Code Standards
Read more about this Go project's [code standards](CODE_STANDARDS.md).

## Usage
View the [examples](examples/examples.go)

Basic implementation:
```golang
package main

import (
	"log"

	"github.com/mrz1836/go-cache"
)

func main() {

	// Create the pool and first connection
	_ = cache.Connect("redis://localhost:6379", 0, 10, 0, 240, true, redis.DialKeepAlive(10*time.Second))

	// Set a key
	_ = cache.Set("key-name", "the-value", "dependent-key-1", "dependent-key-2")

	// Get a key
	value, _ := cache.Get("key-name")
	log.Println("Got value:", value)
	//Output: Got Value: the-value

	// Kill keys by dependency
	keys, _ := cache.KillByDependency("dependent-key-1")
	log.Println("Keys Removed:", keys)
	//Output: Keys Removed: 2
}
```

## Maintainers

| [<img src="https://github.com/mrz1836.png" height="50" alt="MrZ" />](https://github.com/mrz1836) | [<img src="https://github.com/kayleg.png" height="50" alt="MrZ" />](https://github.com/kayleg) |
|:---:|:---:|
| [MrZ](https://github.com/mrz1836) | [kayleg](https://github.com/kayleg) |

## Contributing

This project uses Gary Burd's [Redigo](https://github.com/gomodule/redigo) package.

View the [contributing guidelines](CONTRIBUTING.md) and follow the [code of conduct](CODE_OF_CONDUCT.md).

Support the development of this project 🙏

[![Donate](https://img.shields.io/badge/donate-bitcoin-brightgreen.svg)](https://mrz1818.com/?tab=tips&af=go-cache)

## License

![License](https://img.shields.io/github/license/mrz1836/go-cache.svg?style=flat)