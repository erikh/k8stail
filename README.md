# k8stail is for tailing logs

This is a tiny program you can use to spawn lots of simultaneous log readers into kubernetes to read logs. It is a little resource intensive if you point it at at the full cluster. It has a few options:

```
GLOBAL OPTIONS:
   --config value, -c value     kubectl config file [$KUBECONFIG]
   --namespace value, -n value  Namespaces you care about; may provide multiple
   --since value, -s value      Will only show logs after this time in seconds (default: 0)
   --after value, -a value      Will only show logs that appear after this time (default: (*time.Time)(nil))
   --timestamp, -t              Timestamp log messages (default: true)
   --help, -h                   show help (default: false)
   --version, -v                print the version (default: false)
```

This is also available via `--help`.

The default is to tail the entire cluster from the point in time the application was launched.

## BUGS

Marty, where we're going, we don't need bugs.

## Author

Erik Hollensbe <erik+github@hollensbe.org>
