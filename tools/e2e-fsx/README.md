# Mayastor E2E fsx-linux test pod
## Introduction
`FSX (File System Exerciser)` tool is used to test the file systems that are part of `Mayastor`

Derived from [linux-test-project](https://github.com/linux-test-project/ltp/tree/master/testcases/kernel/fs/fsx-linux)

For more details check [MQ-1784](https://mayadata.atlassian.net/browse/MQ-1784)

### Arguments
 * `first argument will be scratch device to use for testing`
 * `second argument will be file system type(optional)`
 * `third argument will be number of operations to perform`

### Note
 1. `fsx` is only run if all mandatory arguments are specified.
 2. all arguments should be specified in same manner as specified above.
 2. the pod will only complete after `fsx` process have completed as well as `fsck` or `e2fsck` command.


### Exit value
* If instance of `fsx` and `fsck` or `e2fsck`ran successfully then exit value is 0

## Building
Run `./build.sh`

This builds the image `mayadata/e2e-fsx`

## fsx-linux usage
```
usage: fsx-linux [-dnqLOW] [-b opnum] [-c Prob] [-l flen] [-m start:end] [-o oplen] [-p progressinterval] [-r readbdy] [-s style] [-t truncbdy] [-w writebdy] [-D startingop] [-N numops] [-P dirpath] [-S seed] [ -I random|rotate ] fname [additional paths to fname..]
	-b opnum: beginning operation number (default 1)
	-c P: 1 in P chance of file close+open at each op (default infinity)
	-d: debug output for all operations [-d -d = more debugging]
	-l flen: the upper bound on file size (default 262144)
	-m start:end: monitor (print debug) specified byte range (default 0:infinity)
	-n: no verifications of file size
	-o oplen: the upper bound on operation size (default 65536)
	-p progressinterval: debug output at specified operation interval
	-q: quieter operation
	-r readbdy: 4096 would make reads page aligned (default 1)
	-s style: 1 gives smaller truncates (default 0)
	-t truncbdy: 4096 would make truncates page aligned (default 1)
	-w writebdy: 4096 would make writes page aligned (default 1)
	-D startingop: debug output starting at specified operation
	-L: fsxLite - no file creations & no file size changes
	-N numops: total # operations to do (default infinity)
	-O: use oplen (see -o flag) for every op (default random)
	-P: save .fsxlog and .fsxgood files in dirpath (default ./)
	-S seed: for random # generator (default 1) 0 gets timestamp
	-W: mapped write operations DISabled
	-R: read() system calls only (mapped reads disabled)
	-I: When multiple paths to the file are given each operation uses
	    a different path.  Iterate through them in order with 'rotate'
	    or chose then at 'random'.  (defaults to random)
	fname: this filename is REQUIRED (no default)
```

## reference
* https://github.com/linux-test-project/ltp
* https://web.archive.org/web/20190115144026/http://codemonkey.org.uk/projects/fsx/
* https://stackoverflow.com/questions/21565865/filesystem-test-suites/25940371