package coverage

import (
	"context"
	"os"
	"time"

	"dagger.io/dagger"
)

func NewVolume(ctx context.Context, client *dagger.Client, c *dagger.Container, dir string) (*dagger.CacheVolume, func() error, error) {
	cover := client.CacheVolume("cover")

	owner, err := c.User(ctx)
	if err != nil {
		return nil, nil, err
	}

	_, err = c.WithUser("root").
		WithMountedCache("/cover", cover).
		WithEnvVariable("BUST", time.Now().String()).
		WithExec([]string{"sh", "-c", "rm -r /cover/* || echo empty"}).
		WithExec([]string{"chown", owner + ":" + owner, "/cover"}).
		ExitCode(ctx)
	if err != nil {
		return nil, nil, err
	}

	return cover, func() error {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}

		_, err := client.Container().From("golang:1.20-alpine3.16").
			WithEnvVariable("BUST", time.Now().String()).
			WithMountedCache("/cover", cover).
			WithExec([]string{"mkdir", "/cover/total"}).
			WithExec([]string{"sh", "-c", "go tool covdata merge -i $(find /cover/* -type d | paste -sd \",\" -) -o /cover/total"}).
			WithExec([]string{"sh", "-c", "go tool covdata textfmt -i /cover/total -o /cover/total.txt"}).
			ExitCode(ctx)
		if err != nil {
			return err
		}

		_, err = client.Container().From("alpine:3.16").
			WithEnvVariable("BUST", time.Now().String()).
			WithMountedCache("/cover", cover).
			WithExec([]string{"cp", "-r", "/cover", "/cover-out"}).
			Directory("/cover-out").Export(ctx, dir)
		return err
	}, nil
}
