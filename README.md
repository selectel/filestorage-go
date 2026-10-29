# filestorage-go: Go SDK for the Selectel File Storage API
[![Go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/selectel/filestorage-go/)
[![Go Report Card](https://goreportcard.com/badge/github.com/selectel/filestorage-go)](https://goreportcard.com/report/github.com/selectel/filestorage-go)
![Build Status](https://github.com/selectel/filestorage-go/actions/workflows/unit-tests.yml/badge.svg)

Package filestorage-go provides Go SDK to work with the Selectel File Storage API.

The `pkg/v1` import path identifies the first version of the SDK's public Go contract. It does
not identify the remote API version; every request carries the API microversion 2.75.

The SDK accepts a regional endpoint, a project-scoped token, a `User-Agent` and an HTTP client
from the caller. It does not perform Keystone authentication, poll for resource state transitions
or retry requests on its own.

## Documentation

The Go library documentation is available at [go.dev](https://pkg.go.dev/github.com/selectel/filestorage-go/).

## What this library is capable of

You can use this library to work with the following objects of the Selectel File Storage API:

* [share](https://pkg.go.dev/github.com/selectel/filestorage-go/pkg/v1/share)
* [share access rule](https://pkg.go.dev/github.com/selectel/filestorage-go/pkg/v1/shareaccess)
* [share network](https://pkg.go.dev/github.com/selectel/filestorage-go/pkg/v1/sharenetwork)
* [share type](https://pkg.go.dev/github.com/selectel/filestorage-go/pkg/v1/sharetype)
* [availability zone](https://pkg.go.dev/github.com/selectel/filestorage-go/pkg/v1/availabilityzone)
* [user message](https://pkg.go.dev/github.com/selectel/filestorage-go/pkg/v1/message)

## Getting started

### Installation

You can install needed `filestorage-go` packages via `go get` command:

```bash
go get github.com/selectel/filestorage-go/pkg/v1/share github.com/selectel/filestorage-go/pkg/v1/shareaccess
```

### Authentication

To work with the Selectel File Storage API you first need to:

* Create a Selectel account: [registration page](https://my.selectel.ru/registration).
* Create a project in Selectel Cloud Platform [projects](https://my.selectel.ru/vpc/projects).
* Retrieve a token for your project via API or [go-selvpcclient](https://github.com/selectel/go-selvpcclient).

### Endpoints

The File Storage API endpoint is region-specific. Retrieve the endpoint of the `sharev2` service
type for the required region from the Identity catalog.

### Usage example

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	v1 "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/selectel/filestorage-go/pkg/v1/share"
	"github.com/selectel/filestorage-go/pkg/v1/shareaccess"
)

func main() {
	ctx := context.Background()

	// Token of the project and endpoint of the File Storage API in the required region.
	token := "gAAAAABeVNzu-..."
	endpoint := "<sharev2-endpoint-from-the-identity-catalog>"

	// Initialize the File Storage API client.
	client, err := v1.NewClient(v1.Config{
		Endpoint:  endpoint,
		Token:     token,
		UserAgent: "my-application/v1.0.0",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create an NFS share in a private network.
	created, _, err := share.Create(ctx, client, share.CreateOpts{
		ShareProto: "NFS",
		Size:       50,
		ShareType:  "basic.ru-9a",
		Name:       "test-share",
		NetworkID:  "<private-network-id>",
		SubnetID:   "<private-subnet-id>",
	})
	if err != nil {
		log.Fatal(err)
	}

	// The SDK does not wait: poll the share until it is available.
	current := created
	for current.Status != "available" {
		if current.Status == "error" {
			log.Fatalf("share %s entered status error", current.ID)
		}
		time.Sleep(10 * time.Second)

		if current, _, err = share.Get(ctx, client, created.ID); err != nil {
			log.Fatal(err)
		}
	}

	// Allow read-write access for a subnet.
	if _, _, err := shareaccess.Allow(ctx, client, created.ID, shareaccess.AllowOpts{
		AccessTo:    "10.0.0.0/24",
		AccessLevel: "rw",
	}); err != nil {
		log.Fatal(err)
	}

	// Print the paths to mount.
	locations, _, err := share.ListExportLocations(ctx, client, created.ID)
	if err != nil {
		log.Fatal(err)
	}

	for _, location := range locations {
		fmt.Println(location.Path)
	}
}
```
