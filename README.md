# rkt-sidekick
[![Go Report Card](https://goreportcard.com/badge/github.com/monder/rkt-sidekick)](https://goreportcard.com/report/github.com/monder/rkt-sidekick)
[![license](https://img.shields.io/github/license/monder/rkt-sidekick.svg?maxAge=2592000&style=flat-square)]()
[![GitHub tag](https://img.shields.io/github/tag/monder/rkt-sidekick.svg?style=flat-square)]()

An `ACI` that writes container ip address to [etcd].
That provides the more [rkt]-like approach to [sidekick] model implementation which comes in handy when using multihost network solutions such as [flannel]

## Images
Signed `ACI`s for `linux-arm64` are available at `monder.cc/rkt-sidekick` with versions matching git tags.

## Usage

```
rkt run \
   --insecure-options=image \
   --net=flannel \
   docker://redis \
   monder.cc/rkt-sidekick:v0.1.0 -- -cidr 10.0.0.0/16 -format '{"host":"$ip", "port":3000}' /services/redis-a6f43b/ip
```

The script above will launch redis and a sidekick in the same pod. The sidekick will enumerate all network interfaces and write the first one matching `10.0.0.0/16` to the formatted string to `/services/redis-a6f43b/ip` 

Please note how to pass arguments to multiple images: https://coreos.com/rkt/docs/latest/subcommands/run.html#passing-arguments

### Other options

```
Usage: rkt-sidekick etcd [options] [KEY_IN_ETCD]

Options:
    -cidr              cidr to match the ip (default: "0.0.0.0/0")
    -etcd-endpoint     an etcd address in the cluster (default: "http://172.16.28.1:2379")
    -format            format of the etcd key value. '$ip' will be replace by container's ip address (default: "$ip")
    -expire-dir        set expiration TTLs for all items under that directory, not only the leaf node
    -interval          refresh interval (default: "1m")
```

## License
MIT

[rkt]: https://github.com/coreos/rkt
[etcd]: https://github.com/coreos/etcd
[flannel]: https://github.com/coreos/flannel
[sidekick]: https://coreos.com/fleet/docs/latest/examples/service-discovery.html#sidekick-model
