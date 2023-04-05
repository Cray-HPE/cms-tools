# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.10.7] - 2023-04-05

### Changed

- cmsdev: Modify BOS test to make pod checks more flexible

## [1.10.6] - 2023-04-04

### Changed

- cmsdev: Bump golang.org/x/sys to 0.1.0
  - Resolves CVE-2022-29526
- cmsdev: Bump golang.org/x/net to 0.7.0
  - Resolves CVE-2022-41723

## [1.10.5] - 2023-03-30

### Changed

- cmsdev: Update default CLI BOS version to v2

## [1.10.4] - 2023-03-22

### Removed

- cmsdev: Stop building SP3 RPM. This is version of the cmsdev tool is only intended to run on NCNs of CSM 1.4 or higher,
  and those will be running at least SP4.
- cmsdev: Remove cmsdev binary which somehow made it into the repository.

## [1.10.3] - 2023-03-22

### Added

- cmsdev: Add -l option to "tests" to list possible tests to run. --exclude-alises may be specified
  to exclude aliases from this listing.

### Changed

- cmsdev: Fixed null pointer exception when run before CSM is deployed.

## [1.10.2] - 2023-03-16

### Changed

- cmsdev: Cosmetic changes made to appease latest gofmt version
- cmsdev: Do not run tftp file transfer test from master NCNs

## [1.10.1] - 2022-12-20

### Added
- Add Artifactory authentication to Jenkinsfile

## [Unreleased]

## [1.10.0] - 2022-09-12

### Changed

- cmsdev
  - Updated several dependencies to remedy [CVE-2020-26160](https://github.com/advisories/GHSA-w73w-5m7g-f7qc)
  - Build with Golang 1.18
  - Stop building RPMs for SLES 15 SP2, because they are no longer needed on the CSM version for which this RPM is intended

## [1.9.0] - 2022-09-09

### Added

- cmsdev: Add testing of 3 BOS CLI commands: "bos list", "bos v1 list", and "bos v2 list".

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

[Unreleased]: https://github.com/Cray-HPE/cms-tools/compare/1.10.7...HEAD

[1.10.7]: https://github.com/Cray-HPE/cms-tools/compare/1.10.6...1.10.7

[1.10.6]: https://github.com/Cray-HPE/cms-tools/compare/1.10.5...1.10.6

[1.10.5]: https://github.com/Cray-HPE/cms-tools/compare/1.10.4...1.10.5

[1.10.4]: https://github.com/Cray-HPE/cms-tools/compare/1.10.3...1.10.4

[1.10.3]: https://github.com/Cray-HPE/cms-tools/compare/1.10.2...1.10.3

[1.10.2]: https://github.com/Cray-HPE/cms-tools/compare/1.10.1...1.10.2

[1.10.1]: https://github.com/Cray-HPE/cms-tools/compare/1.10.0...1.10.1

[1.10.0]: https://github.com/Cray-HPE/cms-tools/compare/1.9.0...1.10.0

[1.9.0]: https://github.com/Cray-HPE/cms-tools/compare/1.8.1...1.9.0

[1.8.1]: https://github.com/Cray-HPE/cms-tools/compare/1.8.0...1.8.1

[1.8.0]: https://github.com/Cray-HPE/cms-tools/compare/1.7.0...1.8.0

[1.7.0]: https://github.com/Cray-HPE/cms-tools/compare/1.6.1...1.7.0

[1.6.1]: https://github.com/Cray-HPE/cms-tools/compare/1.6.0...1.6.1

[1.6.0]: https://github.com/Cray-HPE/cms-tools/compare/1.5.0...1.6.0

[1.5.0]: https://github.com/Cray-HPE/cms-tools/compare/1.4.0...1.5.0

[1.4.0]: https://github.com/Cray-HPE/cms-tools/compare/1.3.3...1.4.0
