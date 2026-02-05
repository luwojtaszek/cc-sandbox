# Changelog

## [1.5.0](https://github.com/luwojtaszek/cc-sandbox/compare/cc-sandbox-v1.4.0...cc-sandbox-v1.5.0) (2026-02-05)


### Features

* add Claude config repo support with install.sh hook ([dc7c3fb](https://github.com/luwojtaszek/cc-sandbox/commit/dc7c3fbc87063cfd4b80d2a6e3d5dbbe96185ed4))
* auth command ([ce7cef5](https://github.com/luwojtaszek/cc-sandbox/commit/ce7cef544063c4722b2d0148809558221c6a7ad5))

## [1.4.0](https://github.com/luwojtaszek/cc-sandbox/compare/cc-sandbox-v1.3.2...cc-sandbox-v1.4.0) (2026-01-23)


### Features

* docs and cli updates ([44310b0](https://github.com/luwojtaszek/cc-sandbox/commit/44310b0be02d92de17ea9805b17ba1313c9dc385))

## [1.3.2](https://github.com/luwojtaszek/cc-sandbox/compare/cc-sandbox-v1.3.1...cc-sandbox-v1.3.2) (2026-01-23)


### Bug Fixes

* upgrade Go version to 1.24 in CI for faster CLI binaries ([a9fd612](https://github.com/luwojtaszek/cc-sandbox/commit/a9fd612e4f16cfa25fb48b5b7c4017d0627b0c04))

## [1.3.1](https://github.com/luwojtaszek/cc-sandbox/compare/cc-sandbox-v1.3.0...cc-sandbox-v1.3.1) (2026-01-23)


### Bug Fixes

* reduce subprocess calls via batching and parallel detection ([72b7fa1](https://github.com/luwojtaszek/cc-sandbox/commit/72b7fa10b337cb0e63632b0d3f56a5718ad8fe48))

## [1.3.0](https://github.com/luwojtaszek/cc-sandbox/compare/cc-sandbox-v1.2.0...cc-sandbox-v1.3.0) (2026-01-22)


### Features

* add self-update command and various improvements ([0a8972a](https://github.com/luwojtaszek/cc-sandbox/commit/0a8972a69e27b4829298bb94aaeb441ceda919ed))
* Code improvements ([ffb1996](https://github.com/luwojtaszek/cc-sandbox/commit/ffb1996d25c0bceee05bde6e126c8675d69df6a0))

## [1.2.0](https://github.com/luwojtaszek/cc-sandbox/compare/cc-sandbox-v1.1.2...cc-sandbox-v1.2.0) (2026-01-18)


### Features

* add self-update command for CLI and Docker images ([4daf438](https://github.com/luwojtaszek/cc-sandbox/commit/4daf438e60b0d6577eb23cf0b923a4066491d8db))

## [1.1.2](https://github.com/luwojtaszek/cc-sandbox/compare/cc-sandbox-v1.1.1...cc-sandbox-v1.1.2) (2026-01-18)


### Bug Fixes

* resolve CLI argument parsing and Docker permission issues ([7684183](https://github.com/luwojtaszek/cc-sandbox/commit/7684183b9839314757ea0cd45bd06fb6bf16b7d7))

## [1.1.1](https://github.com/luwojtaszek/cc-sandbox/compare/cc-sandbox-v1.1.0...cc-sandbox-v1.1.1) (2026-01-18)


### Bug Fixes

* resolve Windows cross-compilation and build on release only ([20d6423](https://github.com/luwojtaszek/cc-sandbox/commit/20d642309d6bbff6779aa0e1fcbe33f9a983a68a))

## [1.1.0](https://github.com/luwojtaszek/cc-sandbox/compare/cc-sandbox-v1.0.0...cc-sandbox-v1.1.0) (2026-01-18)


### Features

* add cleanup job for untagged container images ([af62e73](https://github.com/luwojtaszek/cc-sandbox/commit/af62e73ebde4f8a7c52e75032d7cb6a0181dbb84))


### Bug Fixes

* add missing go.sum for CLI dependencies ([639b3bc](https://github.com/luwojtaszek/cc-sandbox/commit/639b3bcc2a49b9b2e5384e867a47db5ce724a9d8))
* simplify ghcr-cleanup-action configuration ([6a4728c](https://github.com/luwojtaszek/cc-sandbox/commit/6a4728cac4c41e4b1719ae97ff0910ddd203974f))
* use version instead of tag_name for image tags ([04a93aa](https://github.com/luwojtaszek/cc-sandbox/commit/04a93aa604c7d652a81427efb2ca71e2ebf837a5))


### Performance

* skip redundant image builds on release ([7f3ceeb](https://github.com/luwojtaszek/cc-sandbox/commit/7f3ceebdc36c030dac353bee993c7e52edfac5fb))

## 1.0.0 (2026-01-18)


### Features

* consolidate CI workflows and add image versioning ([6d0e88c](https://github.com/luwojtaszek/cc-sandbox/commit/6d0e88c1c1ecaf73d7fc4cdb1a3aa2ae2d404f21))
