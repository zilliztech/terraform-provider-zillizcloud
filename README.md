# Terraform Zilliz Cloud Provider

This is the repository for the terraform-provider-zillizcloud, which allows one to use Terraform with Zilliz Cloud. Learn more about [Zilliz Cloud](https://zilliz.com/cloud)

For general information about Terraform, visit the [official website](https://www.terraform.io) and the [GitHub project page](https://github.com/hashicorp/terraform).

## Table of Contents
<!-- toc -->

- [User Guide](#user-guide)
- [API Documentation](#api-documentation)
- [Requirements](#requirements)
- [Building The Provider](#building-the-provider)

<!-- tocstop -->


## User Guide

See [Zilliz Cloud Terraform Integration Overview](./docs/README.md) for more information.

## API Documentation

API Documentation can be found on the [Terraform Registry](https://registry.terraform.io/providers/zilliztech/zillizcloud/latest/docs).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

Fill this in for each provider

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```

