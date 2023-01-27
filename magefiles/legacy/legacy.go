package legacy

import (
	"context"
	"os"

	"dagger.io/dagger"
	"github.com/magefile/mage/mg"
	"github.com/thechangelog/changelog.com/magefiles/docker"
	"github.com/thechangelog/changelog.com/magefiles/env"
)

type Legacy mg.Namespace

// Build, test & publish using dagger v0.1.0 pipeline
func (t Legacy) Shipit(ctx context.Context) error {
	dag, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer dag.Close()

	return ShipIt(ctx, dag)
}

func ShipIt(ctx context.Context, dag *dagger.Client) error {
	user := env.Val(ctx, env.HostEnv(ctx, dag.Host(), "USER"))

	daggerLogLevel := env.Val(ctx, env.HostEnv(ctx, dag.Host(), "DAGGER_LOG_LEVEL"))
	daggerLogFormat := env.Val(ctx, env.HostEnv(ctx, dag.Host(), "DAGGER_LOG_FORMAT"))

	dockerhubUsername := env.Val(ctx, env.HostEnv(ctx, dag.Host(), "DOCKERHUB_USERNAME"))
	dockerhubPassword := env.HostEnv(ctx, dag.Host(), "DOCKERHUB_PASSWORD").Secret()

	app := dag.Host().Directory(".", dagger.HostDirectoryOpts{
		Include: []string{
			".git",
			"2021.dagger",
			"assets",
			"config",
			"docker",
			"lib",
			"priv",
			"test",
			".dockerignore",
			"Makefile",
			"mix.exs",
			"mix.lock",
		},
		Exclude: []string{
			"**/node_modules",
		},
	})

	_, err := docker.Container(ctx, dag).
		WithEnvVariable("USER", user).
		WithEnvVariable("DAGGER_LOG_LEVEL", daggerLogLevel).
		WithEnvVariable("DAGGER_LOG_FORMAT", daggerLogFormat).
		WithEnvVariable("DOCKERHUB_USERNAME", dockerhubUsername).
		WithSecretVariable("DOCKERHUB_PASSWORD", dockerhubPassword).
		WithDirectory("/app", app).
		WithWorkdir("/app").
		WithExec([]string{
			"make", "--directory=2021.dagger", "ship-it",
		}).
		ExitCode(ctx)

	return err
}
