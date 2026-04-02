# OpenRiot Makefile
# Single source of truth for all version numbers.
# Scripts should read OPENRIOT_VERSION and OPENBSD_VERSION from here.

# ============================================================
# Canonical Versions — single source of truth: VERSION file
# ============================================================
OPENRIOT_VERSION = $(shell cat VERSION 2>/dev/null || echo "0.6")
OPENBSD_VERSION  = 7.9

# ============================================================
# Build Config
# ============================================================
BINARY_NAME = openriot
SOURCE_DIR  = source
INSTALL_DIR = install
ARCH        = amd64

# Inject versions into the Go binary at link time — no hardcoding in .go files
LDFLAGS = -s -w \
	-X main.version=$(OPENRIOT_VERSION) \
	-X main.openbsdVersion=$(OPENBSD_VERSION)

# Export so child processes (scripts) can read them
export OPENRIOT_VERSION
export OPENBSD_VERSION
export ARCH

.PHONY: all build clean deps test verify dev release ultra iso download-packages help

# ============================================================
# Default
# ============================================================
all: build

# ============================================================
# Build targets
# ============================================================

# Standard build — optimized, cross-compiled for OpenBSD
build:
	@echo "=== Building OpenRiot v$(OPENRIOT_VERSION) for OpenBSD $(OPENBSD_VERSION) ==="
	@cd $(SOURCE_DIR) && \
		CGO_ENABLED=0 GOOS=openbsd GOARCH=$(ARCH) \
		go build \
		-ldflags="$(LDFLAGS)" \
		-trimpath \
		-o $(BINARY_NAME) .
	@mv $(SOURCE_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@chmod 0755 $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "=== Build complete: $(INSTALL_DIR)/$(BINARY_NAME) ==="

# Development build — native arch, no cross-compile, faster iteration
dev:
	@echo "=== Development build (native) ==="
	@cd $(SOURCE_DIR) && \
		go build \
		-ldflags="-X main.version=$(OPENRIOT_VERSION) -X main.openbsdVersion=$(OPENBSD_VERSION)" \
		-o $(BINARY_NAME) .
	@mv $(SOURCE_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@chmod 0755 $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "=== Dev build complete: $(INSTALL_DIR)/$(BINARY_NAME) ==="

# Release build — same as build, explicit target
release: build
	@echo "=== Release v$(OPENRIOT_VERSION) ready ==="

# Ultra build — maximum size reduction, optional UPX compression
ultra:
	@echo "=== Ultra-optimized build ==="
	@cd $(SOURCE_DIR) && \
		CGO_ENABLED=0 GOOS=openbsd GOARCH=$(ARCH) \
		go build \
		-ldflags="$(LDFLAGS) -extldflags '-static'" \
		-trimpath \
		-o $(BINARY_NAME) .
	@mv $(SOURCE_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@chmod 0755 $(INSTALL_DIR)/$(BINARY_NAME)
	@if command -v upx > /dev/null 2>&1; then \
		echo "Compressing with UPX..."; \
		upx --best --lzma $(INSTALL_DIR)/$(BINARY_NAME); \
	else \
		echo "UPX not found — skipping compression"; \
	fi
	@echo "=== Ultra build complete ==="

# ============================================================
# ISO / Package targets
# ============================================================

# Download all offline packages (populates ~/.pkgcache/7.9/amd64/)
download-packages:
	@echo "=== Downloading OpenBSD $(OPENBSD_VERSION) packages ==="
	@scripts/download-packages.sh

# Build the full bootable ISO — downloads packages first if needed
iso: build download-packages
	@echo "=== Building OpenRiot $(OPENRIOT_VERSION) ISO ==="
	@./build-iso.sh

# ============================================================
# Utility targets
# ============================================================

# Tidy Go module dependencies
deps:
	@echo "=== Updating Go dependencies ==="
	@cd $(SOURCE_DIR) && go mod tidy
	@echo "=== Dependencies updated ==="

# Run Go tests
test:
	@echo "=== Running tests ==="
	@cd $(SOURCE_DIR) && go test ./...

# Build then smoke-test the binary (uses Linux build for local testing)
verify: build
	@echo "=== Verifying build ==="
	@cd $(SOURCE_DIR) && go build -ldflags="$(LDFLAGS)" -trimpath -o $(BINARY_NAME) .
	@$(SOURCE_DIR)/$(BINARY_NAME) --version
	@echo "=== Binary OK ==="

# Remove build artifacts (keep downloaded ISO and package cache)
clean:
	@echo "=== Cleaning build artifacts ==="
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@rm -f $(SOURCE_DIR)/$(BINARY_NAME)
	@echo "=== Clean complete ==="

# Remove everything including work directory (keeps ~/.pkgcache)
distclean: clean
	@echo "=== Removing work directory ==="
	@rm -rf .work/iso_contents
	@echo "=== distclean complete ==="

# ============================================================
# Help
# ============================================================
help:
	@echo "OpenRiot Build System"
	@echo "====================="
	@echo "Version : $(OPENRIOT_VERSION)"
	@echo "OpenBSD : $(OPENBSD_VERSION)"
	@echo ""
	@echo "Targets:"
	@echo "  build              Build openriot binary (cross-compiled for OpenBSD)"
	@echo "  dev                Fast native build for local testing"
	@echo "  release            Alias for build"
	@echo "  ultra              Maximum-optimized build with optional UPX"
	@echo "  download-packages  Download all offline packages to ~/.pkgcache/"
	@echo "  iso                Build full bootable ISO (runs build first)"
	@echo "  deps               Tidy Go module dependencies"
	@echo "  test               Run Go tests"
	@echo "  verify             Build and smoke-test the binary"
	@echo "  clean              Remove build artifacts"
	@echo "  distclean          clean + remove .work/iso_contents"
	@echo "  help               Show this message"
	@echo ""
	@echo "Typical workflow:"
	@echo "  make download-packages   # downloads packages + index.txt"
	@echo "  make iso                 # builds binary + repacks ISO"
	@echo "  make verify              # smoke-test the binary"
