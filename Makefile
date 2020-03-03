build:
	go build ./cmd/puma-dev

clean:
	rm -f ./puma-dev

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
	go tool cover -html=coverage.out -o coverage.html

test-macos-interactive:
	@echo "This will break your existing puma-dev setup. You'll need to run setup/install again. Cool? Cool."
	@echo "Also, prepare to provide your system password several times."
	@read -p "Press [return] to continue..."
	go test ./... -v -test.run=DarwinInteractive -count=1

test-macos-manual-setup-install: clean build
	sudo launchctl unload "$$HOME/Library/LaunchAgents/io.puma.dev.plist"
	rm -rf "$$HOME/Library/ApplicationSupport/io.puma.dev"
	rm -f "$$HOME/Library/LaunchAgents/io.puma.dev.plist"
	rm -f "$$HOME/Library/Logs/puma-dev.log"

	sudo ./puma-dev -d 'test:localhost:loc.al:puma' -setup
	./puma-dev -d 'test:localhost:loc.al:puma' -install

	test -f "$$HOME/Library/LaunchAgents/io.puma.dev.plist"
	launchctl list io.puma.dev > /dev/null
	test -f "$$HOME/Library/Logs/puma-dev.log"
	test 'Hi Puma!' == "$$(curl -s https://rack-hi-puma.puma)" && echo "PASS"

.PHONY: all release
