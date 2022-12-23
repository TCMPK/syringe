# Name. (required)
name: ${T_ARTIFACT_NAME}

# Platform.
# This is only used by the rpm packager.
# Examples: `linux` (default), `darwin`
platform: linux

# Version. (required)
# This will expand any env var you set in the field, e.g. version: ${SEMVER}
# Some package managers, like deb, require the version to start with a digit.
# Hence, you should not prefix the version with 'v'.
version: ${T_ARTIFACT_VERSION}

# Version Schema allows you to specify how to parse the version string.
# Default is `semver`
#   `semver` attempt to parse the version string as a valid semver version.
#       The parser is lenient; it will strip a `v` prefix and will accept
#       versions with fewer than 3 components, like `v1.2`.
#       If parsing succeeds, then the version will be molded into a format
#       compatible with the specific packager used.
#       If parsing fails, then the version is used as-is.
#   `none` skip trying to parse the version string and just use what is passed in
version_schema: semver

# Version Epoch.
# A package with a higher version epoch will always be considered newer.
# See: https://www.debian.org/doc/debian-policy/ch-controlfields.html#epochs-should-be-used-sparingly
epoch: 2

# Version Prerelease.
# Default is extracted from `version` if it is semver compatible.
# This is appended to the `version`, e.g. `1.2.3+beta1`. If the `version` is
# semver compatible, then this replaces the prerelease component of the semver.
# prerelease: beta1

# Version Metadata (previously deb.metadata).
# Default is extracted from `version` if it is semver compatible.
# Setting metadata might interfere with version comparisons depending on the
# packager. If the `version` is semver compatible, then this replaces the
# version metadata component of the semver.
version_metadata: git

# Section.
# This is only used by the deb packager.
# See: https://www.debian.org/doc/debian-policy/ch-archive.html#sections
section: net

# Maintainer. (required)
# This will expand any env var you set in the field, e.g. maintainer: ${GIT_COMMITTER_NAME} <${GIT_COMMITTER_EMAIL}>
# Defaults to empty on rpm and apk
# Leaving the 'maintainer' field unset will not be allowed in a future version
maintainer: Peter Klein <peter@tcmpk.de>

# Description.
# Defaults to `no description given`.
# Most packagers call for a one-line synopsis of the package. Some (like deb)
# also call for a multi-line description starting on the second line.
description: The syringe dns-preheating daemon

# Vendor.
# This will expand any env var you set in the field, e.g. vendor: ${VENDOR}
# This is only used by the rpm packager.
vendor: GoReleaser

# Package's homepage.
homepage: https://github.com/TCMPK/syringe

# License.
license: Apache License, Version 2.0

# Changelog YAML file, see: https://github.com/goreleaser/chglog
# changelog: "changelog.yaml"

# Disables globbing for files, config_files, etc.
disable_globbing: false

# Contents to add to the package
# This can be binaries or any other files.
contents:
  - src: ${T_ARTIFACT_NAME}
    dst: /usr/local/bin/syringe

  - dst: /etc/syringe
    type: dir
    file_info:
      mode: 0700

  - src: .github/packaging/systemd/syringe.service
    dst: /usr/lib/systemd/system/syringe@.service

  - src: .github/packaging/domains
    dst: /etc/syringe/domains
    type: config|noreplace

  - src: syringe.yml
    dst: /etc/syringe/127.0.0.1.yml
    type: config|noreplace

  - src: syringe.yml
    dst: /etc/syringe/syringe.sample.yml

# Scripts to run at specific stages. (overridable)
scripts:
  postinstall: .github/packaging/post-install.sh
  preremove: .github/packaging/post-remove.sh
  postremove: .github/packaging/pre-remove.sh