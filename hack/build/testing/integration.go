package testing

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"path"
	"strings"
	"time"

	"dagger.io/dagger"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"
)

var (
	protocolPorts = map[string]string{"http": "8080", "grpc": "9000"}
	replacer      = strings.NewReplacer(" ", "-", "/", "-")
)

type testConfig struct {
	name      string
	namespace string
	address   string
	token     string
}

type IntegrationRequest struct {
	Base  *dagger.Container
	Flipt *dagger.Container
	Cover *dagger.CacheVolume
}

func Integration(ctx context.Context, client *dagger.Client, req IntegrationRequest) error {
	var (
		base  = req.Base
		flipt = req.Flipt
		cover = req.Cover

		// unique volumes per invocation
		logs = client.CacheVolume(fmt.Sprintf("logs-%s", uuid.New()))
	)

	_, err := flipt.WithUser("root").
		WithMountedCache("/logs", logs).
		WithExec([]string{"chown", "flipt:flipt", "/logs"}).
		ExitCode(ctx)
	if err != nil {
		return err
	}

	var cases []testConfig

	for _, namespace := range []string{
		"",
		fmt.Sprintf("%x", rand.Int()),
	} {
		for protocol, port := range protocolPorts {
			address := fmt.Sprintf("%s://flipt:%s", protocol, port)
			cases = append(cases,
				testConfig{
					name:      fmt.Sprintf("%s namespace %v no authentication", strings.ToUpper(protocol), namespace != ""),
					namespace: namespace,
					address:   address,
				},
				testConfig{
					name:      fmt.Sprintf("%s namespace %v with authentication", strings.ToUpper(protocol), namespace != ""),
					namespace: namespace,
					address:   address,
					token:     "some-token",
				},
			)
		}
	}

	rand.Seed(time.Now().Unix())

	var g errgroup.Group

	for _, test := range []struct {
		name string
		fn   func(_ context.Context, base, flipt *dagger.Container, cover *dagger.CacheVolume, conf testConfig) func() error
	}{
		{
			name: "api",
			fn:   api,
		},
		{
			name: "import/export",
			fn:   importExport,
		},
	} {
		for _, config := range cases {
			config := config

			flipt := flipt
			if config.token != "" {
				flipt = flipt.
					WithEnvVariable("FLIPT_AUTHENTICATION_REQUIRED", "true").
					WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_ENABLED", "true").
					WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_BOOTSTRAP_TOKEN", config.token)
			}

			var (
				name      = strings.ToLower(replacer.Replace(fmt.Sprintf("flipt-test-%s-config-%s", test.name, config.name)))
				logPath   = fmt.Sprintf("/var/opt/flipt/logs/%s.log", name)
				coverPath = path.Join("/var/opt/flipt/cover", name)
			)

			flipt = flipt.
				WithEnvVariable("FLIPT_LOG_LEVEL", "debug").
				WithEnvVariable("FLIPT_LOG_FILE", logPath).
				WithEnvVariable("GOCOVERDIR", coverPath).
				WithMountedCache("/var/opt/flipt/logs", logs).
				WithMountedCache("/var/opt/flipt/cover", cover).
				WithExec([]string{"mkdir", coverPath})

			_, err := flipt.ExitCode(ctx)
			if err != nil {
				return err
			}

			g.Go(test.fn(ctx, base, flipt, cover, config))
		}
	}

	err = g.Wait()

	logsCopy := client.Container().From("alpine:3.16").
		WithMountedCache("/logs", logs).
		WithExec([]string{"cp", "-r", "/logs", "/logs-out"})

	if _, lerr := logsCopy.Directory("/logs-out").Export(ctx, "hack/build/logs"); lerr != nil {
		log.Println("error copying logs", lerr)
	}

	return err
}

func api(ctx context.Context, base, flipt *dagger.Container, cover *dagger.CacheVolume, conf testConfig) func() error {
	return suite(ctx, "api", base,
		// create unique instance for test case
		flipt.
			WithEnvVariable("UNIQUE", uuid.New().String()).
			WithExec(nil), cover, conf)
}

