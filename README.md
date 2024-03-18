# devnet-explorer

Gevulot Devnet Explorer

## Running

Build devnet-explorer by running:

```sh
mage go:build
```

See available configuration parameters:

```sh
./target/bin/devnet-explorer -h
```

Set configuration parameters to match your environment and run devnet-explorer:

```sh
LOG_LEVEL="debug" STATS_TTL="1s" ./target/bin/devnet-explorer
```

### With Postgres

To use devnet-explorer against real database pass postgres DNS via DNS env variable.
Run `mage go:run` and open UI at [http://127.0.0.1:8383](http://127.0.0.1:8383).

### With mock data

Devnet explorer can be executed without DB using mock data.
Run `mage go:runWithMockDB` and open UI at [http://127.0.0.1:8383](http://127.0.0.1:8383).

## Development

### Requirements

- [Go](https://go.dev/) >= 1.22
- [Mage](https://magefile.org/)
- Docker/Podman

### Get started

Clone repository and execute `mage` to get started.
