# gelfconv  [![Travis-CI](https://travis-ci.org/m-mizutani/gelfconv.svg)](https://travis-ci.org/m-mizutani/gelfconv) [![Report card](https://goreportcard.com/badge/github.com/m-mizutani/gelfconv)](https://goreportcard.com/report/github.com/m-mizutani/gelfconv) 

`gelfconv` is a converter of [GELF (Graylog Extended Log Format)](http://docs.graylog.org/en/latest/pages/gelf.html). Graylog allows a log provider to send logs as not only HTTP and syslog but also TCP/UDP + GELF. TCP/UDP + GELF is lower overhead than HTTP and can handle structured data. GELF is based on JSON, but it has several additional rules.

- Several fields are mandatry.
- Do not allow nested data structure.
- Restricted field name.

This library converts structured data (`struct` with JSON tag or JSON string) to GELF encoded byte sequence according to the rules.

## Getting Started

### Basic Usage

```go
package main

import (
    "github.com/m-mizutani/gelfconv"
    "fmt"
)

type LogData struct {
    IPAddr  string `json:"ipaddr"`
    Port    int    `json:"port"`
    Request string `json:"request"`
}

func main() {
    log := LogData{"10.1.2.3", 51234, "GET xxx"}

    msg := gelfconv.NewMessage("test message")
    msg.SetData(log)
    rawGELF, err := msg.Gelf()
    if err != nil {
        fmt.Errorf("convert error %v", err)
    }

    fmt.Println(string(rawGELF))
    // Output:
    // {
    //   "_ipaddr": "10.1.2.3",
    //   "_port": 51234,
    //   "_request": "GET xxx",
    //   "host": "your_host_name",
    //   "short_message": "test message",
    //   "timestamp": 1543051370,
    //   "version": "1.1"
    // }
}
```

### Change hostname, time, and optional fields

```go
    msg := gelfconv.NewMessage("test message")
    msg.Host = "myhost"
    msg.ShortMessage = "another test message"
    msg.Timestamp = time.Now().UTC() // default

    // Optional fields
    msg.FullMessage = "it's full message"
    msg.Level = 3
```

### Convert nested data

```go
    data := map[string]interface{}{
        "k1": "v1",
        "k2": map[string]string{
            "k3": "v3",
        },
        "k4": []int{1, 2, 3},
    }
    m := gelfconv.NewMessage("test")
    m.SetData(data)
    rawGELF, _ := m.Gelf()
    fmt.Println(string(rawGELF))
    // {
    //   "_k1": "v1",
    //   "_k2_k3": "v3",  -> nest key is flattened
    //   "_k4": "[1,2,3]",  -> array data is converted to string
    //   ...
    // }
```

## Author

- Masayoshi Mizutani <mizutani@sfc.wide.ad.jp>