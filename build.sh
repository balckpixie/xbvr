GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-H windowsgui" -tags=json1 -o xbvr.exe pkg/tray/main.go
