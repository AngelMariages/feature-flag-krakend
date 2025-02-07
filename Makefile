# Command to run: make
# main build file for the project
# Output: krakend-debugger.so
# Using golang

build:
	docker run --platform linux/amd64 -e CGO_ENABLED=1 -e GOARCH=amd64 -e ARCH=amd64 -it -v "$(shell pwd)/amplitude_forwarder:/app" -w /app krakend/builder:2.9.1 go build -buildmode=plugin -o amplitude_forwarder.so .

start:
	docker run -p "8080:8080" -v $(shell pwd):/etc/krakend/ devopsfaith/krakend:watch run -c krakend.json

stop:
	docker stop -t 0 $(shell docker ps -q --filter ancestor=devopsfaith/krakend:watch)