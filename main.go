package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cilium/ebpf/rlimit"
	containercollection "github.com/inspektor-gadget/inspektor-gadget/pkg/container-collection"
	containerutils "github.com/inspektor-gadget/inspektor-gadget/pkg/container-utils/types"
	"github.com/inspektor-gadget/inspektor-gadget/pkg/types"
	"github.com/inspektor-gadget/inspektor-gadget/pkg/utils/host"
	"k8s.io/client-go/rest"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err.Error())
		os.Exit(1)
	}
}

func run() error {
	if err := rlimit.RemoveMemlock(); err != nil {
		return fmt.Errorf("removing memory limit: %v", err)
	}

	err := host.Init(host.Config{
		AutoMountFilesystems: false,
	})
	if err != nil {
		return fmt.Errorf("initializing host filesystem: %w", err)
	}

	containerEventFuncs := []containercollection.FuncNotify{
		containerCallback,
	}

	nodeName := os.Getenv("NODE_NAME")
	kubeconfig, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("getting in-cluster config: %w", err)
	}

	opts := []containercollection.ContainerCollectionOption{
		containercollection.WithPubSub(containerEventFuncs...),
		containercollection.WithOCIConfigEnrichment(),
		containercollection.WithCgroupEnrichment(),
		containercollection.WithLinuxNamespaceEnrichment(),
		containercollection.WithMultipleContainerRuntimesEnrichment(
			[]*containerutils.RuntimeConfig{
				{Name: types.RuntimeNameDocker},
				{Name: types.RuntimeNameContainerd},
			}),
		//containercollection.WithContainerRuntimeEnrichment(ch.runtime),
		containercollection.WithContainerFanotifyEbpf(),
		containercollection.WithKubernetesEnrichment(nodeName, kubeconfig),
	}

	containerCollection := &containercollection.ContainerCollection{}

	err = containerCollection.Initialize(opts...)
	if err != nil {
		return fmt.Errorf("initializing container collection: %w", err)
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
	<-exit

	return nil
}

func containerCallback(notif containercollection.PubSubEvent) {
	fmt.Printf("Container event: name %s, image name %s, image digest %s\n", notif.Container.Runtime.ContainerName, notif.Container.Runtime.ContainerImageName, notif.Container.Runtime.ContainerImageDigest)
}
