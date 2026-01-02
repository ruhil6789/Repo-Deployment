package docker

// Docker client wrapper will be implemented here
// This will handle Docker API operations for building images

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type Client struct {
	cli *client.Client
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &Client{cli: cli}, nil
}

func (c *Client) BuildImage(ctx context.Context, buildContext io.Reader, imageTag string, dockerfile string) error {
	buildOptions := types.ImageBuildOptions{
		Tags:       []string{imageTag},
		Dockerfile: dockerfile,
		Remove:     true,
	}

	response, err := c.cli.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Read build output (logs)
	_, err = io.Copy(io.Discard, response.Body)
	return err
}

func (c *Client) PushImage(ctx context.Context, imageTag string) error {
	// TODO: Implement image push to registry
	return nil
}
