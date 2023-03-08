# Changelog

## [Unreleased]

### Added 

- ...

### Changed 

- ...

### Deprecated 

- ...

### Removed 

- ...

### Fixed 

- ...

### Security 

- ...

## [v0.0.4]

### Fixed 

- Improved overall performance by using "github.com/json-iterator/go" instead
  of "encoding/json".
- Updated go-ssb to the latest version which should improve message validation
  performance.

## [v0.0.3]

### Changed 

- Moved package `di` to `service` and moved `di.Service` and `di.Config` to `service`.

### Fixed 

- Fewer goroutines are created during EBT replication.
- Optimized marshaling and unmarshaling EBT notes.
- Optimized marshaling and unmarshaling createHistoryStream arguments.

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

[unreleased]: https://github.com/planetary-social/scuttlego/compare/v0.0.4...HEAD
[v0.0.4]: https://github.com/planetary-social/scuttlego/releases/tag/v0.0.4
[v0.0.3]: https://github.com/planetary-social/scuttlego/releases/tag/v0.0.3
[v0.0.2]: https://github.com/planetary-social/scuttlego/releases/tag/v0.0.2
[v0.0.1]: https://github.com/planetary-social/scuttlego/releases/tag/v0.0.1
