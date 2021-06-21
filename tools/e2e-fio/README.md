# Mayastor E2E fio test pod
## Introduction
Derived from `dmonakhov/alpine-fio`


### Arguments
 * `sleep <sleep seconds>`
 * `segfault-after <delay seconds>`
 * `exitv <exit value>`
 * `-- <fio args> &` ( Note deprecated, use `--- fio ....` instead)
   * delimited by `--` and `&`
   * fork and run fio, fio arguments are delimited by `--` and `&`
   * multiple occurrences of this sequence are supported, a new separate process is created for each occurrence 
 * `--- <executable> <args> &`
   * delimited by `---` and `&`
   * each executable is run as a forked process
   * current list of executables
     * fio
     * disktest
   * multiple occurrences of this sequence are supported, a new separate process is created for each occurrence 
 * `--- <executable> <args> $` 
   * delimited by `---` and `&`
   * `executable` is launched and on completion parsing resumes.
 * `command <args> +`
   * delimited by `+`
   * `command` is launched as a system call it will run in asynchronous
 1. `fio` is only run if `fio` arguments are specified.
 2. all `fio` instances are run as a forked processes.
 3. all options can be specified multiple time for example
  *  `sleep 10  -- <fio args> & sleep 30` will sleep 10 seconds launch fio then sleep 30 seconds
 4. `exitv <v>` override exit value - this is to simulate failure.
 5. argument parsing is simple, invalid specifications are skipped over for example `"sleep --"` => `sleep` is skipped over, parsing resumes from `--`, execution does not fail.
 6. the pod will only complete after all fio processes (if any) have completed as well as sleep and segfault commands.


### Exit value
* If `exitv` is specified that is *always* returned.
* If all instances of `fio` ran successfully exit value is 0
* If a single instance of `fio` fails the exit value is the exit value of the failing instance of `fio`
* If multiple instances of `fio` fail the exit value is the exit value of a failing instance of `fio`

## building
Run `./build.sh`

This builds the image `mayadata/e2e-fio`

