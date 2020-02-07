all:
	go build ./cmd/puma-dev

install:
	go install ./cmd/puma-dev

release:
	rm -rf ./pkg
	mkdir -p ./pkg

	gox -os="darwin linux" -arch="amd64" -ldflags "-X main.Version=$$RELEASE" ./cmd/puma-dev

	mv puma-dev_linux_amd64 puma-dev
	tar czvf pkg/puma-dev-$$RELEASE-linux-amd64.tar.gz puma-dev

	mv puma-dev_darwin_amd64 puma-dev
	zip pkg/puma-dev-$$RELEASE-darwin-amd64.zip puma-dev

test:
	go test -v ./...

.PHONY: all release
