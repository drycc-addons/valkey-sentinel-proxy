all: build

build:
	@mkdir -p cmd
	go build -o _dist/redis-sentinel-proxy ./cmd

clean:
	@rm -rf _dist

test:
	go test -v ./... -race