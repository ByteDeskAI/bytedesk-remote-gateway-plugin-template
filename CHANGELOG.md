# Changelog

## [0.2.0-rc.2] - 2026-09-09

### Changed

- Adopt Gateway SDK v0.4.0-rc.5 and common SDK v0.4.0-rc.4 with canonical dependency version constraint validation.

## [0.2.0-rc.1] - 2026-09-08

### Changed

- Replace the handler-only scaffold with an independently constructed SDK Plugin package and shared ServePlugin entrypoint.
- Pin released SDK prereleases and remove sibling-directory replacements.
- Generate plugin.json from the implementation and test lifecycle withdrawal and manifest parity.

### Added

- Built-executable conformance verifies denial before socket publication, admitted serving, and graceful withdrawal without Gateway implementation imports.
