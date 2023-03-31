package testing

import (
	"context"
	"os/exec"
	"path"
	"strings"

	"dagger.io/dagger"
)

func Unit(ctx context.Context, client *dagger.Client, flipt *dagger.Container) error {
	// create Redis service container
	redisSrv := client.Container().
		From("redis").
		WithExposedPort(6379).
		WithExec(nil)

	testCmd := []string{"go", "test", "-race", "-tags", "assets,netgo", "-p", "1",
		"-cover", "-covermode", "atomic"}
	suffix := []string{"./..."}

	if dir, err := flipt.EnvVariable(ctx, "GOCOVERDIR"); err == nil && dir != "" {
		out, err := exec.Command("go", "list", "go.flipt.io/flipt/...").CombinedOutput()
		if err != nil {
			return err
		}

		pkgs := strings.Split(string(out), "\n")

		// filter out empty entries (trailing newline)
		for i := 0; (i < len(pkgs)) && (len(pkgs) > 0); {
			if strings.TrimSpace(pkgs[i]) == "" {
				// remove empty item
				pkgs = append(pkgs[:i], pkgs[i+1:]...)
				continue
			}
			i++
		}

		testCmd = append(testCmd, "-coverpkg", strings.Join(pkgs, ","))
		suffix = append(suffix, "-args", "-test.gocoverdir", dir)
	}

	_, err := flipt.
		WithServiceBinding("redis", redisSrv).
		WithEnvVariable("REDIS_HOST", "redis:6379").
		WithExec([]string{"sh", "-xc", strings.Join(append(testCmd, suffix...), " ")}).
		ExitCode(ctx)
	return err
}

var All = map[string]Wrapper{
	"sqlite":    WithSQLite,
	"postgres":  WithPostgres,
	"mysql":     WithMySQL,
	"cockroach": WithCockroach,
}

type Wrapper func(context.Context, *dagger.Client, *dagger.Container) (context.Context, *dagger.Client, *dagger.Container)

func WithCoverage(mount string, cover *dagger.CacheVolume) Wrapper {
	return func(ctx context.Context, client *dagger.Client, container *dagger.Container) (context.Context, *dagger.Client, *dagger.Container) {
		mount := path.Join("/cover", mount)
		container = container.
			WithMountedCache("/cover", cover).
			WithEnvVariable("GOCOVERDIR", mount).
			WithExec([]string{"mkdir", mount})

		if _, err := container.ExitCode(ctx); err != nil {
			panic(err)
		}

		return ctx, client, container
	}
}

func WithSQLite(ctx context.Context, client *dagger.Client, container *dagger.Container) (context.Context, *dagger.Client, *dagger.Container) {
	return ctx, client, container
}

func WithPostgres(ctx context.Context, client *dagger.Client, flipt *dagger.Container) (context.Context, *dagger.Client, *dagger.Container) {
	return ctx, client, flipt.
		WithEnvVariable("FLIPT_TEST_DATABASE_PROTOCOL", "postgres").
		WithEnvVariable("FLIPT_TEST_DB_URL", "postgres://postgres:password@postgres:5432").
		WithServiceBinding("postgres", client.Container().
			From("postgres").
			WithEnvVariable("POSTGRES_PASSWORD", "password").
			WithExposedPort(5432).
			WithExec(nil))
}

func WithMySQL(ctx context.Context, client *dagger.Client, flipt *dagger.Container) (context.Context, *dagger.Client, *dagger.Container) {
	return ctx, client, flipt.
		WithEnvVariable("FLIPT_TEST_DATABASE_PROTOCOL", "mysql").
		WithEnvVariable(
			"FLIPT_TEST_DB_URL",
			"mysql://flipt:password@mysql:3306/flipt_test?multiStatements=true",
		).
		WithServiceBinding("mysql", client.Container().
			From("mysql:8").
			WithEnvVariable("MYSQL_USER", "flipt").
			WithEnvVariable("MYSQL_PASSWORD", "password").
			WithEnvVariable("MYSQL_DATABASE", "flipt_test").
			WithEnvVariable("MYSQL_ALLOW_EMPTY_PASSWORD", "true").
			WithExposedPort(3306).
			WithExec(nil))
}

func WithCockroach(ctx context.Context, client *dagger.Client, flipt *dagger.Container) (context.Context, *dagger.Client, *dagger.Container) {
	return ctx, client, flipt.
		WithEnvVariable("FLIPT_TEST_DATABASE_PROTOCOL", "cockroachdb").
		WithEnvVariable("FLIPT_TEST_DB_URL", "cockroachdb://root@cockroach:26257/defaultdb?sslmode=disable").
		WithServiceBinding("cockroach", client.Container().
			From("cockroachdb/cockroach:latest-v21.2").
			WithEnvVariable("COCKROACH_USER", "root").
			WithEnvVariable("COCKROACH_DATABASE", "defaultdb").
			WithExposedPort(26257).
			WithExec([]string{"start-single-node", "--insecure"}))
}
