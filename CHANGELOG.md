# Changelog

## [Unreleased]


### Added 

- ...

### Changed 

- Moved package `di` to `service` and moved `di.Service` and `di.Config` to `service`.

### Deprecated 

- ...

### Removed 

- ...

### Fixed 

- Fewer goroutines are created during EBT replication.

### Security 

- ...


## [v0.0.2]

### Changed

- Logging interfaces have changed to improve logging performance and enable
  optimizations.

### Fixed

- Improved blob replication performance by caching the list of blobs which
  should be pushed.
- Improved overall performance by optimising a hot path related to receiving
  RPC messages.

## [v0.0.1]

### Added

- This CHANGELOG file.

[unreleased]: https://github.com/planetary-social/scuttlego/compare/v0.0.2...HEAD
[v0.0.2]: https://github.com/planetary-social/scuttlego/releases/tag/v0.0.2
[v0.0.1]: https://github.com/planetary-social/scuttlego/releases/tag/v0.0.1
