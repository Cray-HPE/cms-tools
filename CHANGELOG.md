# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.32.0] - 2025-05-20

### Added
- CASMCMS-9349: cmsdev CFS Add create/modify/delete tests
   * API tests for sources and configurations
   * CLI tests for sources and configurations

### Fixed
- CASMTRIAGE-8181: cmsdev test cfs fails intermittently in the pipeline, but passes on manual execution

## [1.31.0] - 2025-05-14

### Added
- CASMCMS-9348: cmsdev BOS Add create/modify/delete tests
   * API tests for sessiontemplates and sessions
   * CLI tests for sessiontemplates and sessions
- CASMCMS-9409: IMS cli test should suppress expected failures from log

## [1.30.0] - 2025-04-29

### Added
- CASMCMS-9379: TESTS: IMS: cmsdev: Test v2 API and default version API

### Fixed
- CASMTRIAGE-8135: IMS: cmsdev failed during post-install health check
- CASMCMS-9384: cmsdev should warn if docs-csm RPM is not installed

## [1.29.0] - 2025-04-24

### Added
- CASMCMS-9350: IMS Add create/modify/delete tests for images, recipes and public-keys
- CASMCMS-9377: IMS: cmsdev: Verify delete operations
- CASMCMS-9378: IMS: cmsdev: Verify create operations through listing

## [1.28.0] - 2025-03-31

### Changed
- Build `cmsdev` using Golang 1.23 (up from 1.20)

