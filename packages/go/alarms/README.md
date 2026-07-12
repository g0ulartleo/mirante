# Mirante Go Alarms SDK

Go SDK for building Mirante alarm runtimes.

## Versioning

This package is a nested Go module and has independent versions from the root Mirante app and the NPM SDK.

Release tags must include the module path prefix:

```sh
git tag packages/go/alarms/v0.1.0
git push origin packages/go/alarms/v0.1.0
```

Generated runtime repositories should require the latest published SDK version:

```go
require github.com/g0ulartleo/mirante/packages/go/alarms v0.1.0
```

## Local Development

Before the SDK version is tagged, test generated runtimes with a local replace:

```sh
go mod edit -replace github.com/g0ulartleo/mirante/packages/go/alarms=../mirante/packages/go/alarms
go mod tidy
```

Do not commit local replaces in generated runtime repositories unless that repository intentionally depends on a local SDK checkout.

## Proto Generation

`proto/alarms/v1/alarms.proto` in the root repository is the schema source of truth.

Regenerate SDK stubs from the root repository:

```sh
make proto-generate-go-sdk
```
