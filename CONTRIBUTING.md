# Contributing

## Code tour

The project loosely follows [hexagonal architecture][hexagonal-architecture].
Following this pattern we can distinguish four general areas of this codebase.
To group them together and keep them apart from the usual noise such as
logging or test related code they were placed in the package `service`.

The **domain** layer is located in `service/domain`. In this directory you can
find the implementation of various components of the Secure Scuttlebutt
protocol. The types placed here should be well tested and implement the base
behaviours of our program. They should also be storage agnostic and shouldn't
know how (and usually if) they are persisted.

The **application** layer is located in `service/app`. Application layer
effectively acts as glue and links other layers together. For example, it may be
responsible for using an adapter to load a domain object from the database,
processing it using other domain types and then saving it. Application layer is
divided between commands and queries. Commands should modify the state of the
program while queries should not. Commands and queries are called from external
sources e.g. ports or your own program using this project.

**Ports** (primary adapters/driving adapters) can be found in `service/ports`.
Ports drive the application layer by triggering commands and queries which
reside in it. Ports handle incoming events which should trigger some kind of a
behaviour in our program. Those can be incoming TCP network connections, local
UDP broadcasts or internal events created by this implementation itself.

**Adapters** (secondary adapters/driven adapters) can be found in
`service/adapters`. Those types are injected into the application layer (and
sometimes domain layer) and are used when our program needs to interface with
external systems such as the persistence layer e.g. the database or the file
system. You will find everything related to databases, persisting data or moving
it around here.

## Go version

The project usually uses the latest Go version as declared by the `go.mod` file.
You may not be able to build it using older compilers.

## Local development

We recommend reading the `Makefile` to discover some targets which you can
execute. It can be used as a shortcut to run various useful commands.

You may have to run the following command to install a linter and a code
formatter before executing certain targets:

    $ make tools

If you want to check if the pipeline will pass for your commit it should be
enough to run the following command:

    $ make ci

It is also useful to often run just the tests during development:

    $ make test

Easily format your code with the following command:

    $ make fmt

## Writing code

Resources which are in my opinion informative and good to read:

- [Effective Go][effective-go]
- [Go Code Review Comments][code-review-comments]
- [Uber Go Style Guide][uber-style-guide]

### Naming tests

When naming tests which tests a specific behaviour it is recommended to follow a
pattern `TestNameOfType_ExpectedBehaviour`. Example:
`TestCreateHistoryStream_IfOldAndLiveAreNotSetNothingIsWrittenAndStreamIsClosed`
.

### Panicking constructors

Some constructors are prefixed with the word `Must`. Those constructors panic
and should always be accompanied by a normal constructor which isn't prefixed
with the `Must` and returns an error. The panicking constructors should only be
used in the following cases:
- when writing tests
- when a static value has to be created e.g. `MustNewHops(1)` and this branch of
  logic in the code is covered by tests


## Opening a pull request

Pull requests are verified using CI, see the previous section to find out how to
run the same checks locally. Thanks to that you won't have to push the code to
see if the pipeline passes.

It is always a good idea to try to [write a good commit message][commit-message]
and avoid bundling unrelated changes together. If your commit history is messy
and individual commits don't work by themselves it may be a good idea to squash
your changes.

### Feature branches

When naming long-lived feature branches please follow the pattern `feature/...`.
This enables CI for that branch.


[hexagonal-architecture]: https://en.wikipedia.org/wiki/Hexagonal_architecture_(software)

[commit-message]: https://cbea.ms/git-commit/

[effective-go]: http://golang.org/doc/effective_go.html
[code-review-comments]: https://github.com/golang/go/wiki/CodeReviewComments
[uber-style-guide]: https://github.com/uber-go/guide/blob/master/style.md
