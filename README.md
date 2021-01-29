go-atomicredteam is a Golang application to execute tests as defined in the
atomics folder of Red Canary's Atomic Red Team project. The "atomics folder"
contains a folder for each Technique defined by the MITRE ATT&CK™ Framework.
Inside of each of these "T#" folders you'll find a yaml file that defines the
attack procedures for each atomic test as well as an easier to read markdown
(md) version of the same data.

* Executing atomic tests may leave your system in an undesirable state. You are
  responsible for understanding what a test does before executing.

* Ensure you have permission to test before you begin.

* It is recommended to set up a test machine for atomic test execution that is
  similar to the build in your environment. Be sure you have your
  collection/EDR solution in place, and that the endpoint is checking in and
  active.

## Goals

The goal of go-atomicredteam (`goart` for short) is to package all the atomic
test definitions and documentation directly into the executable such that it
can be used in environments without Internet access. As of now this is not
fool proof, however, since some tests themselves include calls to download
additional resources from the Internet.

`goart` also supports local development of tests by making it possible for a
user to dump existing test definitions to a local directory and load tests
present in the local directory when executing. If a test is both defined
locally and within the executable, the local test will be used.

## Building

To build `goart` for all operating systems, simply run `make release` from
the root directory. Otherwise, run the appropriate `make bin/goart-<os>`
target for your OS.

Obviously, Golang needs to be installed locally in order to build.

Note: This execution framwork works on Windows, MacOS, and Linux (assuming
it's cross-compiled).
