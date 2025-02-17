
# Pocketbase Experiments extended with Go

## Install

```bash
go mod tidy
```

## Run Dev

```bash
go run . serve
```

## Build to publish

```bash
$env:GOOS = "linux"
go generate ./...
go build -ldflags "-s -w"

./pocketbase-experiments serve
```

## Module creation
```bash
go mod init github.com/shujink0/pocketbase-experiments
```

## Update All Go Modules
```bash
go get -u -t ./...
go mod tidy
```
