# Logastic

[![Build Status](https://cloud.drone.io/api/badges/danil/logastic/status.svg)](https://cloud.drone.io/danil/logastic)
[![Go Reference](https://pkg.go.dev/badge/github.com/danil/logastic.svg)](https://pkg.go.dev/github.com/danil/logastic)

JSON logging for Go.

## About

The software is considered to be at a alpha level of readiness -
its extremely slow and allocates a lots of memory)

## Install

    go get github.com/danil/logastic@v0.77.0

## Usage

Set Logastic as global logger

```go
package main

import (
    "os"
    "log"

    "github.com/danil/logastic"
)

func main() {
    l := logastic.Log{
        Output: os.Stdout,
        Trunc: 12,
        Keys: [4]json.Marshaler{logastic.String("message"), logastic.String("excerpt")},
        Marks: [3][]byte{[]byte("…")},
        Replace: [][]byte{[]byte("\n"), []byte(" ")},
    }
    log.SetFlags(0)
    log.SetOutput(l)

    log.Print("Hello,\nWorld!")
}
```

Output:

```json
{
    "message":"Hello,\nWorld!",
    "excerpt":"Hello, World…"
}
```

## Use as GELF formater

```go
package main

import (
    "log"
    "os"

    "github.com/danil/logastic"
)

func main() {
    l := logastic.GELF()
    l.Output = os.Stdout
    log.SetFlags(0)
    log.SetOutput(l)
    log.Print("Hello,\nGELF!")
}
```

Output:

```json
{
    "version":"1.1",
    "short_message":"Hello, GELF!",
    "full_message":"Hello,\nGELF!",
    "timestamp":1602785340
}
```

## Caveat: numeric types appears in the message as a string

```go
package main

import (
    "log"
    "os"

    "github.com/danil/logastic"
)

func main() {
    l := logastic.Log{
        Output: os.Stdout,
        Keys: [4]json.Marshaler{logastic.String("message")},
    }
    log.SetFlags(0)
    log.SetOutput(l)

    log.Print(123)
    log.Print(3.21)
}
```

Output 1:

```json
{
    "message":"123"
}
```

Output 2:

```json
{
    "message":"3.21"
}
```

## Benchmark

```
goos: linux
goarch: amd64
pkg: github.com/danil/logastic
BenchmarkLogastic/io.Writer_36-8         	  276825	      4120 ns/op
BenchmarkLogastic/fmt.Fprint_io.Writer_1006-8         	  121680	      9697 ns/op
PASS
ok  	github.com/danil/logastic	2.476s
PASS
ok  	github.com/danil/logastic/encode	0.002s
```

## License

Copyright (C) 2020 [Danil Kutkevich](https://github.com/danil)  
See the [LICENSE](./LICENSE) file for license rights and limitations (MIT)
