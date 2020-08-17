# go-cache
> Simple cache dependency system on-top of the famous [redigo](https://github.com/gomodule/redigo) package

[![Release](https://img.shields.io/github/release-pre/mrz1836/go-cache.svg?logo=github&style=flat)](https://github.com/mrz1836/go-cache/releases)
[![Build Status](https://travis-ci.com/mrz1836/go-cache.svg?branch=master)](https://travis-ci.com/mrz1836/go-cache)
[![Report](https://goreportcard.com/badge/github.com/mrz1836/go-cache?style=flat)](https://goreportcard.com/report/github.com/mrz1836/go-cache)
[![codecov](https://codecov.io/gh/mrz1836/go-cache/branch/master/graph/badge.svg)](https://codecov.io/gh/mrz1836/go-cache)
[![Go](https://img.shields.io/github/go-mod/go-version/mrz1836/go-cache)](https://golang.org/)
[![Sponsor](https://img.shields.io/badge/sponsor-MrZ-181717.svg?logo=github&style=flat&v=3)](https://github.com/sponsors/mrz1836)
[![Donate](https://img.shields.io/badge/donate-bitcoin-ff9900.svg?logo=bitcoin&style=flat)](https://mrz1818.com/?tab=tips&af=go-cache)

<br/>

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

<br/>

## Installation

**go-cache** requires a [supported release of Go](https://golang.org/doc/devel/release.html#policy) and [redis](https://formulae.brew.sh/formula/redis).
```shell script
go get -u github.com/mrz1836/go-cache
```

<br/>

## Documentation
View the generated [documentation](https://pkg.go.dev/github.com/mrz1836/go-cache?tab=doc)

[![GoDoc](https://godoc.org/github.com/mrz1836/go-cache?status.svg&style=flat)](https://pkg.go.dev/github.com/mrz1836/go-cache?tab=doc)

### Features
- Cache Dependencies Between Keys (toggle functionality)
- Connect via URL
- Better Pool Management & Creation
- Register Scripts
- Helper Methods (Get, Set, HashGet, etc)
- Basic Lock/Release (from [bgentry lock.go](https://gist.github.com/bgentry/6105288))

<details>
<summary><strong><code>Library Deployment</code></strong></summary>
<br/>

[goreleaser](https://github.com/goreleaser/goreleaser) for easy binary or library deployment to Github and can be installed via: `brew install goreleaser`.

The [.goreleaser.yml](.goreleaser.yml) file is used to configure [goreleaser](https://github.com/goreleaser/goreleaser).

Use `make release-snap` to create a snapshot version of the release, and finally `make release` to ship to production.
</details>

<details>
<summary><strong><code>Makefile Commands</code></strong></summary>
<br/>

View all `makefile` commands
```shell script
make help
```

List of all current commands:
```text
all                    Runs multiple commands
clean                  Remove previous builds and any test cache data
clean-mods             Remove all the Go mod cache
coverage               Shows the test coverage
godocs                 Sync the latest tag with GoDocs
help                   Show this help message
install                Install the application
install-go             Install the application (Using Native Go)
lint                   Run the Go lint application
release                Full production release (creates release in Github)
release                Runs common.release then runs godocs
release-snap           Test the full release (build binaries)
release-test           Full production test release (everything except deploy)
replace-version        Replaces the version in HTML/JS (pre-deploy)
run-examples           Runs all the examples
tag                    Generate a new tag and push (tag version=0.0.0)
tag-remove             Remove a tag if found (tag-remove version=0.0.0)
tag-update             Update an existing tag to current commit (tag-update version=0.0.0)
test                   Runs vet, lint and ALL tests
test-short             Runs vet, lint and tests (excludes integration tests)
test-travis            Runs all tests via Travis (also exports coverage)
test-travis-short      Runs unit tests via Travis (also exports coverage)
uninstall              Uninstall the application (and remove files)
vet                    Run the Go vet application
```
</details>

<details>
<summary><strong><code>Package Dependencies</code></strong></summary>
<br/>

- Gary Burd's [Redigo](https://github.com/gomodule/redigo)
</details>

<br/>

## Examples & Tests
All unit tests and [examples](examples/examples.go) run via [Travis CI](https://travis-ci.org/mrz1836/go-cache) and uses [Go version 1.15.x](https://golang.org/doc/go1.15). View the [deployment configuration file](.travis.yml).

Run all tests (including integration tests)
```shell script
make test
```

Run tests (excluding integration tests)
```shell script
make test-short
```

Run the [examples](examples/examples.go):
```shell script
make run-examples
```

<br/>

## Benchmarks
Run the Go benchmarks:
```shell script
make bench
```

<br/>

## Code Standards
Read more about this Go project's [code standards](CODE_STANDARDS.md).

<br/>

## Usage
View the [examples](examples/examples.go)

Basic implementation:
```go
package main

import (
    "log"
    "time"
    
    "github.com/gomodule/redigo/redis"
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
	// Output: Got Value: the-value

	// Kill keys by dependency
	keys, _ := cache.KillByDependency("dependent-key-1")
	log.Println("Keys Removed:", keys)
	// Output: Keys Removed: 2
}
```

<br/>

## Maintainers
| [<img src="https://github.com/mrz1836.png" height="50" alt="MrZ" />](https://github.com/mrz1836) | [<img src="https://github.com/kayleg.png" height="50" alt="MrZ" />](https://github.com/kayleg) |
|:---:|:---:|
| [MrZ](https://github.com/mrz1836) | [kayleg](https://github.com/kayleg) |

<br/>

## Contributing
View the [contributing guidelines](CONTRIBUTING.md) and follow the [code of conduct](CODE_OF_CONDUCT.md).

### How can I help?
All kinds of contributions are welcome :raised_hands:! 
The most basic way to show your support is to star :star2: the project, or to raise issues :speech_balloon:. 
You can also support this project by [becoming a sponsor on GitHub](https://github.com/sponsors/mrz1836) :clap: 
or by making a [**bitcoin donation**](https://mrz1818.com/?tab=tips&af=go-cache) to ensure this journey continues indefinitely! :rocket:

<br/>

## License

![License](https://img.shields.io/github/license/mrz1836/go-cache.svg?style=flat)