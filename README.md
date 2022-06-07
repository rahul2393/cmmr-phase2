# Cloud Spanner Client Library for Golang

This version includes support for the Configurable read-only replicas functionality.

The subdirectories have the following contents:

* `google-cloud-go` -- Contains the updated code to support the Configurable read-only replicas functionality.
* `go-genproto` -- Contains the updated service protos to support the Configurable
  read-only replicas functionality.

### Client Library

To use this version of the client library, please modify your `go.mod` file
to replace the `cloud.google.com/go/spanner` and `google.golang.org/genproto`
dependencies with the versions in the sub-directory by adding the following
lines:

```
replace cloud.google.com/go/spanner => <path>/google-cloud-go/spanner
replace google.golang.org/genproto => <path>/go-genproto
```

### Run the Quickstart samples
1. Edit `cmmr-phase2-quickstart.go` with your projectPath, baseConfigName, withName, withDisplayName and withLabels.
2. Run the sample using the following commands.
```
go run cmmr-phase2-quickstart.go create_instance_config
go run cmmr-phase2-quickstart.go update_instance_config
go run cmmr-phase2-quickstart.go delete_instance_config
go run cmmr-phase2-quickstart.go list_instance_config_operations
```
