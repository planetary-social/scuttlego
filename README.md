# scuttlego [![CI](https://github.com/planetary-social/scuttlego/actions/workflows/ci.yml/badge.svg)](https://github.com/planetary-social/scuttlego/actions/workflows/ci.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/planetary-social/scuttlego.svg)](https://pkg.go.dev/github.com/planetary-social/scuttlego)

A Go implementation of the [Secure Scuttlebutt][ssb] protocol. This
implementation was designed to be used by the [Planetary][planetary] client and
attempts to be efficient, stable and keep a relatively low memory footprint.

**Work in progress. The exposed interfaces and format of the persisted data may
change.**

## Features

### Supported

- Transport (handshake, box stream, RPC layer)
- Support for the default feed format
- Tracking the social graph
- Connection manager (local peers, predefined pubs)
- Replicating messages using `createHistoryStream`
- Replication scheduler (prioritise closer feeds, avoid replicating the same
  messages simultaneously from various peers etc.)
- Replicating and creating blobs

### Planned in the near future

- Connection manager (dynamic discovery of pubs from feeds)
- Handling blob wants recevied from remote peers
- Cleaning up old blobs and messages
- Rooms
- Replicating messages using EBTs

### Planned

- Private messages
- Private groups
- Support for other feed formats
- Metafeeds

## Protocol

To get an overview of the technical aspects of the Secure Scuttlebutt protocol
check out the following resources:

- [Official Secure Scuttlebutt website][ssb]
- [Scuttlebutt Protocol Guide][protocol-guide]
- [Planetary Developer Portal][planetary-developer-portal]

## Contributing

Check out our [contributing documentation](CONTRIBUTING.md).

If you find an issue, please report it on the [issue tracker][issue-tracker].

## Acknowledgements

This implementation uses [go-ssb][go-ssb] and associated libraries under the
hood. The elements which were reused are:

- the handshake mechanism
- the box stream protocol
- the verification and signing of messages
- broadcasting and receiving local UDP advertisements

[ssb]: https://scuttlebutt.nz/

[go-ssb]: https://github.com/cryptoscope/ssb

[protocol-guide]: https://ssbc.github.io/scuttlebutt-protocol-guide/

[planetary-developer-portal]: https://dev.planetary.social

[planetary]: https://www.planetary.social/

[issue-tracker]: https://github.com/planetary-social/scuttlego/issues