### Dependencies
- Bump `dangoslen/dependabot-changelog-helper` from 3 to 4 ([#243](https://github.com/Cray-HPE/cms-tools/pull/243))
- Bump `github.com/go-openapi/jsonpointer` from 0.21.0 to 0.21.1 ([#245](https://github.com/Cray-HPE/cms-tools/pull/245))
- Bump `google.golang.org/protobuf` from 1.35.1 to 1.35.2 ([#242](https://github.com/Cray-HPE/cms-tools/pull/242))
- Bump `golang.org/x/net` from 0.33.0 to 0.36.0 ([#244](https://github.com/Cray-HPE/cms-tools/pull/244))

### Fixed
- CASMCMS-9181: cmsdev Should record installed version of Cray CLI RPM
- CASMCMS-9328: IMS: Store Artifact logs after signing key test failure
- CASMCMS-8703: TESTS: Add CFS node personalization to the barebones image boot test

## [1.27.0] - 2025-02-03

### Dependencies
- Bump `github.com/google/gnostic-models` from 0.6.9-0.20230804172637-c7be7c783f49 to 0.6.9 ([#233](https://github.com/Cray-HPE/cms-tools/pull/233))
- Bump `github.com/mattn/go-colorable` from 0.1.13 to 0.1.14 ([#235](https://github.com/Cray-HPE/cms-tools/pull/235))
- Bump `github.com/spf13/pflag` from 1.0.5 to 1.0.6 ([#236](https://github.com/Cray-HPE/cms-tools/pull/236))
- Bump `sigs.k8s.io/structured-merge-diff/v4` from 4.4.2 to 4.4.3 ([#232](https://github.com/Cray-HPE/cms-tools/pull/232))
- Bump `github.com/magiconair/properties` from 1.8.7 to 1.8.9 ([#234](https://github.com/Cray-HPE/cms-tools/pull/234))

## [1.26.0] - 2025-02-03

### Dependencies
- Bump `golang.org/x/net` from 0.23.0 to 0.33.0 ([#237](https://github.com/Cray-HPE/cms-tools/pull/237))

## [1.25.0] - 2024-11-06

### Dependencies
- Bump `sigs.k8s.io/structured-merge-diff/v4` from 4.4.1 to 4.4.2 ([#228](https://github.com/Cray-HPE/cms-tools/pull/228))

### Fixed
- cmsdev: Update CFS and IMS tests to explicitly check for 0-length resource ID fields

## [1.24.1] - 2024-10-25

### Fixed
- cmsdev: BOS test should pass if there are no cray-bos-migration pods, or if at least one
  of them Succeeded. In other words, it should only fail if there are one or more
  cray-bos-migration pods, and none of them Succeeded. This is because during an upgrade,
  the initial migration pods may fail because the BOS database is not yet ready. As long
  as the migration job eventually succeeded, we do not care about the initial failures.

## [1.24.0] - 2024-09-03

### Changed
- cmsdev: BOS test should expect cray-bos-migration pod to be Succeeded, not Running

## [1.23.1] - 2024-08-27

### Added
- barebones image test: Improved logging of API calls

### Fixed
- barebones image test: Remove `rootfs_provider` and `rootfs_provider_passthrough` fields from
  session template that is created. They are set to empty strings, and this is not legal under
  the now-enforced (starting in CSM 1.6) API spec.

## [1.23.0] - 2024-08-26

### Dependencies
- barebones image test: Use `requests-retry-session` Python module instead of duplicating its code
- barebones image test: Pin major/minor but take latest patch version
- barebones image test: CSM 1.6 now uses Kubernetes 1.24, so use Python client v24.x
- cmsdev: CSM 1.6 now uses Kubernetes 1.24, so change dependabot `k8s.io` package restriction to match
- Bump `k8s.io/api` from 0.22.17 to 0.24.17 ([#191](https://github.com/Cray-HPE/cms-tools/pull/191))
- Bump `k8s.io/apimachinery` from 0.22.17 to 0.24.17 ([#191](https://github.com/Cray-HPE/cms-tools/pull/191))
- Bump `k8s.io/client-go` from 0.22.17 to 0.24.17 ([#191](https://github.com/Cray-HPE/cms-tools/pull/191))
- Bump `k8s.io/klog/v2` from 2.9.0 to 2.60.1 ([#191](https://github.com/Cray-HPE/cms-tools/pull/191))
- Bump `k8s.io/utils` from 0.0.0-20211116205334-6203023598ed to 0.0.0-20220210201930-3a6ce19ff2f9 ([#191](https://github.com/Cray-HPE/cms-tools/pull/191))
- Bump `github.com/emicklei/go-restful` from 2.9.5+incompatible to 2.16.0+incompatible ([#207](https://github.com/Cray-HPE/cms-tools/pull/207))
- Bump `github.com/mailru/easyjson` from 0.7.6 to 0.7.7 ([#208](https://github.com/Cray-HPE/cms-tools/pull/208))
- Bump `sigs.k8s.io/yaml` from 1.2.0 to 1.4.0 ([#213](https://github.com/Cray-HPE/cms-tools/pull/213))
- Bump `sigs.k8s.io/structured-merge-diff/v4` from 4.2.3 to 4.4.1 ([#202](https://github.com/Cray-HPE/cms-tools/pull/202))
- Bump `github.com/fatih/color` from 1.15.0 to 1.17.0 ([#214](https://github.com/Cray-HPE/cms-tools/pull/214))
- Bump `github.com/PuerkitoBio/purell` from 1.1.1 to 1.2.1 ([#215](https://github.com/Cray-HPE/cms-tools/pull/215))
- Bump `github.com/subosito/gotenv` from 1.4.2 to 1.6.0 ([#218](https://github.com/Cray-HPE/cms-tools/pull/218))
- Bump `github.com/go-openapi/jsonpointer` from 0.19.5 to 0.21.0 ([#217](https://github.com/Cray-HPE/cms-tools/pull/217))
- Bump `github.com/go-openapi/jsonreference` from 0.19.5 to 0.21.0 ([#212](https://github.com/Cray-HPE/cms-tools/pull/212))
- Bump `github.com/google/gnostic` from 0.5.7-v3refs to 0.7.0 ([#211](https://github.com/Cray-HPE/cms-tools/pull/211))
- Bump `github.com/spf13/afero` from 1.9.5 to 1.11.0 ([#196](https://github.com/Cray-HPE/cms-tools/pull/196))
- Bump `github.com/fsnotify/fsnotify` from 1.6.0 to 1.7.0 ([#197](https://github.com/Cray-HPE/cms-tools/pull/197))
- Bump `github.com/google/gofuzz` from 1.1.0 to 1.2.0 ([#201](https://github.com/Cray-HPE/cms-tools/pull/201))
- Bump `github.com/go-logr/logr` from 1.2.0 to 1.2.4 ([#219](https://github.com/Cray-HPE/cms-tools/pull/219))

## [1.22.0] - 2024-06-20

### Changed
- Compile `cmsdev` with verbose flag
- Modify VCS test to specify git credentials using environment variables rather than URL
- Look up Kubernetes secrets using golang module instead of `kubectl` command
- Do not log VCS credentials
- When installing cmsdev RPM, remove from the log any previously-logged credentials
- Do not log credentials when getting API token

### Dependencies
- Bump `golang.org/x/net` from 0.17.0 to 0.23.0 ([#183](https://github.com/Cray-HPE/cms-tools/pull/183))

## [1.21.0] - 2024-04-03

### Added/Changed
- `cmsdev`: Overhauled CFS API and CLI checks, primarily to prevent them from failing with the addition of pagination
  in CFS v3. However, prior to this, the test only covered CFS v2. This modifies it to include CFS v3. In addition, it
  adds coverage to some overlooked endpoints.

### Removed
- `cmsdev` CFS test no longer checks status of components in CFS. Customers use SAT for this already, and this really only
  served to make people think there were CFS errors just because a CFS configuration session failed on a node.

## [1.20.0] - 2024-03-15

### Changed
- Remove BOS-v1-specific checks from `cmsdev` BOS test.

### Dependencies
- Bump `k8s.io/api` from 0.22.13 to 0.22.17 ([#176](https://github.com/Cray-HPE/cms-tools/pull/176))
- Bump `k8s.io/apimachinery` from 0.22.13 to 0.22.17 ([#176](https://github.com/Cray-HPE/cms-tools/pull/176))
- Bump `k8s.io/client-go` from 0.22.13 to 0.22.17 ([#176](https://github.com/Cray-HPE/cms-tools/pull/176))
- Bump `github.com/golang/protobuf` from 1.5.3 to 1.5.4 ([#174](https://github.com/Cray-HPE/cms-tools/pull/174))

## [1.19.1] - 2024-02-23

### Changed
- Added a step in the Jenkinsfile to test install the RPM after building it, to validate
  there are no obvious problems.

### Fixed
- Fixed a bug causing the RPM Python requirements to be filled with invalid data.

## [1.19.0] - 2024-02-12

### Changed
- If only building for a single Python version, use a simple symbolic link to run the barebones
  test, instead of the `run_barebones_image_test.sh` script.
- Build only for Python 3.11

### Fixed
- Fixed a couple bugs in the barebones test S3 module.

## [1.18.1] - 2024-02-08

### Changed
- Update to cray-tftp-upload to work with more than one ipxe pod and fixed return error.

## [1.18.0] - 2024-02-08

### Changed
- Make RPM spec file more precise in its requirements (`python3-base` should not simply have a minimum requirement, but
  also a maximum, to ensure the correct Python version is on the system)
- Combine multiple Python versions in single RPM, allowing RPM to go back to being `noos`.

## [1.17.0] - 2024-02-01

### Changed
- Barebones boot test overhaul
  - Use BOSv2
  - Add support for arm nodes and images
  - By default, select node and image with x86_64 hardware
  - Default image is the compute image listed in the Cray Product Catalog for the chosen architecture
  - Default image is customized using CFS to make it fully bootable by BOS v2
  - Add options to allow user to specify how the image is customized
  - Add option to allow the user to specify a different base image, or a pre-customized image
  - If test passes, delete resources that it created at the end.
  - Add option to not delete the resources even if the test passes.
  - Package test in a Python virtual environment in order to manage and control its dependencies
- The previous change also entailed once again building the RPM as OS-specific rather than `noos`

### Removed
- Remove BOSv1 from cmsdev

### Dependencies
- Bump `github/codeql-action` from 2 to 3 ([#156](https://github.com/Cray-HPE/cms-tools/pull/156))

## [1.16.0] - 2023-11-30

### Added

- cmsdev: Added `--no-cleanup` option to prevent deleting of temporary test files, to help in debug.

### Dependencies

- Bump `stefanzweifel/git-auto-commit-action` from 4 to 5 ([#149](https://github.com/Cray-HPE/cms-tools/pull/149))
- Bump `google.golang.org/appengine` from 1.6.7 to 1.6.8 ([#151](https://github.com/Cray-HPE/cms-tools/pull/151))
- Bump `github.com/mattn/go-isatty` from 0.0.19 to 0.0.20 ([#152](https://github.com/Cray-HPE/cms-tools/pull/152))

## [1.15.0] - 2023-09-28

### Changed

- cray-upload-recovery-images: Added chmod to make sure files are world readable after upload.

### Dependencies
- Bump `actions/checkout` from 3 to 4 ([#145](https://github.com/Cray-HPE/cms-tools/pull/145))

## [1.14.1] - 2023-08-16

### Added

- cmsdev: Added good path BOS CLI list/describe tests with tenant specified for supported v2 endpoints.

## [1.14.0] - 2023-08-14

### Changed

- cmsdev
  - Simplified `lib.common.Restful()` function
  - Added good path BOS API GET tests with tenant specified for supported v2 endpoints
  - Updated v2 sessions CLI, session templates CLI (v1 and v2), and v1 session templates API tests to
    handle multi-tenancy in their responses from BOS (while not including it in their queries to BOS).

## [1.13.0] - 2023-08-10

### Changed

- cmsdev
  - Build using Golang 1.20 (up from 1.18)
  - Build RPM as `noos`

### Dependencies
- Bump `github.com/sirupsen/logrus` from 1.8.1 to 1.9.3 (#123)
- Bump `github.com/fatih/color` from 1.12.0 to 1.15.0 (#121)
- Bump `k8s.io/api`, `k8s.io/apimachinery`, and `k8s.io/client-go` from 0.21.14 to 0.22.13 (#127)
- Bump `github.com/spf13/cobra` from 1.2.1 to 1.7.0 (#122)
- Bump `github.com/spf13/viper` from 1.8.1 to 1.16.0 (#119)
- Bump `github.com/pelletier/go-toml/v2` from 2.0.8 to 2.0.9 ([#129](https://github.com/Cray-HPE/cms-tools/pull/129))
- Bump `github.com/imdario/mergo` from 0.3.5 to 0.3.16 ([#131](https://github.com/Cray-HPE/cms-tools/pull/131))
- Bump `sigs.k8s.io/structured-merge-diff/v4` from 4.2.1 to 4.2.3 ([#128](https://github.com/Cray-HPE/cms-tools/pull/128))
- Bump `github.com/mattn/go-isatty` from 0.0.17 to 0.0.19 ([#130](https://github.com/Cray-HPE/cms-tools/pull/130))

## [1.12.0] - 2023-06-27

### Added

- cmsdev: Added support for ARM binaries to iPXE/TFTP test

## [1.11.13] - 2023-06-27

### Fixed

- cmsdev: Compress artifacts using gzip instead of bzip2

## [1.11.12] - 2023-06-21

### Added

- Support for SLES SP5

## [1.11.11] - 2023-05-09

### Changed

- cmsdev: iPXE/TFTP test improvements:
  - Now uses cray-ipxe-settings ConfigMap to determine which iPXE binaries are being built and what their names are
  - Now tests TFTP file transfer test for all binaries being built.

## [1.11.10] - 2023-04-27

### Changed

- cmsdev: Add list of service tests to --help output.

## [1.11.9] - 2023-04-07

### Changed

- cmsdev: Changed default log directory to `/opt/cray/tests/install/logs/cmsdev/` to be consistent with other CSM tests.
- cmsdev: Simplified Kubernetes artifact collection functions; collect additional information
- cmsdev: Collect Kubernetes artifacts on failure by default
- cmsdev: Compress test artifacts, if collected. If none collected, delete empty artifact directory, if cmsdev created it

### Removed

- cmsdev: Removed reference to long removed `cmslogs` tool from [`cmsdev` README file](cmsdev/README.md)

## [1.11.8] - 2023-04-06

### Changed

- cmsdev: Limit redundant logging and output related to KUBECONFIG environment variable
- cmsdev: Update VCS test to reflect change to logical DB backups in CSM 1.5
- cmsdev: Modify BOS test to handle change to Bitnami for etcd; make pod checks more flexible

## [1.11.7] - 2023-04-04

### Changed

- cmsdev: Bump golang.org/x/sys to 0.1.0
  - Resolves CVE-2022-29526
- cmsdev: Bump golang.org/x/net to 0.7.0
  - Resolves CVE-2022-41723

## [1.11.6] - 2023-04-03

### Changed

- cmsdev: Fixed typo in Kubernetes pod check function.
- cmsdev: Added namespace detail to output of some Kubernetes check functions.

## [1.11.5] - 2023-03-30

### Changed

- cmsdev: Update default CLI BOS version to v2

## [1.11.4] - 2023-03-27

### Removed

- cmsdev: Remove the long unmaintained and unused non-test code paths.

## [1.11.3] - 2023-03-22

### Removed

- cmsdev: Stop building SP3 RPM. This is only intended to run on NCNs of CSM 1.5 or higher, and those
  will be running at least SP4.
- cmsdev: Remove cmsdev binary which somehow made it into the repository.

## [1.11.2] - 2023-03-22

### Added

- cmsdev: Add -l option to "tests" to list possible tests to run. --exclude-alises may be specified
  to exclude aliases from this listing.

### Changed

- cmsdev: Fixed null pointer exception when run before CSM is deployed.

## [1.11.1] - 2023-03-16

### Changed

- cmsdev: Cosmetic changes made to appease latest gofmt version
- cmsdev: Do not run tftp file transfer test from master NCNs

## [1.11.0] - 2023-01-30

### Removed

- cmsdev: Removed CRUS component, to reflect its removal in CSM 1.5.

## [1.10.1] - 2022-12-20

### Added

- Add Artifactory authentication to Jenkinsfile

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

[Unreleased]: https://github.com/Cray-HPE/cms-tools/compare/1.27.0...HEAD

[1.27.0]: https://github.com/Cray-HPE/cms-tools/compare/1.26.0...1.27.0

[1.26.0]: https://github.com/Cray-HPE/cms-tools/compare/1.25.0...1.26.0

[1.25.0]: https://github.com/Cray-HPE/cms-tools/compare/1.24.1...1.25.0

[1.24.1]: https://github.com/Cray-HPE/cms-tools/compare/1.24.0...1.24.1

[1.24.0]: https://github.com/Cray-HPE/cms-tools/compare/1.23.1...1.24.0

[1.23.1]: https://github.com/Cray-HPE/cms-tools/compare/1.23.0...1.23.1

[1.23.0]: https://github.com/Cray-HPE/cms-tools/compare/1.22.0...1.23.0

[1.22.0]: https://github.com/Cray-HPE/cms-tools/compare/1.21.0...1.22.0

[1.21.0]: https://github.com/Cray-HPE/cms-tools/compare/1.20.0...1.21.0

[1.20.0]: https://github.com/Cray-HPE/cms-tools/compare/1.19.1...1.20.0

[1.19.1]: https://github.com/Cray-HPE/cms-tools/compare/1.19.0...1.19.1

[1.19.0]: https://github.com/Cray-HPE/cms-tools/compare/1.18.1...1.19.0

[1.18.1]: https://github.com/Cray-HPE/cms-tools/compare/1.18.0...1.18.1

[1.18.0]: https://github.com/Cray-HPE/cms-tools/compare/1.17.0...1.18.0

[1.17.0]: https://github.com/Cray-HPE/cms-tools/compare/1.16.0...1.17.0

[1.16.0]: https://github.com/Cray-HPE/cms-tools/compare/1.15.0...1.16.0

[1.15.0]: https://github.com/Cray-HPE/cms-tools/compare/1.14.1...1.15.0

[1.14.1]: https://github.com/Cray-HPE/cms-tools/compare/1.14.0...1.14.1

[1.14.0]: https://github.com/Cray-HPE/cms-tools/compare/1.13.0...1.14.0

[1.13.0]: https://github.com/Cray-HPE/cms-tools/compare/1.12.0...1.13.0

[1.12.0]: https://github.com/Cray-HPE/cms-tools/compare/1.11.13...1.12.0

[1.11.13]: https://github.com/Cray-HPE/cms-tools/compare/1.11.12...1.11.13

[1.11.12]: https://github.com/Cray-HPE/cms-tools/compare/1.11.11...1.11.12

[1.11.11]: https://github.com/Cray-HPE/cms-tools/compare/1.11.10...1.11.11

[1.11.10]: https://github.com/Cray-HPE/cms-tools/compare/1.11.9...1.11.10

[1.11.9]: https://github.com/Cray-HPE/cms-tools/compare/1.11.8...1.11.9

[1.11.8]: https://github.com/Cray-HPE/cms-tools/compare/1.11.7...1.11.8

[1.11.7]: https://github.com/Cray-HPE/cms-tools/compare/1.11.6...1.11.7

[1.11.6]: https://github.com/Cray-HPE/cms-tools/compare/1.11.5...1.11.6

[1.11.5]: https://github.com/Cray-HPE/cms-tools/compare/1.11.4...1.11.5

[1.11.4]: https://github.com/Cray-HPE/cms-tools/compare/1.11.3...1.11.4

[1.11.3]: https://github.com/Cray-HPE/cms-tools/compare/1.11.2...1.11.3

[1.11.2]: https://github.com/Cray-HPE/cms-tools/compare/1.11.1...1.11.2

[1.11.1]: https://github.com/Cray-HPE/cms-tools/compare/1.11.0...1.11.1

[1.11.0]: https://github.com/Cray-HPE/cms-tools/compare/1.10.1...1.11.0

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
