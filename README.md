mackerel-plugin-puma
====================

puma custom metrics plugin for mackerel.io agent.

## Synopsis

```shell
mackerel-plugin-puma -state=<socket file> -token=<puma control token> [-metric-key-prefix=gostats]
```

## Example of mackerel-agent.conf

```
[plugin.metrics.puma]
command = "/path/to/mackerel-plugin-puma -state /tmp/puma.sock -token foo"
```
