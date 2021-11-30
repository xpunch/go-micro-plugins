A repository for go-micro kafka plugin

# Overview
Micro tooling is built on a powerful pluggable architecture. Plugins can be swapped out with zero code changes.
This repository contains plugins only for kafka broker, try to avoid too much unused dependencies of micro/go-plugins.

## Getting Started

* [Contents](#contents)
* [Usage](#usage)
* [Build Pattern](#build)

## Contents

Contents of this repository:

| Directory | Description                                                     |
| --------- | ----------------------------------------------------------------|
| Broker    | PubSub messaging; kafka                                         |

## Usage

Plugins can be added to go-micro in the following ways. By doing so they'll be available to set via command line args or environment variables.

Import the plugins in a Go program then call service.Init to parse the command line and environment variables.

```go
import (
	"github.com/asim/go-micro/v3"
	_ "github.com/x-punch/micro-kafka/v3"
	_ "github.com/micro/go-plugins/registry/kubernetes"
	_ "github.com/micro/go-plugins/transport/nats"
)

func main() {
	service := micro.NewService(
		// Set service name
		micro.Name("my.service"),
	)

	// Parse CLI flags
	service.Init()
}
```

### Flags

Specify the plugins as flags

```shell
go run service.go --broker=kafka --registry=kubernetes --transport=nats
```

### Env

Use env vars to specify the plugins

```
MICRO_BROKER=kafka \
MICRO_REGISTRY=kubernetes \ 
MICRO_TRANSPORT=nats \ 
go run service.go
```

### Options

Import and set as options when creating a new service

```go
import (
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/plugins/registry/kubernetes/v3"
)

func main() {
	registry := kubernetes.NewRegistry() //a default to using env vars for master API

	service := micro.NewService(
		// Set service name
		micro.Name("my.service"),
		// Set service registry
		micro.Registry(registry),
	)
}
```

## Build

An anti-pattern is modifying the `main.go` file to include plugins. Best practice recommendation is to include
plugins in a separate file and rebuild with it included. This allows for automation of building plugins and
clean separation of concerns.

Create file plugins.go

```go
package main

import (
	_ "github.com/x-punch/micro-kafka/v3"
	_ "github.com/micro/go-plugins/registry/kubernetes"
	_ "github.com/micro/go-plugins/transport/nats"
)
```

Build with plugins.go

```shell
go build -o service main.go plugins.go
```

Run with plugins

```shell
MICRO_BROKER=kafka \
MICRO_REGISTRY=kubernetes \
MICRO_TRANSPORT=nats \
service
```
