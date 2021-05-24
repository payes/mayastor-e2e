# Mayastor E2E fio test pod
## Introduction
Derived from `dmonakhov/alpine-fio`


### Arguments
 * `sleep <sleep seconds>`
 * `segfault-after <delay seconds>`
 * `exitv <exit value>`
 * `-- <fio args 1> -- <fio args 2> .... -- <fio args N>`
   * a single instance of `fio` is launched for each set of arguments.

 1. `fio` is only run if `fio` arguments are specified.
 2. all `fio` instances are run as a forked processes.
 3. the segfault directive takes priority over the sleep directive
 4. `exitv <v>` override exit value - this is to simulate failure.
 5. argument parsing is simple, invalid specifications are skipped over for example `"sleep --"` => `sleep` is skipped over, parsing resumes from `--`, execution does not fail.

### Exit value
* If `exitv` is specified that is *always* returned.
* If all instances of `fio` ran successfully exit value is 0
* If a single instance of `fio` fails the exit value is the exit value of the failing instance of `fio`
* If multiple instances of `fio` fail the exit value is the exit value of a failing instance of `fio`

## building
Run `./build.sh`

This builds the image `mayadata/e2e-fio`


