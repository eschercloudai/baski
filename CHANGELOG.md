# Changelog

## [ Upcoming Release - v0.1.0-beta.9 ]

### Changed/Added
* Added a server so baski can serve up scans via an API
* Added initial tests
* Updated builds to include server as a separate binary and docker image
* Updated golang version in pipeline

### Fixed

### Deprecated

## [ 12/07/2023 - v0.1.0-beta.8 ]

### Changed/Added

* Added additional scan command to enable scanning multiple images.
* Changed log.fatals to returns so that RunE can handle the error.

### Fixed
* Fixed code to support new repo changes in kubernetes.

### Deprecated

* Removed references to publish command so that it can no longer be called - code will be removed in coming release.

## [ 12/07/2023 - v0.1.0-beta.7 ]

### Changed/Added

* Added ability to pass in a list of container images to bake in.

### Fixed

* Switched out panics for logging errors.
* Corrected names in GitHub actions.
* fixed date tag on image as it was just setting RFC3339 rather than using it as the format.

## [ 12/07/2023 - v0.1.0-beta.6 ]

### Changed/Added

* Added ability to pass in a list of container images to bake in.

### Fixed

* Switched out panics for logging errors.
* Corrected names in GitHub actions.
* fixed date tag on image as it was just setting RFC3339 rather than using it as the format.

## [ 16/05/2023 - v0.1.0-beta.5 ]

### Changed/Added

* Enabled support for S3 backends when using S3 buckets.

### Fixed

* Build command was missing some flags - these have been added.

## [ 2023/05/15 - v0.1.0-beta.4 ]

### Added

* Support for trivyignore and adding lists of CVEs to ignore.

## [ 2023/05/09 - v0.1.0-beta.3 ]

### Fixed

* Ensured FIP creation and removal rather than just looking for one in the account to prevent race condition when
  attaching an IP.

## [ 2023/04/28 - v0.1.0-beta.2 ]

### Fixed

* Fixed Nvidia and security inclusions.

## [ 2023/04/12 - v0.1.0-beta.1 ]

### Changed/Added

* Added changelog
* Refactored code to:
    * Prevent using `viper.GetXXXX` across the codebase - now gets put into struct to allow one location to be updated
      rather than multiples.
    * Begin work to allow more clouds to be added - still work to be done #36.
    * Begin work to improve flags - still work to be done #33.
* Updated the config file requirements. *This is a breaking change and old configs will no longer work.*.

### Fixed

* Trivy checksum now used to validate trivy download if required #32.
* Added flags, which were previously missing, to support adding Trivy and Falco to the image #34.

### Deprecated

* The publish command will be reworked in an upcoming release to prevent the GitHub requirement. Instead, it will
  generate the files require to publish a single images scan results as an artifact with which the user can then decide
  what to do.

## [ Previous versions ]

* Up to this point, there has been no changelog supplied for previous versions as it was a rapid iterative process.
* With the release of v0.1.0-beta.1, any changes will be logged and should one be a breaking change, it will incur a
  version bump.
* Minor version bumps will be reserved for general changes.
* Patch version bumps will be for fixes and patches.
* The beta tags will be for superficial changes within a patch that require testing before a final release is created.
