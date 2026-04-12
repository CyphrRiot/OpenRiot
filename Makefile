# OpenRiot Makefile
# Single source of truth for all version numbers.

# ============================================================
# Build Config
# ============================================================
BINARY_NAME = openriot
SOURCE_DIR  = source
INSTALL_DIR = install
ARCH        = amd64
OPENBSD_VERSION = 7.9

# ============================================================
# Targets
# ============================================================
.PHONY: all build clean deps test verify dev release ultra iso isotest binary-push help

all: build

# Standard build — cross-compiled for OpenBSD
build:
	@OPENRIOT_VERSION=`cat VERSION` && \
	echo "=== Building OpenRIOT v$$OPENRIOT_VERSION for OpenBSD $(OPENBSD_VERSION) ===" && \
	cd $(SOURCE_DIR) && \
	CGO_ENABLED=0 GOOS=openbsd GOARCH=$(ARCH) \
	go build \
	-ldflags="-s -w -X main.version=$$OPENRIOT_VERSION -X main.openbsdVersion=$(OPENBSD_VERSION)" \
	-trimpath \
	-o ../$(INSTALL_DIR)/$(BINARY_NAME) . && \
	chmod 0755 ../$(INSTALL_DIR)/$(BINARY_NAME) && \
	echo "=== Build complete: $(INSTALL_DIR)/$(BINARY_NAME) ==="

# Development build — native arch, faster iteration
dev:
	@OPENRIOT_VERSION=`cat VERSION` && \
	echo "=== Development build (native) ===" && \
	cd $(SOURCE_DIR) && \
	go build \
	-ldflags="-X main.version=$$OPENRIOT_VERSION -X main.openbsdVersion=$(OPENBSD_VERSION)" \
	-trimpath \
	-o ../$(INSTALL_DIR)/$(BINARY_NAME) . && \
	chmod 0755 ../$(INSTALL_DIR)/$(BINARY_NAME) && \
	echo "=== Dev build complete: $(INSTALL_DIR)/$(BINARY_NAME) ==="

# Release build with versioning
release:
	@echo "Building binary..."; make build; echo ""; \
	OPENRIOT_VERSION=`cat VERSION` && \
	CURRENT_BRANCH=`git branch --show-current` && \
	echo "Current version: v$$OPENRIOT_VERSION"; \
	echo "Branch: $$CURRENT_BRANCH"; \
	echo ""; \
	MAJOR=`echo $$OPENRIOT_VERSION | cut -d. -f1` && \
	MINOR=`echo $$OPENRIOT_VERSION | cut -d. -f2` && \
	if [ "$(BUMP)" = "major" ]; then \
		NEW_VERSION=$$((MAJOR+1)).0; \
		echo "Bump type: major"; \
	else \
		NEW_VERSION=$$MAJOR.$$((MINOR+1)); \
		echo "Bump type: minor (default)"; \
	fi && \
	echo "New version: v$$NEW_VERSION"; \
	echo ""; \
	echo "Changes since last commit:"; \
	git diff --stat; \
	echo ""; \
	if [ -n "$$(git status --porcelain)" ]; then \
		echo "WARNING: Uncommitted changes exist:"; \
		git status --short; \
		echo ""; \
		echo "Press Enter to proceed or Ctrl+C to abort."; \
		read -r; \
	fi; \
	if [ -n "$(DRYRUN)" ]; then \
		echo "=== DRY RUN - No changes made ==="; \
		echo "Would do:"; \
		echo "  1. Build binary"; \
		echo "  2. Update VERSION to $$NEW_VERSION"; \
		echo "  3. Update README.md badge to v$$NEW_VERSION"; \
		echo "  4. git add -A (all changes)"; \
		echo "  5. git commit (opens Helix editor)"; \
		echo "  6. git tag 'v$$NEW_VERSION'"; \
		echo "  7. Push (if confirmed)"; \
	else \
		echo "Updating VERSION..."; \
		echo $$NEW_VERSION > VERSION; \
		echo "Updating README badge..."; \
		sed -i "s/version-[0-9.]*/version-$$NEW_VERSION/" README.md; \
		echo "Committing..."; \
		git add -A; \
		git commit; \
		echo "Creating tag..."; \
		git tag -a "v$$NEW_VERSION" -m "OpenRiot v$$NEW_VERSION"; \
		echo ""; \
		echo "=== Release v$$NEW_VERSION created ==="; \
		echo ""; \
		echo "Would you like to push and tag? [Y/n]"; \
		read -r PUSH_CONF; \
		if [ "$${PUSH_CONF:-y}" = "y" ] || [ "$${PUSH_CONF:-y}" = "Y" ]; then \
			git push && git push origin "v$$NEW_VERSION"; \
		fi \
	fi

