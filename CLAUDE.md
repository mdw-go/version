# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A CLI tool for tagging git repositories with semantic versions. Supports both single-module and multi-module Go repositories.

## Build Commands

- `make install` - Format, test, and install the binary (preferred for development)
- `make test` - Run tests with race detection and coverage
- `make fmt` - Format code and tidy modules

## Architecture

Single-file CLI application (`cmd/version/main.go`) that:
1. Detects Go module context by traversing from cwd toward git root looking for go.mod
2. Finds the latest semver tag from git (filtered by module prefix for multi-module repos)
3. Prompts user to select next version (major/minor/patch bump or custom dev version)
4. Creates an annotated git tag

### Multi-module Repository Support

- Single-module repo: tags are plain semver (e.g., `v1.2.3`)
- Multi-module repo: tags include relative path prefix (e.g., `cmd/foo/v1.2.3`)

### Custom Version Format

Dev versions follow: `{major}.{minor}.{patch}-dev-{username}-{timestamp}`

Username is derived from (in order): `VERSION_USERNAME` env var, git branch prefix before `/`, or OS username.
