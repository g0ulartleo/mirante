# Sentinel Runtime (gRPC)

Mirante worker does not execute sentinels directly.
Instead, it calls a sentinel runner through gRPC using the contract in `proto/sentinelruntime/v1/runtime.proto`.

## Why `config_json` is raw bytes

`config_json` is intentionally opaque to Mirante core.
The worker forwards raw JSON bytes, and each runner parses and validates config in its own runtime.

This design isolates sentinel dependencies from Mirante core and will allow developers to build custom sentinels in languages where we have a working sentinel runtime server.

## Runner responsibilities

- Implement `SentinelRuntime.Check`
- Parse `config_json`
- Resolve `sentinel_type`
- Execute check logic
- Return `status` and `message`
- Return `error.code` and `error.message` for runtime failures

