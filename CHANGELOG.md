## Unreleased

## [0.2.7](https://github.com/paralus/paralus/compare/v0.2.6...v0.2.7) (2024-02-28)

### Features

* Ability to regularly check in on the target cluster connection status ([#245](https://github.com/paralus/paralus/issues/245)) ([0cd2a35](https://github.com/paralus/paralus/commit/0cd2a35ab52b9f86d69b385ef183187c5e224cf3))

### Bug Fixes

* add custom Empty instead of google.protobuf.Empty ([#291](https://github.com/paralus/paralus/issues/291)) ([56fdc1c](https://github.com/paralus/paralus/commit/56fdc1c27bb6e66905826ee69649e1343ceeb2a7))
* migration error on Postgresql version below 14 ([#295](https://github.com/paralus/paralus/issues/295)) ([8b54b40](https://github.com/paralus/paralus/commit/8b54b4067337cd11a640af921cba333eb2f3d2a0))


## [0.2.6](https://github.com/paralus/paralus/compare/v0.2.5...v0.2.6) (2023-12-21)

### Added

* mapper for keycloak ([01d9160](https://github.com/paralus/paralus/commit/01d91606a6e08c806e6b929d205817362fd57805))

### Bug Fixes

* added yaml marshal/unmarshal for enums ([#34](https://github.com/paralus/cli/issues/34)) ([35da062](https://github.com/paralus/paralus/commit/35da06272f8abe9e4e8ed6a8806d62e65fa7eeab))

## [0.2.5](https://github.com/paralus/paralus/compare/v0.2.4...v0.2.5) (2023-09-25)

### Features

* changes to view auditlogs by project role users ([#225](https://github.com/paralus/paralus/issues/225)) ([1b7a9a1](https://github.com/paralus/paralus/commit/1b7a9a1fa32efbaa7a4c4024145adda260a96d3e))

### ⚠ BREAKING CHANGES

Prior to v0.2.4, users will not have org, partner metadata information in kratos identities which will impact audit logs screens, apply below migrations if you are upgrading paralus

update identities set metadata_public = jsonb_set(metadata_public, '{organization}', '"replace-with-your-organization-id"', true);
update identities set metadata_public = jsonb_set(metadata_public, '{partner}', '"replace-with-your-partner-id"', true);

## [0.2.4](https://github.com/paralus/paralus/compare/v0.2.3...v0.2.4) (2023-08-11)

### Bug Fixes

* change relays annotation of Cluster to paralus.dev/relays ([#227](https://github.com/paralus/paralus/issues/227)) ([749dcb4](https://github.com/paralus/paralus/commit/749dcb46d4f82341c9e2f5168ef15ac71011694e))
* cluster list API send internal error for non-exist project ([a30f80f](https://github.com/paralus/paralus/commit/a30f80f426f95327acf25ba095755fed19a566c6))
* generate fixtures for download.yaml ([#236](https://github.com/paralus/paralus/issues/236)) ([f5e2e77](https://github.com/paralus/paralus/commit/f5e2e7739d66c73803b7a231961ce5d316eb2408))
* fix for org admins to view secrets with org restrictions ([#235](https://github.com/paralus/paralus/issues/235)) ([2965dd9](https://github.com/paralus/paralus/commit/2965dd9fdf15d71f9b0fd18aa063daad505485a3))


### [0.2.3](https://github.com/paralus/paralus/compare/v0.2.2...v0.2.3) (2023-04-28)


### Bug Fixes

* incorrect number of wg.add ([#203](https://github.com/paralus/paralus/issues/203)) ([da418fd](https://github.com/paralus/paralus/commit/da418fd3d583ef3cf0a49e22b566b1ac020beb12))
* re-running admindb migration fails ([#205](https://github.com/paralus/paralus/issues/205)) ([d88c82e](https://github.com/paralus/paralus/commit/d88c82e0df7c6cb959d800ba87f5eee565f0dc31))
* remove references to admindbuser user in admindb migrations ([#200](https://github.com/paralus/paralus/issues/200)) ([e203d15](https://github.com/paralus/paralus/commit/e203d15b8f0bcd8feba189324aa9545ed637b0fc))

## [0.2.2] - 2023-03-31

## Breaking Change

- Okta JSONNet mapper configuration for SSO login got changed to support multiple groups. This may impact the existing Okta user logins configured with paralus versions prior to v0.2.1. As a workaround use [pinned Okta mapper URL](https://raw.githubusercontent.com/paralus/paralus/v0.2.1/_kratos/oidc-mappers/okta.jsonnet) to your existing Okta OIdC configuration.

## Added
- Support more than one IdP groups mapping from [akshay196](https://github.com/akshay196)

## Fixed
- Add project name validation [hiteshwani29](https://github.com/hiteshwani29)
- Fix to error out multiple bootstrap agent registration requests from [niravparikh05](https://github.com/niravparikh05)
- Updated Documentation for APIs [mabhi](https://github.com/mabhi) and [niravparikh05](https://github.com/niravparikh05)

## [0.2.1] - 2023-02-24
### Added
-  Configure the service account lifetime from [mabhi](https://github.com/mabhi)

## Fixed
- User should not be able to delete project with clusters in it from [mabhi](https://github.com/mabhi)
- Namespace limitation input on roles [mabhi](https://github.com/mabhi)

## [0.2.0] - 2023-01-27

## Fixed
- Fix project id is recorded as part of cluster related auditlogs from [niravparikh05](https://github.com/niravparikh05)

## Added
- Enhance: Ability to set auto generated password during installation and force reset during first login [mabhi](https://github.com/mabhi)

## Changed
- Upgraded Ory Kratos to v0.10.1 [akshay196](https://github.com/akshay196)

## [0.1.9] - 2022-12-29

## Added
- Enhance: record user.login event via kratos hooks [mabhi](https://github.com/mabhi)

## Fixed
- Fix modify userinfo service to include scope in response [mabhi](https://github.com/mabhi)
- Kubectl commands work even after deleting an imported cluster from [mabhi](https://github.com/mabhi) and [niravparikh05](https://github.com/niravparikh05)

## [0.1.8] - 2022-11-25

## Added

- Added database auditlog storage option [niravparikh05](https://github.com/niravparikh05)

## [0.1.7] - 2022-11-04

## Added

- Added last login field to user spec from [akshay196](https://github.com/akshay196)

## [0.1.6] - 2022-10-14

## Fixed

- Fixed creating project scoped role failed from cli [niravparikh05](https://github.com/niravparikh05)

## [0.1.5] - 2022-10-10

## Fixed

- Fixed issue where relay server is not coming up in arm64 (Mac M1) from [niravparikh05](https://github.com/niravparikh05)

## [0.1.4] - 2022-09-30

## Fixed

- Fixed issue where relay server is not coming up in arm64 (Mac M1) from [sandeep540](https://github.com/sandeep540)
- Fixed cluster lister and set group created at property [niravparikh05](https://github.com/niravparikh05)

## [0.1.3] - 2022-08-26

## Added

- Added more audit points for better visibility from [vivekhiwarkar](https://github.com/vivekhiwarkar)
- Added audit point for kubeconfig download [meain](https://github.com/meain)

## Fixed

- Fixed lint issues due to buf from [vivekhiwarkar](https://github.com/vivekhiwarkar)

## [0.1.2] - 2022-08-12

## Fixed
- Fixed init failing with db validation error from [meain](https://github.com/meain)

## [0.1.1] - 2022-08-09

### Fixed
- Fix to validate bare minimum role permissions for custom roles from [niravparikh05](https://github.com/niravparikh05)

## [0.1.0] - 2022-06-22
### Added
- Initial release

[Unreleased]: https://github.com/paralus/paralus/compare/v0.2.7...HEAD
[0.2.7]: https://github.com/paralus/paralus/compare/v0.2.6...v0.2.7
[0.2.6]: https://github.com/paralus/paralus/compare/v0.2.5...v0.2.6
[0.2.5]: https://github.com/paralus/paralus/compare/v0.2.4...v0.2.5
[0.2.4]: https://github.com/paralus/paralus/compare/v0.2.3...v0.2.4
[0.2.3]: https://github.com/paralus/paralus/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/paralus/paralus/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/paralus/paralus/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/paralus/paralus/compare/v0.1.9...v0.2.0
[0.1.9]: https://github.com/paralus/paralus/compare/v0.1.8...v0.1.9
[0.1.8]: https://github.com/paralus/paralus/compare/v0.1.7...v0.1.8
[0.1.7]: https://github.com/paralus/paralus/compare/v0.1.6...v0.1.7
[0.1.6]: https://github.com/paralus/paralus/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/paralus/paralus/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/paralus/paralus/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/paralus/paralus/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/paralus/paralus/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/paralus/paralus/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/paralus/paralus/releases/tag/v0.1.0
