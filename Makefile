.PHONY: archive

default:
	go build -o dist/ls3 ./*.go

archive: darwin linux windows

darwin:
	GOOS=darwin GOARCH=amd64 go build -o dist/ls3 ./*.go
	cd dist && tar cfz ls3_darwin.tar.gz ls3
	rm ./dist/ls3

linux:
	GOOS=linux GOARCH=amd64 go build -o dist/ls3 ./*.go
	cd dist && tar cfz ls3_linux.tar.gz ls3
	rm ./dist/ls3

windows:
	GOOS=darwin GOARCH=amd64 go build -o dist/ls3.exe ./*.go
	cd dist && zip ls3_windows.zip ls3.exe
	rm ./dist/ls3.exe

