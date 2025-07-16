
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
go generate ./...
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w"
./pocketbase-experiments serve
```

## Copy to Server
```bash
scp pocketbase-experiments root@208.117.87.150:/var/www/pb/
```

## Module creation
```bash
go mod init github.com/dorianlgs/pocketbase-experiments
```

## Update All Go Modules
```bash
go get -u -t ./...
go mod tidy
```
