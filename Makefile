test:
	go test ./...

build:
	go build -o "bin/maven-runner" ./cmd/maven-runner/...