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
.PHONY: all build linux clean deps test verify validate release ultra img isotest binary-push prune install help test-img

# Build only (no install) - use for testing
all:
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

# Install to local (build + copy)
install: all
	@mkdir -p $$HOME/.local/share/openriot/install && \
	cp $(INSTALL_DIR)/$(BINARY_NAME) $$HOME/.local/share/openriot/install/ && \
	echo "=== Local install updated ==="

# Build target fails - use 'make' for dev or 'make install' to deploy
build:
	@echo "ERROR: Use 'make' (dev) or 'make install' (deploy)" && exit 1

# Linux build — native
linux:
	@OPENRIOT_VERSION=`cat VERSION` && \
	echo "=== Building OpenRIOT v$$OPENRIOT_VERSION for Linux ===" && \
	cd $(SOURCE_DIR) && \
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) \
	go build \
	-ldflags="-s -w -X main.version=$$OPENRIOT_VERSION" \
	-trimpath \
	-o ../$(INSTALL_DIR)/$(BINARY_NAME) . && \
	chmod 0755 ../$(INSTALL_DIR)/$(BINARY_NAME) && \
	echo "=== Linux build complete: $(INSTALL_DIR)/$(BINARY_NAME) ==="

# Release build with versioning
release:
	@$(MAKE) validate || { echo "[FAIL] Validation failed. Release aborted."; exit 1; }
	@$(MAKE) test || { echo "[FAIL] Tests failed. Release aborted."; exit 1; }
	@echo "Syncing packages.yaml to latest installed versions..."; ./install/openriot --sync-packages || true; \
	OPENRIOT_VERSION=`cat VERSION` && \
	CURRENT_BRANCH=`git branch --show-current` && \
	echo "Current version: v$$OPENRIOT_VERSION"; \
	echo "Branch: $$CURRENT_BRANCH"; \
	echo ""; \
	MAJOR=`echo $$OPENRIOT_VERSION | cut -d. -f1` && \
	MINOR=`echo $$OPENRIOT_VERSION | cut -d. -f2` && \
	PATCH=`echo $$OPENRIOT_VERSION | cut -d. -f3` && \
	[ -n "$$PATCH" ] || PATCH=0 && \
	if [ "$(BUMP)" = "major" ]; then \
		if [ "$$MINOR" -eq 9 ]; then \
			NEW_VERSION=$$((MAJOR+1)).0.0; \
		else \
			NEW_VERSION=$$MAJOR.$$((MINOR+1)).0; \
		fi; \
		echo "Bump type: major"; \
	else \
		NEW_VERSION=$$MAJOR.$$MINOR.$$((PATCH+1)); \
		echo "Bump type: patch (default)"; \
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
		echo "  1. Sync packages.yaml to latest installed"; \
		echo "  2. Update VERSION to $$NEW_VERSION"; \
		echo "  3. Build binary with new version"; \
		echo "  4. Update README.md badge to v$$NEW_VERSION"; \
		echo "  5. git add -A (all changes)"; \
		echo "  6. git commit (opens Helix editor)"; \
		echo "  7. git tag 'v$$NEW_VERSION'"; \
		echo "  8. Push (if confirmed)"; \
	else \
		echo "Updating VERSION..."; \
		echo $$NEW_VERSION > VERSION; \
		echo "Building binary with v$$NEW_VERSION..."; make; \
		echo "Updating README badge..."; \
		sed -i "s/version-[0-9.]*/version-$$NEW_VERSION/" README.md; \
		echo "Committing..."; \
		git add -A; \
		GIT_EDITOR=hx git commit; \
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
	PATCH=`echo $$OPENRIOT_VERSION | cut -d. -f3` && \
	[ -n "$$PATCH" ] || PATCH=0 && \
	if [ "$(BUMP)" = "major" ]; then \
		if [ "$$MINOR" -eq 9 ]; then \
			NEW_VERSION=$$((MAJOR+1)).0.0; \
		else \
			NEW_VERSION=$$MAJOR.$$((MINOR+1)).0; \
		fi; \
		echo "Bump type: major"; \
	else \
		NEW_VERSION=$$MAJOR.$$MINOR.$$((PATCH+1)); \
		echo "Bump type: patch (default)"; \
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
# Use ~/.tmp to avoid /tmp which is mounted noexec on OpenBSD
test:
	@echo "=== Running tests ==="
	@mkdir -p $(HOME)/.tmp && cd $(SOURCE_DIR) && TMPDIR=$(HOME)/.tmp go test ./... 2>&1 | grep -v 'no test files' | column -t

