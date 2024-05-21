go-build:
	go build -ldflags -H=windowsgui -o ./release/yd_backup.exe cmd\main\main.go
	cp ./example/config/config.json ./release/config.json
