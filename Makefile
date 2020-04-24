build:
	go build ./cmd/puma-dev

install:
	go install ./cmd/puma-dev

lint:
	golangci-lint run

release:
	rm -rf ./pkg
	mkdir -p ./pkg

	gox -os="darwin linux" -arch="amd64" -ldflags "-X main.Version=$$RELEASE" ./cmd/puma-dev

	mv puma-dev_linux_amd64 puma-dev
	tar czvf pkg/puma-dev-$$RELEASE-linux-amd64.tar.gz puma-dev

	mv puma-dev_darwin_amd64 puma-dev
	zip pkg/puma-dev-$$RELEASE-darwin-amd64.zip puma-dev

test:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

coverage: test
	go tool cover -html=coverage.out

test-macos-interactive-dev-setup-install: build
	sudo launchctl unload "$$HOME/Library/LaunchAgents/io.puma.dev.plist"
	rm "$$HOME/Library/LaunchAgents/io.puma.dev.plist"
	rm "$$HOME/Library/Logs/puma-dev.log"
	sudo ./puma-dev -d 'test:localhost:loc.al:puma' -setup
	./puma-dev -d 'test:localhost:loc.al:puma' -install
	test -f "$$HOME/Library/LaunchAgents/io.puma.dev.plist"
	sleep 2
	test -f "$$HOME/Library/Logs/puma-dev.log"

test-macos-interactive-certificate-install:
	go test -coverprofile=coverage_osx.out -v -test.run=TestSetupOurCert_InteractiveCertificateInstall ./dev

.PHONY: all release
