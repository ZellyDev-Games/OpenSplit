# Makefile (cross-OS: Linux/macOS + Windows CMD/PowerShell)
SHELL := /usr/bin/env bash

FRONTEND_DIST := frontend/dist
GITKEEP := $(FRONTEND_DIST)/.gitkeep

# Detect Windows vs others
ifeq ($(OS),Windows_NT)
    POWERSHELL := powershell -NoProfile -ExecutionPolicy Bypass
define MKDIRP
	$(POWERSHELL) -Command "New-Item -ItemType Directory -Force -Path '$(1)' | Out-Null"
endef
define TOUCH
	$(POWERSHELL) -Command "New-Item -ItemType File -Force -Path '$(1)' | Out-Null"
endef
define RM_RF
	$(POWERSHELL) -Command "Remove-Item -Recurse -Force -Path '$(1)'"
endef
else
define MKDIRP
	mkdir -p $(1)
endef
define TOUCH
	: > $(1)
endef
define RM_RF
	rm -rf $(1)
endef
endif

.PHONY: test ensure-dist clean

ensure-dist:
	$(call MKDIRP,$(FRONTEND_DIST))
	$(call TOUCH,$(GITKEEP))

test: ensure-dist
	go test ./... -count=1

clean:
	$(call RM_RF,$(FRONTEND_DIST))
	$(call MKDIRP,$(FRONTEND_DIST))
	$(call TOUCH,$(GITKEEP))
