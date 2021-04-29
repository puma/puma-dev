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

	SDKROOT=$$(xcrun --sdk macosx --show-sdk-path) gox -cgo -os="darwin" -arch="amd64 arm64" -ldflags "-X main.Version=$$RELEASE" ./cmd/puma-dev
	gox -os="linux" -arch="amd64 arm64" -ldflags "-X main.Version=$$RELEASE" ./cmd/puma-dev

	for arch in amd64 arm64; do \
		mv -v "puma-dev_linux_$$arch" puma-dev; \
		tar czvf "pkg/puma-dev-$$RELEASE-linux-$$arch.tar.gz" puma-dev; \
	done

	for arch in amd64 arm64; do \
		mv -v "puma-dev_darwin_$$arch" puma-dev; \
		zip -v "pkg/puma-dev-$$RELEASE-darwin-$$arch.zip" puma-dev; \
	done

test:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

clean-test:
	rm -rf ~/.gotest-macos-puma-dev

test-macos-filesystem-setup:
	sudo mkdir -p /etc/resolver;
	sudo chmod 0775 /etc/resolver;
	sudo chown :staff /etc/resolver;

coverage: test
	go tool cover -html=coverage.out -o coverage.html

test-macos-interactive:
	@echo "This will break your existing puma-dev setup. You'll need to run setup/install again. Cool? Cool."
	@echo "Also, prepare to provide your system password several times."
	@read -p "Press [return] to continue..."
	rm -rf "$$HOME/Library/Application\ Support/io.puma.dev"
	go test ./... -v -test.run=DarwinInteractive -count=1
	rm -rf "$$HOME/Library/Application\ Support/io.puma.dev"

test-macos-manual-setup-install: clean build
	sudo launchctl unload "$$HOME/Library/LaunchAgents/io.puma.dev.plist"
	rm -rf "$$HOME/Library/Application\ Support/io.puma.dev"
	rm -f "$$HOME/Library/LaunchAgents/io.puma.dev.plist"
	rm -f "$$HOME/Library/Logs/puma-dev.log"

	sudo ./puma-dev -d 'test:localhost:loc.al:puma' -setup
	./puma-dev -d 'test:localhost:loc.al:puma' -install

	test -f "$$HOME/Library/LaunchAgents/io.puma.dev.plist"
	launchctl list io.puma.dev > /dev/null
	test -f "$$HOME/Library/Logs/puma-dev.log"
	test 'Hi Puma!' == "$$(curl -s https://rack-hi-puma.puma)" && echo "PASS"

.PHONY: release
