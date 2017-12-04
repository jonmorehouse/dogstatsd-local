# Dogstatsd Local
> A local implementation of the dogstatsd protocol from [Datadog](https://www.datadog.com)

## Why?

[Datadog](https://www.datadog.com) is great for production application metric aggregation. This project was inspired by the need to inspect and debug metrics _before_ sending them to `datadog`.

`dogstatsd-local` is a small program which understands the `dogstatsd` and `statsd` protocols. It listens on a local UDP server and writes metrics, events and service checks per the [dogstatsd protocol](https://docs.datadoghq.com/guides/dogstatsd/) to `stdout` in user configurable formats.

This can be helpful for _debugging_ metrics themselves, and to prevent polluting datadog with noisy metrics from a development environment. **dogstatsd-local** can also be used to pipe metrics as json to other processes for further processing.

## Usage

### Build Manually

This is a go application with no external dependencies. Building should be as simple as running `go build` in the source directory.

Once compiled, the `dogstatsd-local` binary can be run directly:
```bash
$ ./dogstatsd-local -port=8126
```

### Prebuilt Binaries

**Coming soon**

### Docker

```bash
$ docker run -p 8126:8126 jonmorehouse/dogstatsd-local
```

## Sample Formats

### Raw (no formatting)

When writing a metric such as:

```bash
$ printf "namespace.metric:1|c|#test" | nc -cu  localhost 8125
```

Running **dogstatsd-local** with the `-format raw` flag will output the plain udp packet:

```bash
$ docker run -p 8125 jonmorehouse/dogstatsd-local -format raw
2017/12/03 23:11:31 namespace.metric.name:1|c|@1.00|#tag1
```

### Human

When writing a metric such as:

```bash
$ printf "namespace.metric:1|c|#test" | nc -cu  localhost 8125
```

Running **dogstatsd-local** with the `-format human` flag will output a human readable metric:

```bash
$ docker run -p 8125 jonmorehouse/dogstatsd-local -format human
metric:counter|namespace.metric|1.00  test

```

### JSON

When writing a metric such as:
```bash
$ printf "namespace.metric:1|c|#test|extra" | nc -cu  localhost 8125
```

Running **dogstatsd-local** with the `-format json` flag will output json:

```bash
$ docker run -p 8125 jonmorehouse/dogstatsd-local -format json | jq .
{"namespace":"namespace","name":"metric","path":"namespace.metric","value":1,"extras":["extra"],"sample_rate":1,"tags":["test"]}
```

**dogstatsd-local** can be piped to any process that understands json via stdin. For example, to pretty print JSON with [jq](https://stedolan.github.io/jq/):

```bash
$ docker run -p 8125 jonmorehouse/dogstatsd-local -format json | jq .
{
  "namespace": "namespace",
  "name": "metric",
  "path": "namespace.metric",
  "value": 1,
  "extras": [
    "extra"
  ],
  "sample_rate": 1,
  "tags": [
    "test"
  ]
}
```

## TODO

- [ ] support datadog service checks
- [ ] support datadog events
- [ ] support interval aggregation of percentiles