# Create release (called by setup.sh after confirm)
create-release:
	@OPENRIOT_VERSION=`cat VERSION` && \
	MAJOR=`echo $$OPENRIOT_VERSION | cut -d. -f1` && \
	MINOR=`echo $$OPENRIOT_VERSION | cut -d. -f2` && \
	if [ "$(BUMP)" = "major" ]; then \
		NEW_VERSION=$$((MAJOR+1)).0; \
		echo "Bump type: major"; \
	else \
		NEW_VERSION=$$MAJOR.$$((MINOR+1)); \
		echo "Bump type: minor (default)"; \
	fi && \
	echo "Updating VERSION: v$$OPENRIOT_VERSION -> v$$NEW_VERSION"; \
	echo $$NEW_VERSION > VERSION && \
	echo "Updating README badge..."; \
	sed -i "s/version-[0-9.]*/version-$$NEW_VERSION/" README.md && \
	echo "Committing changes..."; \
	git add VERSION README.md && \
	git commit -m "v$$NEW_VERSION" && \
	echo "Creating tag..."; \
	git tag -a "v$$NEW_VERSION" -m "OpenRiot v$$NEW_VERSION" && \
	echo ""; \
	echo "=== Release v$$NEW_VERSION created ==="; \
	echo ""; \
	echo "Next steps:"; \
	echo "  git push && git push --tags"

# Ultra build — static + optional UPX
ultra:
	@OPENRIOT_VERSION=`cat VERSION` && \
	echo "=== Ultra-optimized build ===" && \
	cd $(SOURCE_DIR) && \
	CGO_ENABLED=0 GOOS=openbsd GOARCH=$(ARCH) \
	go build \
	-ldflags="-s -w -X main.version=$$OPENRIOT_VERSION -X main.openbsdVersion=$(OPENBSD_VERSION) -extldflags=-static" \
	-trimpath \
	-o ../$(INSTALL_DIR)/$(BINARY_NAME) . && \
	chmod 0755 ../$(INSTALL_DIR)/$(BINARY_NAME) && \
	if command -v upx > /dev/null 2>&1; then \
		echo "Compressing with UPX..."; \
		upx --best --lzma ../$(INSTALL_DIR)/$(BINARY_NAME); \
	else \
		echo "UPX not found — skipping compression"; \
	fi && \
	echo "=== Ultra build complete ==="

# Dependency management
deps:
	@echo "=== Updating Go dependencies ==="
	@cd $(SOURCE_DIR) && go mod tidy
	@echo "=== Dependencies updated ==="

# Testing
test:
	@echo "=== Running tests ==="
	@cd $(SOURCE_DIR) && go test ./...

# Verify build
verify: dev
	@$(INSTALL_DIR)/$(BINARY_NAME) --version
	@echo "=== Binary OK ==="

# Clean
clean:
	@echo "=== Cleaning build artifacts ==="
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@rm -f $(SOURCE_DIR)/$(BINARY_NAME)
	@echo "=== Clean complete ==="

# ISO build - DEPRECATED
# Custom ISO is no longer needed. Use standard OpenBSD ISO + setup.sh.
# See README.md for installation instructions.

# Binary push
binary-push: build
	@BINARY_BLOBS=`git log --oneline --all -- install/openriot 2>/dev/null | wc -l`; \
	if [ "$$BINARY_BLOBS" -gt 1 ]; then \
		echo "WARNING: Binary has $$BINARY_BLOBS commits in history. Stripping..."; \
		git filter-repo --force --path install/openriot --invert-paths 2>/dev/null; \
		git remote add origin git@github.com:CyphrRiot/OpenRiot.git 2>/dev/null || true; \
		$(MAKE) build; \
	fi
	@OPENRIOT_VERSION=`cat VERSION` && \
	echo "=== Committing binary ===" && \
	git add install/openriot .gitignore && \
	git commit -am "v$$OPENRIOT_VERSION: update openriot binary" && \
	echo "=== Force-pushing ===" && \
	git push --force --all && \
	git push --tags 2>/dev/null || true && \
	echo "=== Binary push complete ==="

# Help
help:
	@echo "OpenRiot Makefile (BSD make compatible)"
	@echo ""
	@OPENRIOT_VERSION=`cat VERSION` && \
	echo "Version : $$OPENRIOT_VERSION"
	@echo "OpenBSD : $(OPENBSD_VERSION)"
	@echo ""
	@echo "Targets:"
	@echo "  build              Build openriot binary (cross-compiled for OpenBSD)"
	@echo "  dev                Fast native build for local testing"
	@echo "  release            Alias for build"
	@echo "  ultra              Maximum-optimized build with optional UPX"
	@echo "  iso                Build full bootable ISO"
	@echo "  isotest            Build ISO and run in QEMU"
	@echo "  deps               Tidy Go module dependencies"
	@echo "  test               Run Go tests"
	@echo "  verify             Build and smoke-test the binary"
	@echo "  clean              Remove build artifacts"
	@echo "  binary-push        Build + strip history + commit + force-push binary"
	@echo "  help               Show this message"
