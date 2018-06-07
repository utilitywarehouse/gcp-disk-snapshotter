# gcp-disk-snapshotter

Service to periodically take snapshots of `gcp` disks, based on particular labels.
Snapshots (only the ones created by the service) are deleted when they become older than the specified retention time.

## Usage
```
Usage of /gcp-disk-snapshotter:
 -conf_file string
        (Required) Path of the configuration file tha contains the targets based on label or description
  -log_level string
        Log Level, defaults to INFO (default "info")
  -project string
        (Required) GCP Project to use
  -snap_prefix string
        Prefix for created snapshots
  -watch_interval int
        Interval between watch cycles in seconds. Defaults to 60s (default 60)
  -zones string
        (Required) Comma separated list of zones where projects disks may live
```

## Configuration File

Example Configuration File:

```
{
  "Descriptions": [
    {
      "retentionPeriodHours" : 2,
      "intervalSeconds" : 100,
      "description": {
        "key": "kubernetes.io/created-for/pvc/name",
        "value": "some-app-pd-pvc"
      }
    }
  ],
  "Labels": [
    {
      "retentionPeriodHours" : 2,
      "intervalSeconds" : 100,
      "label": {
        "key": "name",
        "value": "some-app-name"
      }
    }
  ]
}
```
