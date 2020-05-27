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

Note: This execution framwork works on Windows, MacOS, and Linux (assuming
it's cross-compiled).
