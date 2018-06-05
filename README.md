# gcp-disk-snapshotter

Service to periodically take snapshots of `gcp` disks, based on particular labels.
Snapshots (only the ones created by the service) are deleted when they become older than the specified retention time.

## Usage
```
Usage of /gcp-disk-snapshotter:
  -interval_secs int
        Interval between snapshots in seconds. Defaults to 43200s = 12h (default 43200)
  -labels string
        (Required) Comma separated list of disk labels in format <name>:<value>
  -log_level string
        Log Level, defaults to INFO (default "info")
  -project string
        (Required) GCP Project to use
  -retention_hours int
        Retention Duration in hours. Defaults to 720h = 1 month (default 720)
  -snap_prefix string
        Prefix for created snapshots
  -watch_interval int
        Interval between watch cycles in seconds. Defaults to 60s (default 60)
  -zones string
        (Required) Comma separated list of zones where projects disks may live
```
