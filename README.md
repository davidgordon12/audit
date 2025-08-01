# Audit

A stylish wrapper around Go's Logger package.

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
  audit := audit.NewAudit(config)

  audit.Info("IM ALIVE!!!!!")
}
```

Output: 
```bash
[2025-06-05 22:05:53] 👋INFO IM ALIVE!!!!!
```

If you wan't to change a few options, you can easily configure them by chaining some more methods to the initialization 

```go
package main

import "github.com/davidgordon12/audit"

func main() {
	audit, err := audit.NewAudit().
		Level(audit.DEBUG).
		DateFormat("[2006-01-02 15:04:05]").
		AddFile("logs.txt")

  if err != nil {
		audit.Error("Couldn't add file output to audit")
	}

  audit.Info("Hello again :]")
}
```

More features to come, like JSON parsing, Tracing, and more.

## TODO:
- [ ] Implement TRACE (and the ability to set the max depth)
- [ ] Implement ERROR alerts to 3rd party logging systems like Grafana and Datadog
- [ ] Enable or disable emojis for legacy terminals and editors
- ~~[ ] Fix dangling threads with wait groups / channels~~
- [x] Rewrite the logging to be synchronous, with an asynchronous background worker to pull messages from the queue.
