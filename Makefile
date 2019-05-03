.PHONY: archive

default:
	go build -o dist/ls3 ./*.go

archive: darwin linux windows

darwin:
	GO111MODULE=on GOOS=darwin GOARCH=amd64 go build -o dist/ls3 ./*.go
	cd dist && tar cfz ls3_darwin.tar.gz ls3
	rm ./dist/ls3

linux:
	GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o dist/ls3 ./*.go
	cd dist && tar cfz ls3_linux.tar.gz ls3
	rm ./dist/ls3

windows:
	GO111MODULE=on GOOS=windows GOARCH=amd64 go build -o dist/ls3.exe ./*.go
	cd dist && zip ls3_windows.zip ls3.exe
	rm ./dist/ls3.exe

