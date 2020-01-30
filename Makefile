all:
	go build ./cmd/puma-dev

install:
	go install ./cmd/puma-dev

lint:
	find . -name '*.go' -not -wholename './vendor/*' -exec golint '{}' \;
	golangci-lint run

release:
	gox -os="darwin linux" -arch="amd64" -ldflags "-X main.Version=$$RELEASE" ./cmd/puma-dev
	mv puma-dev_linux_amd64 puma-dev
	tar czvf puma-dev-$$RELEASE-linux-amd64.tar.gz puma-dev
	mv puma-dev_darwin_amd64 puma-dev
	zip puma-dev-$$RELEASE-darwin-amd64.zip puma-dev

test:
	go test -v ./...

.PHONY: all release
