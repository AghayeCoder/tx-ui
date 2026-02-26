SHELL := /bin/bash
.DEFAULT_GOAL := help

PLATFORM ?= amd64
XRAY_VERSION ?= v26.2.6
XRAY_URL_BASE := https://github.com/XTLS/Xray-core/releases/download/$(XRAY_VERSION)
RULES_URL_BASE := https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download
IRAN_RULES_URL_BASE := https://github.com/chocolate4u/Iran-v2ray-rules/releases/latest/download
RU_RULES_URL_BASE := https://github.com/runetfreedom/russia-v2ray-rules-dat/releases/latest/download

BUILD_DIR := x-ui
BIN_DIR := $(BUILD_DIR)/bin

.PHONY: help clean build-linux package-linux release-linux build-windows package-windows release-windows release-all

help:
	@echo "TX-UI local release build (mirrors .github/workflows/release.yml)"
	@echo ""
	@echo "Usage:"
	@echo "  make release-linux PLATFORM=amd64"
	@echo "  make release-linux PLATFORM=arm64"
	@echo "  make release-linux PLATFORM=armv7"
	@echo "  make release-linux PLATFORM=armv6"
	@echo "  make release-linux PLATFORM=armv5"
	@echo "  make release-linux PLATFORM=386"
	@echo "  make release-linux PLATFORM=s390x"
	@echo "  make release-windows"
	@echo "  make release-all"
	@echo ""
	@echo "Artifacts:"
	@echo "  x-ui-linux-<platform>.tar.gz"
	@echo "  x-ui-windows-amd64.zip"

clean:
	rm -rf $(BUILD_DIR) xui-release xui-release.exe x-ui-linux-*.tar.gz x-ui-windows-amd64.zip

build-linux:
	@set -euo pipefail; \
	echo "==> Building Linux ($(PLATFORM))"; \
	rm -rf "$(BUILD_DIR)"; \
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GOARM= CC= GOHOSTOS="$$(go env GOHOSTOS)" GOHOSTARCH="$$(go env GOHOSTARCH)"; \
	case "$(PLATFORM)" in \
		amd64) \
			GOARCH=amd64; \
			if [ "$$GOHOSTOS" = "linux" ] && [ "$$GOHOSTARCH" = "amd64" ]; then CC="$${CC:-musl-gcc}"; \
			else CC="$${CC:-x86_64-linux-musl-gcc}"; fi ;; \
		arm64) GOARCH=arm64; CC="$${CC:-aarch64-linux-musl-gcc}" ;; \
		armv7) GOARCH=arm; GOARM=7; CC="$${CC:-arm-linux-musleabihf-gcc}" ;; \
		armv6) GOARCH=arm; GOARM=6; CC="$${CC:-arm-linux-musleabihf-gcc}" ;; \
		armv5) GOARCH=arm; GOARM=5; CC="$${CC:-arm-linux-musleabi-gcc}" ;; \
		386) GOARCH=386; CC="$${CC:-i686-linux-musl-gcc}" ;; \
		s390x) GOARCH=s390x; CC="$${CC:-s390x-linux-musl-gcc}" ;; \
		*) echo "Unsupported PLATFORM: $(PLATFORM)"; exit 1 ;; \
	esac; \
	if ! command -v "$$CC" >/dev/null 2>&1; then \
		echo "Missing musl compiler: $$CC (set CC=/path/to/compiler to override)"; \
		exit 1; \
	fi; \
	export CGO_ENABLED GOOS GOARCH GOARM CC; \
	go build -ldflags "-w -s" -o xui-release -v main.go; \
	mkdir -p "$(BIN_DIR)"; \
	cp xui-release "$(BUILD_DIR)/"; \
	cp x-ui.service "$(BUILD_DIR)/"; \
	cp x-ui.sh "$(BUILD_DIR)/"; \
	mv "$(BUILD_DIR)/xui-release" "$(BUILD_DIR)/x-ui"; \
	cd "$(BIN_DIR)"; \
	download() { \
		if command -v curl >/dev/null 2>&1; then curl -fsSL "$$1" -o "$$2"; \
		else wget -q "$$1" -O "$$2"; fi; \
	}; \
	case "$(PLATFORM)" in \
		amd64) XRAY_ZIP="Xray-linux-64.zip" ;; \
		arm64) XRAY_ZIP="Xray-linux-arm64-v8a.zip" ;; \
		armv7) XRAY_ZIP="Xray-linux-arm32-v7a.zip" ;; \
		armv6) XRAY_ZIP="Xray-linux-arm32-v6.zip" ;; \
		armv5) XRAY_ZIP="Xray-linux-arm32-v5.zip" ;; \
		386) XRAY_ZIP="Xray-linux-32.zip" ;; \
		s390x) XRAY_ZIP="Xray-linux-s390x.zip" ;; \
	esac; \
	download "$(XRAY_URL_BASE)/$$XRAY_ZIP" "$$XRAY_ZIP"; \
	unzip -q "$$XRAY_ZIP"; \
	rm -f "$$XRAY_ZIP"; \
	rm -f geoip.dat geosite.dat; \
	download "$(RULES_URL_BASE)/geoip.dat" geoip.dat; \
	download "$(RULES_URL_BASE)/geosite.dat" geosite.dat; \
	download "$(IRAN_RULES_URL_BASE)/geoip.dat" geoip_IR.dat; \
	download "$(IRAN_RULES_URL_BASE)/geosite.dat" geosite_IR.dat; \
	download "$(RU_RULES_URL_BASE)/geoip.dat" geoip_RU.dat; \
	download "$(RU_RULES_URL_BASE)/geosite.dat" geosite_RU.dat; \
	mv xray "xray-linux-$(PLATFORM)"; \
	cd - >/dev/null

