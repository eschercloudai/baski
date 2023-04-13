# Changelog

## [ 12/04/2023 - v0.1.0-beta.1 ]

### ADDED
* Added changelog
* Refactored code to:
  * Prevent using `viper.GetXXXX` across the codebase - now gets put into struct to allow one location to be updated rather than multiples.
  * Begin work to allow more clouds to be added - still work to be done #36.
  * Begin work to improve flags - still work to be done #33.
* Updated the config file requirements. *This is a breaking change and old configs will no longer work.*.

### Fixed
 * Trivy checksum now used to validate trivy download if required #32.
 * Added flags, which were previously missing, to support adding Trivy and Falco to the image #34.

### Deprecated
* The publish command will be reworked in an upcoming release to prevent the GitHub requirement. Instead, it will generate the files require to publish a single images scan results as an artifact with which the user can then decide what to do.

## [ Previous versions ]

* Up to this point, there has been no changelog supplied for previous versions as it was a rapid iterative process.
* With the release of v0.1.0-beta.1, any changes will be logged and should one be a breaking change, it will incur a version bump.
* Minor version bumps will be reserved for general changes
* Patch version bumps will be for fixes and patches
* The beta tags will be for superficial changes within a patch that require testing before a final release is created.