func importExport(ctx context.Context, base, flipt *dagger.Container, cover *dagger.CacheVolume, conf testConfig) func() error {
	return func() error {
		// import testdata before running readonly suite
		flags := []string{"--address", conf.address}
		if conf.token != "" {
			flags = append(flags, "--token", conf.token)
		}

		if conf.namespace != "" {
			flags = append(flags, "--namespace", conf.namespace)
		}

		var (
			// create unique instance for test case
			fliptToTest = flipt.
					WithEnvVariable("UNIQUE", uuid.New().String()).
					WithExec(nil)

			importCmd = append([]string{"/bin/flipt", "import"}, append(flags, "--create-namespace", "import.yaml")...)
			seed      = base.File("hack/build/testing/integration/readonly/testdata/seed.yaml")
		)
		// use target flipt binary to invoke import
		_, err := flipt.
			// copy testdata import yaml from base
			WithFile("import.yaml", seed).
			WithServiceBinding("flipt", fliptToTest).
			WithEnvVariable("UNIQUE", uuid.New().String()).
			// it appears it takes a little while for Flipt to come online
			// For the go tests they have to compile and that seems to be enough
			// time for the target Flipt to come up.
			// However, in this case the flipt binary is prebuilt and needs a little sleep.
			WithExec([]string{"sh", "-c", fmt.Sprintf("sleep 4 && %s", strings.Join(importCmd, " "))}).
			ExitCode(ctx)
		if err != nil {
			return err
		}

		// run readonly suite against imported Flipt instance
		if err := suite(ctx, "readonly", base, fliptToTest, cover, conf)(); err != nil {
			return err
		}

		expectedYAML, err := seed.Contents(ctx)
		if err != nil {
			return err
		}

		exportPath := "/var/opt/flipt/export.yaml"
		// use target flipt binary to invoke import
		exportedYAML, err := flipt.
			WithServiceBinding("flipt", fliptToTest).
			WithEnvVariable("UNIQUE", uuid.New().String()).
			WithExec(append([]string{"/bin/flipt", "export", "-o", exportPath}, flags...)).
			File(exportPath).
			Contents(ctx)
		if err != nil {
			return err
		}

		var expected, exported map[string]interface{}
		if err := yaml.Unmarshal([]byte(expectedYAML), &expected); err != nil {
			return err
		}

		if err := yaml.Unmarshal([]byte(exportedYAML), &exported); err != nil {
			return err
		}

		if diff := cmp.Diff(expected, exported); diff != "" {
			fmt.Println("(-expected/+found):\n", diff)

			return errors.New("Exported yaml did not match.")
		}

		return nil
	}
}

func suite(ctx context.Context, dir string, base, flipt *dagger.Container, cover *dagger.CacheVolume, conf testConfig) func() error {
	return func() error {
		flags := []string{"--flipt-addr", conf.address}
		if conf.namespace != "" {
			flags = append(flags, "--flipt-namespace", conf.namespace)
		}

		if conf.token != "" {
			flags = append(flags, "--flipt-token", conf.token)
		}

		testCoverName := fmt.Sprintf("go-test-%s-config-%s", dir, conf.name)
		testCoverDir := path.Join("/cover", strings.ToLower(replacer.Replace(testCoverName)))

		testCmd := strings.Join(append([]string{
			"go", "test",
			"-cover", "-covermode", "atomic",
			"-coverpkg", "go.flipt.io/flipt/sdk/go,go.flipt.io/flipt/sdk/go/http,go.flipt.io/flipt/sdk/go/grpc",
			"-v", "-race", "."},
			append(flags, "-test.gocoverdir", testCoverDir)...), " ")

		_, err := base.
			WithMountedCache("/cover", cover).
			WithExec([]string{"mkdir", testCoverDir}).
			WithWorkdir(path.Join("hack/build/testing/integration", dir)).
			WithServiceBinding("flipt", flipt).
			WithExec([]string{"sh", "-xc", testCmd}).
			// Dagger actually terminates service bound containers using SIGKILL which
			// leads to Go not automatically writing out coverage data on exit.
			// Instead we mount an endpoint on Flipt when coverage is enabled which
			// we trigger before exiting the foreground container.
			WithExec([]string{"sh", "-c", "wget -O - http://flipt:8080/cover || echo failed to cover"}).
			ExitCode(ctx)

		return err
	}
}
