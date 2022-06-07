package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	"google.golang.org/api/iterator"
	instancepb "google.golang.org/genproto/googleapis/spanner/admin/instance/v1"
	"google.golang.org/genproto/protobuf/field_mask"
)

var (
	projectPath = "projects/my-project"
	// baseConfigName = "nam7" OR baseConfigName = "eur6"
	baseConfigName = "base-config-with-optional-replicas"
	// Custom config names must start with the prefix “custom-”.
	withName        = "custom-quickstart-go"
	withDisplayName = "Custom quickstart go"
	withLabels      = map[string]string{
		"cmmr_phase2_quickstart_go": "true",
	}
)

type command func(ctx context.Context, w io.Writer, adminClient *instance.InstanceAdminClient) error

var (
	commands = map[string]command{
		"create_instance_config":          createInstanceConfig,
		"update_instance_config":          updateInstanceConfig,
		"delete_instance_config":          deleteInstanceConfig,
		"list_instance_config_operations": listInstanceConfigOperations,
	}
)

// getInstanceConfig fetches the details of an instance config
func getInstanceConfig(ctx context.Context, w io.Writer, adminClient *instance.InstanceAdminClient, name string) (*instancepb.InstanceConfig, error) {
	config, err := adminClient.GetInstanceConfig(ctx, &instancepb.GetInstanceConfigRequest{
		Name: projectPath + "/instanceConfigs/" + name,
	})
	if err != nil {
		return nil, err
	}
	return config, nil
}

// createInstanceConfig used to create a custom instance config using a base config.
func createInstanceConfig(ctx context.Context, w io.Writer, adminClient *instance.InstanceAdminClient) error {
	baseConfig, err := getInstanceConfig(ctx, w, adminClient, baseConfigName)
	if err != nil {
		return err
	}
	op, err := adminClient.CreateInstanceConfig(ctx, &instancepb.CreateInstanceConfigRequest{
		Parent: projectPath,
		// Custom config names must start with the prefix “custom-”.
		InstanceConfigId: withName,
		InstanceConfig: &instancepb.InstanceConfig{
			Name:        projectPath + "/instanceConfigs/" + withName,
			DisplayName: withDisplayName,
			ConfigType:  instancepb.InstanceConfig_USER_MANAGED,
			// All replicas need to be specified for the replicas argument, including the ones in the base configuration.
			// base config replicas with optional Read-only replicas.
			Replicas:   append(baseConfig.Replicas, baseConfig.OptionalReplicas...),
			BaseConfig: projectPath + "/instanceConfigs/" + baseConfigName,
			Labels:     withLabels},
	})
	if err != nil {
		return err
	}
	if _, err := op.Wait(ctx); err != nil {
		return err
	}

	config, err := getInstanceConfig(ctx, w, adminClient, withName)
	if err != nil {
		return err
	}
	configStr, _ := json.MarshalIndent(config, "", "\t")
	fmt.Fprintf(w, "Created instance config [%s]\n", configStr)
	return nil
}

// updateInstanceConfig is used to change the display name or labels of a custom configuration. Note that the replicas of the configuration are immutable.
func updateInstanceConfig(ctx context.Context, w io.Writer, adminClient *instance.InstanceAdminClient) error {
	config, err := getInstanceConfig(ctx, w, adminClient, withName)
	if err != nil {
		return err
	}
	config.DisplayName = "Updated custom quickstart go"
	config.Labels["updated"] = "true"
	adminClient.UpdateInstanceConfig(ctx, &instancepb.UpdateInstanceConfigRequest{
		InstanceConfig: config,
		UpdateMask:     &field_mask.FieldMask{Paths: []string{"display_name", "labels"}},
	})

	conf, err := getInstanceConfig(ctx, w, adminClient, withName)
	if err != nil {
		return err
	}
	configStr, _ := json.MarshalIndent(conf, "", "\t")
	fmt.Fprintf(w, "Updated instance config [%s]\n", configStr)
	return nil
}

// deleteInstanceConfig deletes the custom instance config unless the config is in use by any running Spanner instance.
func deleteInstanceConfig(ctx context.Context, w io.Writer, adminClient *instance.InstanceAdminClient) error {
	err := adminClient.DeleteInstanceConfig(ctx, &instancepb.DeleteInstanceConfigRequest{
		Name: projectPath + "/instanceConfigs/" + withName,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "Deleted instance config [%s]\n", withName)
	return nil
}

// listInstanceConfigOperations fetched all the custom instance config operations
func listInstanceConfigOperations(ctx context.Context, w io.Writer, adminClient *instance.InstanceAdminClient) error {
	iter := adminClient.ListInstanceConfigOperations(ctx, &instancepb.ListInstanceConfigOperationsRequest{
		Parent: projectPath,
		Filter: `(metadata.@type=type.googleapis.com/google.spanner.admin.instance.v1.CreateInstanceConfigMetadata) AND
		    (metadata.instance_config.name:custom-quickstart-go)`,
	})
	for {
		op, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(w, "Instance config operation for [%v] has status [%v]\n", op.Name, op.Done)

	}
	return nil
}

func createClients(ctx context.Context) *instance.InstanceAdminClient {
	adminClient, err := instance.NewInstanceAdminClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	return adminClient
}

func run(ctx context.Context, adminClient *instance.InstanceAdminClient, w io.Writer, cmd string) error {
	cmdFn := commands[cmd]
	if cmdFn == nil {
		flag.Usage()
		os.Exit(2)
	}
	err := cmdFn(ctx, w, adminClient)
	if err != nil {
		fmt.Fprintf(w, "%s failed with %v", cmd, err)
	}
	return err
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: cmmr-phase2-quickstart <command>
	Command can be one of: create_instance_config, update_instance_config, delete_instance_config, list_instance_config_operations

Examples:
	cmmr-phase2-quickstart create_instance_config
	cmmr-phase2-quickstart update_instance_config
	cmmr-phase2-quickstart delete_instance_config
	cmmr-phase2-quickstart list_instance_config_operations

`)
	}

	flag.Parse()
	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(2)
	}

	cmd := flag.Arg(0)
	// Add timeout to context.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	adminClient := createClients(ctx)
	defer adminClient.Close()
	if err := run(ctx, adminClient, os.Stdout, cmd); err != nil {
		os.Exit(1)
	}
}
