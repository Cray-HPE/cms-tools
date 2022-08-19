# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- cmsdev: Add testing of BOSv2 (both API and CLI). Restored testing of CLI without explicitly specifying the BOS version.
- cmsdev: Include CLI command that failed in corresponding error messages

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

[1.0.0] - (no date)
