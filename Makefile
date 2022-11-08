.PHONY: clean start

build: build/bin/landrop

build/bin/landrop: go.mod go.sum *.go *.html
	go build -trimpath -o $@ .

start:
	mkdir -p test
	go run . -dir test -w

clean:
	rm -rf build
