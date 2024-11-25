# Redis Proxy

Redis Proxy is a lightweight, fast but powerful Redis Sentinel Proxy written in Go.

## Usage

### Build

```bash
make build
```

### Command

```bash
# ./bin/redis-sentinel-proxy --help
Usage of redis-sentinel-proxy:
  -listen string
        listen address (default ":9999")
  -master string
        name of the master redis node
  -max-procs int
        sets the maximum number of CPUs that can be executing (default 1)
  -sentinel-addr string
        remote sentinel address (default ":26379")
  -sentinel-pass string
        password to use when connecting to the sentinel server.
  -sentinel-user string
        username to use when connecting to the sentinel server.
```

## Architecture

Each client connection is wrapped with a session, which spawns two goroutines to read request from and write response to the client. Each session appends it's request to dispatcher's request queue, then dispatcher route request to the right task runner according key hash and slot table. Task runner sends requests to its backend server and read responses from it.
Upon cluster topology changed, backend server will response MOVED or ASK error. These error is handled by session, by sending request to destination server directly. Session will trigger dispatcher to update slot info on MOVED error. When connection error is returned by task runner, session will trigger dispather to reload topology.

## Performance

Redis includes the redis-benchmark utility that simulates running commands done by N clients at the same time sending M total queries (it is similar to the Apache's ab utility). Below you'll find the full output of a benchmark executed against a Linux box.

The following options are supported:

```bash
Usage: redis-benchmark [-h <host>] [-p <port>] [-c <clients>] [-n <requests]> [-k <boolean>]

 -h <hostname>      Server hostname (default 127.0.0.1)
 -p <port>          Server port (default 6379)
 -s <socket>        Server socket (overrides host and port)
 -a <password>      Password for Redis Auth
 --user <username>  Used to send ACL style 'AUTH username pass'. Needs -a.
 -c <clients>       Number of parallel connections (default 50)
 -n <requests>      Total number of requests (default 100000)
 -d <size>          Data size of SET/GET value in bytes (default 3)
 --dbnum <db>       SELECT the specified db number (default 0)
 --threads <num>    Enable multi-thread mode.
 --cluster          Enable cluster mode.
 --enable-tracking  Send CLIENT TRACKING on before starting benchmark.
 -k <boolean>       1=keep alive 0=reconnect (default 1)
 -r <keyspacelen>   Use random keys for SET/GET/INCR, random values for
                    SADD, random members and scores for ZADD.
                    Using this option the benchmark will expand the string
                    __rand_int__ inside an argument with a 12 digits number
                    in the specified range from 0 to keyspacelen-1. The
                    substitution changes every time a command is executed.
                    Default tests use this to hit random keys in the
                    specified range.
 -P <numreq>        Pipeline <numreq> requests. Default 1 (no pipeline).
 -e                 If server replies with errors, show them on stdout.
                    (No more than 1 error per second is displayed.)
 -q                 Quiet. Just show query/sec values
 --precision        Number of decimal places to display in latency output (default 0)
 --csv              Output in CSV format
 -l                 Loop. Run the tests forever
 -t <tests>         Only run the comma separated list of tests. The test
                    names are the same as the ones produced as output.
 -I                 Idle mode. Just open N idle connections and wait.
 --help             Output this help and exit.
 --version          Output version and exit.
```

You need to have a running Redis instance before launching the benchmark. A typical example would be:

```bash
redis-benchmark -p 8088 -c 500 -n 5000000 -P 100 -r 10000 -t get,set
```

Using this tool is quite easy, and you can also write your own benchmark, but as with any benchmarking activity, there are some pitfalls to avoid.

## Development

The Drycc project welcomes contributions from all developers. The high-level process for development matches many other open source projects. See below for an outline.

* Fork this repository
* Make your changes
* [Submit a pull request][prs] (PR) to this repository with your changes, and unit tests whenever possible.
  * If your PR fixes any [issues][issues], make sure you write Fixes #1234 in your PR description (where #1234 is the number of the issue you're closing)
* Drycc project maintainers will review your code.
* After two maintainers approve it, they will merge your PR.

[prs]: https://github.com/drycc-addons/redis-sentinel-proxy/pulls
[issues]: https://github.com/drycc-addons/redis-sentinel-proxy/issues