all: build

build:
	@mkdir -p cmd
	go build -o _dist/valkey-sentinel-proxy ./cmd

clean:
	@rm -rf _dist

test:
	go test -v ./... -race