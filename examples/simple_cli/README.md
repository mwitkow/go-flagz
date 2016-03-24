# Simple CLI demo

This demonstrates how dynamic values are being updated.

## Quick set up:

Download [etcd](https://github.com/coreos/etcd/releases), extract and make it available on your `$PATH`.

Launch `etcd` server serving from a `default.data` in `/tmp`:

```sh
cd /tmp
etcd 
```

Set up a set of flags:

```sh
etcdctl mkdir /example/flagz
etcdctl set /example/flagz/staticint 9090
etcdctl set /example/flagz/dynstring foo
```

Play around:

```sh
./simple_cli &
etcdctl set /example/flagz/dynstring bar
etcdctl set /example/flagz/dynint 7777
etcdctl set /example/flagz/dynstring bad_value
```

Profit.