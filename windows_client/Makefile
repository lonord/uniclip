.PHONY: all build-prod clean

all: build-prod

build-prod:
	go build -ldflags "-s -w" -ldflags -H=windowsgui -o uniclip.exe

build-dev:
	go build -o uniclip.exe

clean:
	rm -rf uniclip.exe