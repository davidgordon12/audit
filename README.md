# Audit

A stylish wrapper around Go's Logger package.

Audit is just a single Go file, that is short and sweet. Easy to read, modify and extend. Getting started is simple, first, download the dependency.

```
go get github.com/davidgordon12/audit
```

Then, just call NewAudit!

```go
package main

import "github.com/davidgordon12/audit"

func main() {
  audit := audit.NewAudit()

  audit.Info("I'M ALIVE!!!!!")
}
```

Output: 

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

Output:

More features to come, like JSON parsing, Tracing, and more.

## TODO:
- [ ] Implement TRACE (and the ability to set the max depth)
- [ ] Implement ERROR alerts to 3rd party logging systems like Grafana and Datadog
- [ ] Maybe enable or disable emojis for legacy terminals and editors (but who else is really using this)