# Smoke tests - integration tests for install paths
smoke-test:
	@echo "=== Running smoke tests ==="
	@mkdir -p $(HOME)/.tmp && cd $(SOURCE_DIR) && TMPDIR=$(HOME)/.tmp go test -v -run Smoke .

# Imaging module tests
test-img:
	@echo "=== Running imaging tests ==="
	@mkdir -p $(HOME)/.tmp && cd $(SOURCE_DIR) && TMPDIR=$(HOME)/.tmp go test -v ./imaging/...

# Verify build
verify: all
	@$(INSTALL_DIR)/$(BINARY_NAME) --version
	@echo "=== Binary OK ==="

# Pre-release validation gate
validate:
	@echo "=== Validating release readiness ==="
	@_font_count=`ls assets/fonts/* 2>/dev/null | wc -l`; \
	if [ "$$_font_count" -eq 0 ]; then \
		echo "[FAIL] No font files in assets/fonts/"; exit 1; \
	fi
	@echo "[DONE] Fonts present"
	@_cursor_count=`ls assets/cursors/* 2>/dev/null | wc -l`; \
	if [ "$$_cursor_count" -eq 0 ]; then \
		echo "[FAIL] No cursor files in assets/cursors/"; exit 1; \
	fi
	@echo "[DONE] Cursors present"
	@if [ ! -f assets/themes/kora.tgz ]; then \
		echo "[FAIL] Kora theme archive not found at assets/themes/kora.tgz"; exit 1; \
	fi
	@echo "[DONE] Kora theme present"
	@./$(INSTALL_DIR)/$(BINARY_NAME) --validate-config || exit 1
	@$(MAKE) verify
	@echo "=== Validation passed ==="

