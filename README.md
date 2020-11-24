# Mujlog

Mujlog (Multiline JSON Log) is a formatter and writer.

Mujlog in pre alpha version is extremely slow and allocates a lots of memory.

## Usage

Set Mujlog as global logger

```go
package main

import (
    "log"

    "github.com/danil/mujlog"
)

func main() {
    l := mujlog.Log{
        Output: os.Stdout,
        Keys: [4]string{"message", "preview"},
        Marks: [3][]byte{[]byte("…")},
        Max: 12,
    }
    log.SetOutput(l)

    log.Println("Hello,\nWorld!")
}
```

Output:

```json
{
    "preview":"Hello, World…",
    "message":"Hello,\nWorld!"
}
```

## Use Mujlog as GLEF formater

```go
package main

import (
    "log"

    "github.com/danil/mujlog"
)

func main() {
    glf := mujlog.GELF()
    glf.Output = os.Stdout
    log.SetOutput(glf)
    log.Println("Hello,\nGELF!")
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

## Caveats

Numeric types appears in the short/full messages as a string. For example:

```go
package main

import (
    "log"

    "github.com/danil/mujlog"
)

func main() {
    l := mujlog.Log{
        Output: os.Stdout,
        Keys: [4]string{"message"},
        Max: 120,
    }
    log.SetOutput(l)

    log.Println(123)
    log.Println(3.21)
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
