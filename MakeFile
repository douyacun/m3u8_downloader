m:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build -ldflags '-extldflags "-static"'  -a -o m3u8_downloader main.go
