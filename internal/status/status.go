package status

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func GetStatus() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "CONTAINER ID\tIMAGE\tCREATED\tSTATUS\tPORTS\tNAMES")

	for _, container := range containers {
		ports := ""

		for _, port := range container.Ports {
			ports += fmt.Sprintf("%s:%d->%d/%s ", port.IP, port.PublicPort, port.PrivatePort, port.Type)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			container.ID[:10],
			container.Image,
			time.Since(time.Unix(container.Created, 0)),
			container.State,
			ports,
			container.Names)
	}

	w.Flush()
}
