## built using bun starter kit

https://github.com/go-bun/bun-starter-kit


## Quickstart

To start using this, clone the repo:

```shell
git clone https://github.com/antonydenyer/block-builder-mempool.git  
```

Copy the env file and change your params
```shell
cp example.env .env
```
Update the `RPC_CLIENT_URL` with your node info.

Then just run

```shell
docker-compose up --build
```

Visit http://localhost:8000/


Not that the data is not 100% accurate, it is merely a tool to use to help identify blocks that rquire further investigation.