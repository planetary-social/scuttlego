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
- Replicating messages using `createHistoryStream` and Epidemic Broadcast Trees
- Replication scheduler (prioritise closer feeds, avoid replicating the same
  messages simultaneously from various peers etc.)
- Replicating and creating blobs
- Tunneling via rooms
- Some commands and queries for managing room aliases

### Planned in the near future

- Connection manager (dynamic discovery of pubs from feeds)
- Handling blob wants recevied from remote peers
- Cleaning up old blobs and messages

### Planned

- Private messages
- Private groups
- Support for other feed formats
- Metafeeds

## Community

If you want to talk about scuttlego feel free to post on Secure Scuttlebutt using the `#scuttlego` channel.

Also check out Matrix channels such as `#golang-ssb-general:autonomic.zone` and `#planetary:matrix.org`.

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

This implementation calls [go-ssb][go-ssb] and associated libraries under the
hood. The elements which didn't have to be reimplemented from scratch thanks to
that are mainly:

- the handshake mechanism
- the box stream protocol
- the verification and signing of messages
- broadcasting and receiving local UDP advertisements

[ssb]: https://scuttlebutt.nz/

[go-ssb]: https://github.com/ssbc/go-ssb

[protocol-guide]: https://ssbc.github.io/scuttlebutt-protocol-guide/

[planetary-developer-portal]: https://dev.planetary.social

[planetary]: https://www.planetary.social/

[issue-tracker]: https://github.com/planetary-social/scuttlego/issues
