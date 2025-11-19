# Changelog

All notable changes to this project will be documented in this file.

## [0.1.0] - 2025-11-19
First public release of DIS Reader.

### Highlights
- Self-contained Go backend that embeds the DIS JDBC runner and exposes typed services for units, invoices (including whole goods), and parts inventory.
- Terminal UI (tview2 default, optional Bubble Tea flavor) with authentication, logging panel, pluggable modules, and the Debug Search modal.
- Debug Search batching engine that fans curated SQL queries across FILEC tables, streams progress/events, and exports to SQLite or CSV with user-configurable defaults.
- Configuration system driven by `disreader.yaml` plus `DISREADER_*` env vars for themes, JDBC settings, and debug-search defaults, backed by documentation and fixtures in `wiki/`.

### Fixed
- Decode epoch-formatted DATE/TIMESTAMP values returned by the JDBC runner so AS/400-native date columns hydrate correctly without warning spam.

