

[![build workflow](https://github.com/antonydenyer/block-builder-mempool/actions/workflows/build.yml/badge.svg)](https://github.com/antonydenyer/block-builder-mempool/actions)


## built using bun starter kit

https://github.com/go-bun/bun-starter-kit


## Quickstart

To start using this kit, clone the repo:

```shell
git clone https://github.com/antonydenyer/block-builder-mempool.git
```


Make sure you have correct information in `app/embed/config/dev.yaml` and then run migrations (database
must exist before running):

```shell
go run cmd/bun/main.go -env=dev db init
go run cmd/bun/main.go -env=dev db migrate
```

To start the server:

```shell
go run cmd/bun/main.go -env=dev server
```

Then run the tests in [example](example) package:

```shell
cd example
go test
```

See [documentation](https://bun.uptrace.dev/guide/starter-kit.html) for more info.
