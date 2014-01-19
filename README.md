# monitoring

This is a collection of monitoring checks, tools, ... which I had to write because I did not find any suitable solutions. Besides this repository I wrote also the following monitoring solutions:

* [Tirion](https://github.com/zimmski/tirion) a complete infrastructure for monitoring applications during benchmarks
* [Glamor](https://github.com/zimmski/glamor) a daemon for monitoring hosts via ICMP echo request (ping)

The repository is structured in the following way:

* **/check** every file holds one specific check and can be used on its own.
* **/check_mk** every folder holds one check_mk check which consist of one agent and one server script.
