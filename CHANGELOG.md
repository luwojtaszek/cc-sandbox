# Changelog

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
