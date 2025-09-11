#!/usr/bin/env bash
set -e
git config core.hooksPath .githooks
echo "Hooks installed (core.hooksPath=.githooks)"
