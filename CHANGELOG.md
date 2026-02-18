# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.1](https://github.com/manuelarte/pagorminator/compare/v0.2.0...v0.2.1) (2026-02-18)


### Bug Fixes

* expose GormString ([#84](https://github.com/manuelarte/pagorminator/issues/84)) ([9634450](https://github.com/manuelarte/pagorminator/commit/96344502033353889d5ffa23df0872d029026c6b))

## [0.2.0](https://github.com/manuelarte/pagorminator/compare/v0.1.0...v0.2.0) (2026-02-18)


### ⚠ BREAKING CHANGES

* **sort:** changing sorting ([#79](https://github.com/manuelarte/pagorminator/issues/79))

### Miscellaneous Chores

* release v0.2.0 ([#82](https://github.com/manuelarte/pagorminator/issues/82)) ([bd31bc9](https://github.com/manuelarte/pagorminator/commit/bd31bc96ae5e542b9ac626b421b8391d28f7f9a6))


### Code Refactoring

* **sort:** changing sorting ([#79](https://github.com/manuelarte/pagorminator/issues/79)) ([c81a761](https://github.com/manuelarte/pagorminator/commit/c81a7613221b4d6c8433272d6672cf495e0a91f0))

## [v0.1.0](https://github.com/manuelarte/pagorminator/compare/v0.0.5...v0.1.0) (2025-07-16)


### Features

* *BREAKING CHANGE* renaming to PaGorminator ([#57](https://github.com/manuelarte/pagorminator/issues/57)) ([032e552](https://github.com/manuelarte/pagorminator/commit/032e55231c6ab11dc2ff8cac9b32a63d2c91d592))

## [v0.0.5](https://github.com/manuelarte/pagorminator/compare/v0.0.4...v0.0.5) (2025-07-01)


### Bug Fixes

* supporting join queries ([#51](https://github.com/manuelarte/pagorminator/issues/51)) ([da7f99d](https://github.com/manuelarte/pagorminator/commit/da7f99df515812642b41f778f06b4c87cb3f00a9))

## [v0.0.4](https://github.com/manuelarte/pagorminator/compare/v0.0.3...v0.0.4) (2025-05-26)


### Added

* Supporting `Distinct()`.

## [v0.0.3](https://github.com/manuelarte/pagorminator/compare/v0.0.2...v0.0.3) (2025-05-23)


### Added

* Added the method `SetTotalElements` giving the ability to set the `totalElements` value.
* Supporting setting the model through `Table()`.

## [v0.0.1] (2025-04-15)


### Added

* Added PaGorminator plugin for gorm
  - size & page
  - sorting
* Added support for sorting
* Added examples
