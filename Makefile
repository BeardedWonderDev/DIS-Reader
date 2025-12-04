VERSION ?= $(shell git describe --tags --always --dirty)
GOOS_LIST = linux darwin windows
GOARCH_LIST = amd64 arm64
DIST := dist
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: release-agent clean test build-binaries package-tar checksum package-linux

release-agent: clean test build-binaries package-tar checksum

test:
	go test ./...

build-binaries:
	@mkdir -p $(DIST)
	@for goos in $(GOOS_LIST); do \
		for goarch in $(GOARCH_LIST); do \
			OUT=$(DIST)/dis-agent-$${goos}-$${goarch}; \
			mkdir -p $$OUT; \
			GOOS=$${goos} GOARCH=$${goarch} go build -ldflags "$(LDFLAGS)" -o $$OUT/dis-agent ./cmd/agent; \
		done; \
	done

package-tar:
	@for dir in $(DIST)/dis-agent-*; do \
		base=$$(basename $$dir); \
		tar czf $(DIST)/$$base.tar.gz -C $(DIST) $$base; \
	done

package-linux: build-binaries
	@which nfpm >/dev/null || (echo "nfpm not installed"; exit 1)
	@for arch in $(GOARCH_LIST); do \
		GOOS=linux ARCH=$${arch} VERSION=$(VERSION) nfpm package -f packaging/nfpm.yaml -p deb -t $(DIST)/dis-agent-linux-$${arch}.deb; \
		GOOS=linux ARCH=$${arch} VERSION=$(VERSION) nfpm package -f packaging/nfpm.yaml -p rpm -t $(DIST)/dis-agent-linux-$${arch}.rpm; \
	done

checksum:
	cd $(DIST) && shasum -a 256 *.tar.gz > SHA256SUMS

clean:
	rm -rf $(DIST)
