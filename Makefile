VERSION ?= $(shell git describe --tags --always --dirty)
GOOS_LIST = linux darwin windows
GOARCH_LIST = amd64 arm64
DIST := dist
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: release-agent clean test build-binaries package-tar checksum package-linux package-macos package-windows

release-agent: clean test build-binaries package-linux package-macos package-windows package-tar checksum

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

package-macos: build-binaries
	@for arch in amd64 arm64; do \
		VERSION=$(VERSION) ARCH=$${arch} packaging/macos/pkgbuild.sh; \
	done

package-windows: build-binaries
	@for arch in $(GOARCH_LIST); do \
		OUT=$(DIST)/dis-agent-windows-$${arch}; \
		cp packaging/examples/agent.yaml $$OUT/agent.yaml; \
		cp packaging/examples/bridge_agents.yaml $$OUT/bridge_agents.yaml; \
		cp packaging/windows/install_service.ps1 $$OUT/install_service.ps1; \
		cp packaging/windows/uninstall_service.ps1 $$OUT/uninstall_service.ps1; \
		(cd $$OUT && zip -r ../dis-agent-windows-$${arch}.zip dis-agent agent.yaml bridge_agents.yaml install_service.ps1 uninstall_service.ps1); \
	done

checksum:
	cd $(DIST) && shasum -a 256 *.tar.gz > SHA256SUMS

clean:
	rm -rf $(DIST)
