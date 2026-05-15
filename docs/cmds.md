```bash
go mod init github.com/prantoran/httpfromtcp
go run .
```

```bash
git init -b main
git add .
git commit -m "ReadCloser interface and channel to read from file"
git remote add origin https://github.com/prantoran/httpfromtcp
git remote -v
git push -u origin main
```

```bash
go run . | tee /tmp/tcp.txt
printf "Are you willing to go all the way?" | nc -D -c -w 1 127.0.0.1 42069
```

```bash
go run ./cmd/tcplistener | tee /tmp/tcplistener.txt
nc -v localhost 42069
```

```bash
go run ./cmd/udpsender | tee /tmp/udpsender.txt
nc -u -l 42069
```

```bash
go run ./cmd/tcplistener | tee /tmp/tcplistener.txt
curl localhost:42069/coffee
curl -X POST -H "Content-Type: application/json" -d '{"flavor": "dark mode"}' http://localhost:42069/coffee
```

```bash
mkdir -p ./internal/request
go test ./...
```