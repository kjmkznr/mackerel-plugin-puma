mackerel-plugin-puma
====================

puma custom metrics plugin for mackerel.io agent.

[![Build Status](https://travis-ci.org/kjmkznr/mackerel-plugin-puma.svg?branch=master)](https://travis-ci.org/kjmkznr/mackerel-plugin-puma)

## Synopsis

```shell
mackerel-plugin-puma -state=<socket file> -token=<puma control token> [-metric-key-prefix=puma]
```

## Example of mackerel-agent.conf

```
[plugin.metrics.puma]
command = "/path/to/mackerel-plugin-puma -state /tmp/puma.sock -token foo"
```
