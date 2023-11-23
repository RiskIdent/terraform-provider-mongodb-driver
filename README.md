<!--
SPDX-FileCopyrightText: 2023 Risk.Ident GmbH <contact@riskident.com>

SPDX-License-Identifier: CC-BY-4.0
-->

# Terraform Provider MongoDB driver

[![REUSE status](https://api.reuse.software/badge/github.com/RiskIdent/terraform-provider-mongodb-driver)](https://api.reuse.software/info/github.com/RiskIdent/terraform-provider-mongodb-driver)

Manage MongoDB itself using MongoDB driver.

## Using the provider

TODO

## Development

If you wish to work on the provider, you'll first need
[Go](http://www.golang.org) installed on your machine
(see [Development requirements](#development-requirements) below).

To compile the provider, run `go install`.
This will build the provider and put the provider binary in the `$GOPATH/bin`
directory (default: `~/go/bin`).

To generate or update documentation, run `go generate`.

To run the tests, run `make test`.

### Development requirements

- [Tofu](https://opentofu.org/docs/intro/install) >= 1.6.0-alpha1
- [Go](https://golang.org/doc/install) >= 1.21

### Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command:

```shell
go install
```

## License

This repository complies with the [REUSE recommendations](https://reuse.software/).

Different licenses are used for different files. In general:

- Go code is licensed under Mozilla Public License v2.0 ([LICENSES/MPL-2.0.txt](LICENSES/MPL-2.0.txt)).
- Documentation licensed under Creative Commons Attribution 4.0 International ([LICENSES/CC-BY-4.0.txt](LICENSES/CC-BY-4.0.txt)).
- Miscellaneous files, e.g `.gitignore`, are licensed under CC0 1.0 Universal ([LICENSES/CC0-1.0.txt](LICENSES/CC0-1.0.txt)).

Please see each file's header or accompanied `.license` file for specifics.
The generated documentation found in the `docs` directory have their licenses
marked by the [.reuse/dep5](.reuse/dep5) file.
