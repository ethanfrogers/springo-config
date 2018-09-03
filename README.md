# springo-config

A Go library for evaluating Spring YAML configuration, primarily developed for evaluating [Spinnaker]() configuration in Go projects.

## Installation

```bash
$ go get github.com/ethanfrogers/springo-config
```

## Usage

The below example assumes you have valid Spinnaker config located at `~/.spinnaker/`. For examples of how other things would be evaluates, see `pkg/parser_test.go`.

```go

import (
    "fmt"
    "github.com/ethanfrogers/springo-config/pkg"
)

func main() {
    cfg := pkg.NewConfig().
        WithApplications("spinnaker", "clouddriver").
        WithProfiles("local")
    
    if err := cfg.Load(pkg.WithEnvironmentVariables()); err != nil {
        panic(err)
    }

    fmt.Println(cfg.Get("services.clouddriver.port"))
}

```