# Convert PNG backgrounds and lock screens to WebP
convert:
	@echo "=== Converting backgrounds to WebP ==="
	@for f in backgrounds/*.png; do \
		cwebp -q 95 "$$f" -o "backgrounds/$$(basename "$$f" .png).webp" >/dev/null 2>&1; \
	done
	@echo "=== Converting lock screens to WebP ==="
	@for f in Locked/*.png; do \
		cwebp -q 95 "$$f" -o "Locked/$$(basename "$$f" .png).webp" >/dev/null 2>&1; \
	done
	@echo "=== Conversion complete ==="

# Clean
clean:
	@echo "=== Cleaning build artifacts ==="
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@rm -f $(SOURCE_DIR)/$(BINARY_NAME)
	@echo "=== Clean complete ==="

# Prune old PNG blobs from git history (Locked/ and backgrounds/)
prune:
	@set -e; \
	export FILTER_BRANCH_SQUELCH_WARNING=1; \
	echo ""; \
	echo "⚠️  DANGER: This rewrites ALL git history."; \
	echo ""; \
	if [ "$(CONFIRM)" != "yes" ]; then \
		echo "  Run 'make prune CONFIRM=yes' to proceed."; \
		exit 1; \
	fi; \
	if ! git diff --quiet --exit-code; then \
		echo "  ERROR: Unstaged changes. Commit or stash first."; \
		exit 1; \
	fi; \
	BEFORE="$$(du -sh .git | cut -f1)"; \
	OLD_BLOBS="$$(git rev-list --all --objects | git cat-file --batch-check='%(objecttype) %(objectsize) %(rest)' 2>/dev/null | awk '/^blob/ && $$3 ~ /^(Locked|backgrounds)\/.*\.png$$/ {count++; total+=$$2} END {printf "%d blobs, %.0f MB", count, total/1048576}')"; \
	echo "  Stale blobs before: $${OLD_BLOBS} ($${BEFORE})"; \
	git filter-branch --force --index-filter \
		'git rm --cached --ignore-unmatch Locked/*.png backgrounds/*.png' \
		--prune-empty --tag-name-filter cat -- --all; \
	git for-each-ref --format='%(refname)' refs/original/ | \
		while read ref; do git update-ref -d "$$ref"; done; \
	git reflog expire --expire=now --all; \
	git gc --prune=now; \
	AFTER="$$(du -sh .git | cut -f1)"; \
	echo ""; \
	echo "=== Prune complete ==="; \
	echo "  .git size before: $${BEFORE}"; \
	echo "  .git size after:  $${AFTER}"; \
	echo ""; \
	echo "  All .webp files preserved (working tree and history)."; \
	echo ""; \
	echo "  Remotes preserved. To publish rewritten history:"; \
	echo "    git push --force --all && git push --force --tags"
# Custom installer image
# Requires OpenBSD 7.9 host - cross-compilation not possible for bsd.rd modification
image: all
	@if [ "$$(uname -s)" != "OpenBSD" ]; then \
		echo "ERROR: image target requires OpenBSD $(OPENBSD_VERSION)"; \
		echo "Current: $$(uname -s) $$(uname -r)"; \
		echo ""; \
		echo "To build the installer image:"; \
		echo "  1. Boot into OpenBSD $(OPENBSD_VERSION)"; \
		echo "  2. cd to this directory"; \
		echo "  3. doas make image"; \
		exit 1; \
	fi
	@doas rm -rf Build/Output Build/work/site79.tgz
	@doas ./install/openriot --make-image

img:
	@echo "ERROR: 'make img' is deprecated. Use 'make image'."
	@exit 1

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
	@echo "  (default)         Build openriot binary (cross-compiled for OpenBSD)"
	@echo "  install           Build + copy to ~/.local/share/openriot/install/"
	@echo "  build             FAIL - use 'make' or 'make install'"
	@echo "  linux             Build for Linux (native)"
	@echo "  release            Version bump, commit, tag, and push"
	@echo "  ultra              Maximum-optimized static build with optional UPX"
	@echo "  image              Build custom OpenBSD installer image (requires OpenBSD host)"
	@echo "  isotest            Build ISO and run in QEMU"
	@echo "  deps               Tidy Go module dependencies"
	@echo "  test               Run Go tests"
	@echo "  test-img           Run imaging module tests"
	@echo "  verify             Build and smoke-test the binary"
	@echo "  clean              Remove build artifacts"
	@echo "  prune              Remove old PNG blobs from git history (requires CONFIRM=yes)"
	@echo "  convert            Convert PNG backgrounds/lock screens to WebP"
	@echo "  binary-push        Build + strip history + commit + force-push binary"
	@echo "  help               Show this message"

# UI refresh — install binary + sync all UI configs + restart bars/daemons
ui: install
	@echo "=== Syncing UI configs ==="
	@cp config/polybar/config.ini.tmpl $$HOME/.local/share/openriot/config/polybar/config.ini.tmpl
	@openriot --polybar-setup
	@pkill -9 polybar 2>/dev/null || true
	@cp config/dunst/dunstrc $$HOME/.local/share/openriot/config/dunst/dunstrc
	@openriot --dunst-setup
	@pkill -9 dunst 2>/dev/null || true
	@cp config/i3/config $$HOME/.config/i3/config
	@i3-msg restart 2>/dev/null || true
	@cp config/rofi/simple-tokyonight.rasi.tmpl $$HOME/.local/share/openriot/config/rofi/simple-tokyonight.rasi.tmpl
	@openriot --rofi-setup
	@pkill -9 rofi 2>/dev/null || true
	@cp config/rofi/config.rasi $$HOME/.config/rofi/config.rasi
	@cp config/picom.conf $$HOME/.config/picom.conf
	@pkill -9 picom 2>/dev/null || true
	@pkill -9 rofi 2>/dev/null || true
	@echo "=== UI refresh complete ==="
	@echo "Run \`Super+Shift+R\` if anything looks off."