package-linux:
	@set -euo pipefail; \
	if [ ! -d "$(BUILD_DIR)" ]; then \
		echo "Build dir missing. Run: make build-linux PLATFORM=$(PLATFORM)"; \
		exit 1; \
	fi; \
	tar -zcvf "x-ui-linux-$(PLATFORM).tar.gz" "$(BUILD_DIR)"

release-linux: build-linux package-linux

build-windows:
	@set -euo pipefail; \
	echo "==> Building Windows (amd64)"; \
	rm -rf "$(BUILD_DIR)"; \
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 GOHOSTOS="$$(go env GOHOSTOS)" CC=; \
	if [ "$$GOHOSTOS" != "windows" ]; then CC="$${CC:-x86_64-w64-mingw32-gcc}"; fi; \
	if [ -n "$$CC" ] && ! command -v "$$CC" >/dev/null 2>&1; then \
		echo "Missing compiler: $$CC (set CC=/path/to/compiler to override)"; \
		exit 1; \
	fi; \
	export CGO_ENABLED GOOS GOARCH CC; \
	go build -ldflags "-w -s" -o xui-release.exe -v main.go; \
	mkdir -p "$(BIN_DIR)"; \
	cp xui-release.exe "$(BUILD_DIR)/"; \
	mkdir -p "$(BUILD_DIR)/bin"; \
	cd "$(BIN_DIR)"; \
	download() { \
		if command -v curl >/dev/null 2>&1; then curl -fsSL "$$1" -o "$$2"; \
		else wget -q "$$1" -O "$$2"; fi; \
	}; \
	download "$(XRAY_URL_BASE)/Xray-windows-64.zip" "Xray-windows-64.zip"; \
	unzip -q "Xray-windows-64.zip"; \
	rm -f "Xray-windows-64.zip"; \
	rm -f geoip.dat geosite.dat; \
	download "$(RULES_URL_BASE)/geoip.dat" geoip.dat; \
	download "$(RULES_URL_BASE)/geosite.dat" geosite.dat; \
	download "$(IRAN_RULES_URL_BASE)/geoip.dat" geoip_IR.dat; \
	download "$(IRAN_RULES_URL_BASE)/geosite.dat" geosite_IR.dat; \
	download "$(RU_RULES_URL_BASE)/geoip.dat" geoip_RU.dat; \
	download "$(RU_RULES_URL_BASE)/geosite.dat" geosite_RU.dat; \
	mv xray.exe xray-windows-amd64.exe; \
	cd - >/dev/null

package-windows:
	@set -euo pipefail; \
	if [ ! -d "$(BUILD_DIR)" ]; then \
		echo "Build dir missing. Run: make build-windows"; \
		exit 1; \
	fi; \
	zip -r "x-ui-windows-amd64.zip" "$(BUILD_DIR)" >/dev/null

release-windows: build-windows package-windows

release-all:
	@set -euo pipefail; \
	for p in amd64 arm64 armv7 armv6 armv5 386 s390x; do \
		$(MAKE) release-linux PLATFORM=$$p; \
	done; \
	$(MAKE) release-windows
