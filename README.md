# ansible-go


## build

```
CGO_ENABLED=0  GOOS=linux  GOARCH=amd64  go build main.go
```

## use
```
go run main.go -cmd 'ls /'
```