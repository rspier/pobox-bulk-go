# Pobox Bulk API Client in Go

This is a small client for the Pobox Bulk Forwarding configuration API, written
in Go.

[![GoDoc](https://godoc.org/github.com/fastmail/pobox-bulk-go?status.svg)](https://godoc.org/github.com/fastmail/pobox-bulk-go)

## Examples

Please see the Tools below for examples for how to use the API.

## Tools

Use `--help` to see all flags.

Authentication goes in a yaml file (`$PWD/.pobox-api-auth`, but configurable
with the `--authfile` flag).  It looks like:

```
user: "GUI-GOES-HERE"
pass: "APIKEY"
```

### sync

```
go run cmd/sync/sync.go  --domain=mydomain.com --mapfile=routes.csv
```

example routes.csv:

```csv
foo,foo@somewhere.com
bar,boo@elsewhere.com
```

NOTE: the left hand side does _not_ contain the domain.

### dump

```
go run cmd/dump/dump.go
```

The `-d` flag changes the output delimiter.