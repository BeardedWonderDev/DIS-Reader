VERSION ?= $(shell git describe --tags --always --dirty)
GOOS_LIST = linux darwin windows
GOARCH_LIST = amd64 arm64
DIST := dist
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: release-agent clean test

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

checksum:
	cd $(DIST) && shasum -a 256 *.tar.gz > SHA256SUMS

clean:
	rm -rf $(DIST)
