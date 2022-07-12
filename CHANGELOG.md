# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

## [0.5.2] - 2022-07-08

### Added

- added rte_ctrl script to test through the terminal

## [0.5.1] - 2019-02-05

### Changed

- Version string is stored in code, not in Makefile. Version string injection
  via `ldflags` can be troublesome for external build systems (e.g. Yocto).

## [0.5.0] - 2019-02-05

### Added

- introduced versioning
- introduced Makefile
- first public release

[Unreleased]: https://github.com/3mdeb/RteCtrl/compare/0.5.2...HEAD
[0.5.2]: https://github.com/3mdeb/RteCtrl/compare/0.5.1...0.5.2
[0.5.1]: https://github.com/3mdeb/RteCtrl/compare/0.5.0...0.5.1
[0.5.0]: https://github.com/3mdeb/RteCtrl/compare/5a814faf0c2a588c5a7ff42b849147c0cbacff1e...0.5.1
