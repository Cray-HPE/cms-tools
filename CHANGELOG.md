# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.8.1] - 2022-09-06

### Changed

- Spelling corrections.
- Removed two lines that dereferenced null error pointers in VCS test error paths.

## [1.8.0] - 2022-08-24

### Changed

- cmsdev: Modify BOS tests to reflect change of default CLI version back to v1.

## [1.7.0] - 2022-08-22

### Added

- cmsdev:
  - Add testing of BOSv2 (both API and CLI), including the new BOSv2 `components` and `options`. Restored testing of CLI without explicitly specifying the BOS version.
  - Include CLI command that failed in corresponding error messages.
  - Add testing of BOS (v1 and v2) `healthz` endpoints and CLI.

## [1.6.1] - 2022-08-17

### Removed

- cmsdev: Removed testing of BOS CLI without explicitly specifying the version. For BOS, the CLI now defaults to BOSv2, but cmsdev has not yet been updated to support BOSv2.

## [1.6.0] - 2022-07-27

### Changed

- Barebones boot test now allows user to specify ID of IMS image used for test

## [1.5.0] - 2022-07-25

### Changed

- Barebones boot test now fails if user specifies a compute node that is not available.
- Modified RPM build process to use RPM release and version fields

## [1.4.0] - 2022-07-13

### Added
- Enabled gitversion and gitflow

### Changed

[Unreleased]: https://github.com/Cray-HPE/cms-tools/compare/1.8.0...HEAD

[1.8.0]: https://github.com/Cray-HPE/cray-product-catalog/compare/1.7.0...1.8.0

[1.7.0]: https://github.com/Cray-HPE/cray-product-catalog/compare/1.6.1...1.7.0

[1.6.1]: https://github.com/Cray-HPE/cray-product-catalog/compare/1.6.0...1.6.1

[1.6.0]: https://github.com/Cray-HPE/cray-product-catalog/compare/1.5.0...1.6.0

[1.5.0]: https://github.com/Cray-HPE/cray-product-catalog/compare/1.4.0...1.5.0

[1.4.0]: https://github.com/Cray-HPE/cray-product-catalog/compare/1.3.3...1.4.0
