# Audit

A stylish logging package.

Audit is just a few Go files, that are short and sweet, easy to read, modify and extend. Getting started is simple, first, download the dependency.

```
go get github.com/davidgordon12/audit
```

Create an AuditConfig object (Let's leave it empty for now)

```go
config := AuditConfig{}
```

Now you can create a NewAudit()

```go
package main

import "github.com/davidgordon12/audit"

func main() {
  config := AuditConfig{}

  audit := audit.NewAudit(config)

  audit.Info("IM ALIVE!!!!!")
}
```

Output: 
```bash
[2025-06-05 22:05:53] 👋INFO IM ALIVE!!!!!
```

Easily configure the auditer with the AuditConfig object

```go
package main

import "github.com/davidgordon12/audit"

func main() {
  config := AuditConfig {
    FlushInterval: 100 * time.Millisecond,
    BatchSize:     128,
    FilePath:      "resources/logs",
    FileSize:      1024 * 1024 * 1024,
    Level:         DEBUG
  }

  audit := audit.NewAudit(config)

  audit.Info("IM ALIVE!!!!!")
}
```

## TODO:
- [ ] Implement TRACE (and the ability to set the max depth)
- [ ] Implement ERROR alerts to 3rd party logging systems like Grafana and Datadog
- [ ] Enable or disable emojis for legacy terminals and editors
- [ ] Implement ability to pass arguments (primitives and objects) to log call
- [x] File rotation
- ~~[ ] Fix dangling threads with wait groups / channels~~
- [x] Rewrite the logging to be synchronous, with an asynchronous background worker to pull messages from the queue.
