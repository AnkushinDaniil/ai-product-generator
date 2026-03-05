## [2.0.0](https://github.com/AnkushinDaniil/memex/compare/v1.0.0...v2.0.0) (2026-03-05)

### ⚠ BREAKING CHANGES

* Project restructured to focus solely on Memex MCP Server

- Remove AI Product Generator specific directories (agents/, feedback/, cmd/server/, specs/)
- Update module name from github.com/ai-product-generator to github.com/AnkushinDaniil/memex
- Update all import paths to use new module name
- Update README.md to describe Memex MCP Server
- Update Dockerfile for Memex (CGO enabled for SQLite)
- Update Makefile to build memex binary
- Update GitHub workflows (CI, release) to reference Memex
- Update .claude/CLAUDE.md project documentation

AI Product Generator is a separate project and has been removed from this repository.
This repository is now dedicated to the Memex MCP Server for persistent AI memory storage.

### Features

* **mcp:** implement MCP protocol handler ([#6](https://github.com/AnkushinDaniil/memex/issues/6)) ([828b0ce](https://github.com/AnkushinDaniil/memex/commit/828b0ce77d74752c7ee725d766067af81a7fe07e)), closes [#2](https://github.com/AnkushinDaniil/memex/issues/2)

### Code Refactoring

* separate Memex project from AI Product Generator ([7783167](https://github.com/AnkushinDaniil/memex/commit/7783167bf47c89a3c77b0c4f98ff7076c89a46ac))

## 1.0.0 (2026-03-04)

### Features

* initial project setup with MCP server and CI/CD infrastructure ([f915ca7](https://github.com/AnkushinDaniil/ai-product-generator/commit/f915ca71aed1fcb6817d67f90f171fd3d1a706ef))

### Bug Fixes

* **ci:** install semantic-release locally instead of globally ([6c5c001](https://github.com/AnkushinDaniil/ai-product-generator/commit/6c5c001d02399127e7e984eec15c273f5a09d66a))
* **ci:** specify golangci-lint v2.10.1 version explicitly ([d245d1a](https://github.com/AnkushinDaniil/ai-product-generator/commit/d245d1a0636aaeec1562be9076695d4788de513d))
* **ci:** upgrade to golangci-lint v2 and adjust coverage threshold ([d8b4386](https://github.com/AnkushinDaniil/ai-product-generator/commit/d8b4386516f2ad3af861c7587f958cdab6f7d243))
* **ci:** upgrade to golangci-lint-action v7 for v2 support ([de652b4](https://github.com/AnkushinDaniil/ai-product-generator/commit/de652b44bd5b09981de779ae6f9d7acb9d10a3c3))
