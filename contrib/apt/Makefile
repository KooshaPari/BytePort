# BytePort Debian/Ubuntu APT Package Builder
#
# This makefile builds a .deb package from the Go backend binary.
# Usage:
#   make deb        — build the .deb
#   make install    — install via dpkg
#   make clean      — remove build artifacts

PACKAGE_NAME     := byteport
VERSION          ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0-dev")
ARCH             := amd64
DEB_DIR          := build/debian

# Go build settings
BINARY           := $(DEB_DIR)/usr/bin/byteport
BUILD_DIR        := build
GO_SRC           := ../backend
GO_FLAGS         := -ldflags="-s -w -X main.version=$(VERSION)"

# Packaging paths
DEB_OUTPUT       := $(BUILD_DIR)/$(PACKAGE_NAME)_$(VERSION)_$(ARCH).deb

.PHONY: deb install clean

$(BINARY):
	@mkdir -p $(dir $@)
	cd $(GO_SRC) && CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) \
		go build $(GO_FLAGS) -o ../../$(BINARY) ./cmd/api-server/

deb: $(BINARY) control postinst prerm
	@mkdir -p $(DEB_DIR)/DEBIAN $(DEB_DIR)/etc/byteport $(DEB_DIR)/lib/systemd/system
	# Copy systemd unit
	cp ../../contrib/systemd/byteport.service $(DEB_DIR)/lib/systemd/system/
	# Copy default config
	cp ../../contrib/apt/byteport.default $(DEB_DIR)/etc/byteport/env
	# Copy control files
	cp control     $(DEB_DIR)/DEBIAN/
	cp postinst    $(DEB_DIR)/DEBIAN/
	cp prerm       $(DEB_DIR)/DEBIAN/
	# Calculate checksums
	cd $(DEB_DIR) && find usr etc lib -type f | xargs md5sum > DEBIAN/md5sums
	# Build .deb
	dpkg-deb --build $(DEB_DIR) $(DEB_OUTPUT)
	@echo "Package built: $(DEB_OUTPUT)"
	@echo "  dpkg -i $(DEB_OUTPUT)"

install: deb
	sudo dpkg -i $(DEB_OUTPUT) || sudo apt-get install -f -y

clean:
	rm -rf $(BUILD_DIR)
