# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.5](https://github.com/manuelarte/pagorminator/compare/v0.0.4...v0.0.5) (2025-07-01)


### Bug Fixes

* supporting join queries ([#51](https://github.com/manuelarte/pagorminator/issues/51)) ([da7f99d](https://github.com/manuelarte/pagorminator/commit/da7f99df515812642b41f778f06b4c87cb3f00a9))

## [v0.0.4] 2025-05-26

### Added

- Supporting `Distinct()`.

## [v0.0.3] 2025-05-23

### Added

- Added the method `SetTotalElements` giving the ability to set the `totalElements` value.
- Supporting setting the model through `Table()`.

## [v0.0.1] 2025-04-15

### Added

- Added PaGorminator plugin for gorm
  - size & page
  - sorting

## [v0.0.1-rc5] 2025-02-19

### Added

- Added support for sorting

## [v0.0.1-rc4] 2024-12-24

### BugFix

- Fixing bug of using Pagination with preload

## [v0.0.1-rc3] 2024-12-17

### BugFix

- Fixing bug for page 0 and page 1 returning same result

## [v0.0.1-rc2] 2024-11-30

### Added

- Added PaGorminator plugin for gorm
- Added examples
- Added badges
