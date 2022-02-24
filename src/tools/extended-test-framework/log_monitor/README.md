## Log monitor

The `log-monitor` is a separate pod that is based on fluentd. The fluentd pod is run on each node and collects all the
logs from entire node into a file. When the log monitor is deployed, it searches all fluentd pods and tails the log
files as a separate goroutines. All lines that will come from the tail are searched using a regular expression -
`level.{0,4}(error|warn)` is a default. When `log-monitor` matches this expression, then we get a test run from the
`test-director` pod that is in the `EXECUTION` state. `log-monitor` sends an event to the `test-director` pod
depending on the severity of the line. These log events are stored within the current test run, but do not affect
the state of the test run, even if the error severity is indicated in the log line. When the test run is over (FAIL|SUCCESS)
all messages belongs to test run are posted to xray - test execution as a comment.

![log_monitor architecture](./log_monitor.png?raw=true)