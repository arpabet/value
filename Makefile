
all: build

clean:
	go clean -i ./...

vet:
	go vet ./...

test: vet
	go test -race -cover ./...

build: test
	go build ./...

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

update:
	go get -u ./...