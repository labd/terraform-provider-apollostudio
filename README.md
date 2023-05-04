# Apollo Studio Terraform Provider

Terraform provider for [Apollo Studio](https://www.apollographql.com/).

The intention of this provider is to cover the [Apollo Studio API](https://studio.apollographql.com/public/apollo-platform/variant/main/explorer), so that one can manage sub graph schemas through Terraform.

A single provider can manage the resources of one graph variant schema.

## Currently supported resources

- [x] Federation sub graph schemas
- [x] Federation sub graph schema validations

# Installation

## Terraform registry

Terraform 0.13 added support for automatically downloading providers from
the terraform registry. Simply add the following to your terraform project to use the latest release

```hcl
terraform {
  required_providers {
    apollostudio = {
      source = "labd/apollostudio"
    }
  }
}
```

## Binaries

Packages of the releases are available at
https://github.com/labd/terraform-provider-apollostudio/releases See the
[terraform documentation](https://www.terraform.io/docs/configuration/providers.html#third-party-plugins)
for more information about installing third-party providers.

# Getting started

[Read our documentation](https://registry.terraform.io/providers/labd/apollostudio/latest/docs).

# Contributing

## Building the provider

Clone repository to: `$GOPATH/src/github.com/labd/terraform-provider-apollostudio`

Then run

```sh
make build
```

### Update apollostudio-go-sdk

The apollostudio-go-sdk always uses the latest (master) version. To update to
the latest version:

```sh
make update-sdk
```

## Generating Documentation

This provider uses the `tfplugindocs` tool to automatically generate documentation based on the descriptions of the
resources and fields. Install the most recent release of the tool by downloading a binary from [the tfplugindocs repository](https://github.com/hashicorp/terraform-plugin-docs/releases).
And running

```shell
make docs
```

In order to ensure up to date documentation make sure to update the description fields of any and all resources upon
creation or editing.

### Testing local changes

As of terraform 0.13 testing local changes requires a little effort.
You can run

```sh
make build-local
```

To build the provider with a very high version number and copy it to your terraform plugins folder (default is for Mac,
change OS_ARCH if running Linux or change path if running Windows)
If you set your provider source as `labd/apollostudio` and the version to your built `version` it should use the local
provider. See also [the Terraform 0.13 upgrade guide](https://www.terraform.io/upgrade-guides/0-13.html#new-filesystem-layout-for-local-copies-of-providers)

## Debugging / Troubleshooting

There is currently one environment settings for troubleshooting:

- `TF_LOG=1` enables debug output for Terraform.

Note this generates a lot of output!

## Testing

### Running the unit tests

```sh
$ make test
```

### Running an Acceptance Test

In order to run the full suite of Acceptance tests, run `make testacc`.

**NOTE:** Acceptance tests create real resources.

Prior to running the tests provider configuration details such as access keys
must be made available as environment variables.

Since we need to be able to create Apollo resources, we need the
graph API key and graph ref. So in order for the acceptance tests to run
correctly please provide all of the following:

```sh
export APOLLO_API_KEY=...
export APOLLO_GRAPH_REF=...
```

```sh
$ make testacc
```

## Releasing

When pushing a new tag prefixed with `v` a GitHub action will automatically
use Goreleaser to build and release the build.

```sh
git tag <release> -m "Release <release>" # please use semantic version, so always vX.Y.Z
git push --follow-tags
```

## TODO List

- Create dedicated apollo graph and key for acceptance testing
- Unit/acceptance tests should be expanded
- Add support for other resources

## Authors

This project is developed by [Lab Digital](https://www.labdigital.nl). We
welcome additional contributors. Please see our
[GitHub repository](https://github.com/labd/terraform-provider-apollstudio) for
for more information.
