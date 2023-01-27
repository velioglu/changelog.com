package kosli

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
	"github.com/magefile/mage/mg"
	"github.com/thechangelog/changelog.com/magefiles/env"
)

const (
	// https://github.com/kosli-dev/cli/pkgs/container/cli
	_Image = "ghcr.io/kosli-dev/cli:v0.1.35"

	// https://app.kosli.com/changelog/environments/fly-changelog-2022-03-13/events/
	EnvironmentName = "fly-changelog-2022-03-13"
	EnvironmentKind = "server"
	PipelineName    = "changelog-2023-02-04"
)

type Kosli mg.Namespace

// Declare an environment
func (t Kosli) Environment(ctx context.Context) error {
	dag, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer dag.Close()

	return Environment(ctx, dag)
}

// Declare a pipeline
func (t Kosli) Pipeline(ctx context.Context) error {
	dag, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer dag.Close()

	return Pipeline(ctx, dag)
}

// Report an image artifact
func (t Kosli) Image(ctx context.Context, ref string, SHA256 string) error {
	dag, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer dag.Close()

	return Image(ctx, dag, ref, SHA256)
}

// Report a deployment
func (t Kosli) Deployment(ctx context.Context, ref string, SHA256 string) error {
	dag, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer dag.Close()

	return Deployment(ctx, dag, ref, SHA256)
}

func Environment(ctx context.Context, dag *dagger.Client) error {
	output, err := kosli(ctx, dag).WithExec([]string{
		"kosli", "environment", "declare",
		"--name", EnvironmentName,
		"--environment-type", EnvironmentKind,
	}).
		Stdout(ctx)

	fmt.Println(output)

	return err
}

func Pipeline(ctx context.Context, dag *dagger.Client) error {
	output, err := kosli(ctx, dag).WithExec([]string{
		"kosli", "pipeline", "declare",
		"--template", "artifact",
	}).
		Stdout(ctx)

	fmt.Println(output)

	return err
}

func Image(ctx context.Context, dag *dagger.Client, ref string, SHA256 string) error {
	githubServerURL := env.Val(ctx, env.HostEnv(ctx, dag.Host(), "GITHUB_SERVER_URL"))
	githubRepository := env.Val(ctx, env.HostEnv(ctx, dag.Host(), "GITHUB_REPOSITORY"))
	githubSHA := env.Val(ctx, env.HostEnv(ctx, dag.Host(), "GITHUB_SHA"))
	commitURL := fmt.Sprintf("%s/%s/commit/%s", githubServerURL, githubRepository, githubSHA)

	output, err := kosliGitHub(ctx, dag).
		WithExec([]string{
			"kosli", "pipeline", "artifact", "report", "creation", ref,
			"--commit-url", commitURL,
			"--git-commit", githubSHA,
			"--sha256", SHA256,
		}).
		Stdout(ctx)

	fmt.Println(output)

	return err
}

func Deployment(ctx context.Context, dag *dagger.Client, ref string, SHA256 string) error {
	output, err := kosliGitHub(ctx, dag).
		WithExec([]string{
			"kosli", "expect", "deployment", ref,
			"--environment", EnvironmentName,
			"--sha256", SHA256,
		}).
		Stdout(ctx)

	fmt.Println(output)

	return err
}

func kosli(ctx context.Context, dag *dagger.Client) *dagger.Container {
	owner := env.Val(ctx, env.HostEnv(ctx, dag.Host(), "KOSLI_OWNER"))
	token := env.HostEnv(ctx, dag.Host(), "KOSLI_API_TOKEN").Secret()

	return dag.Container().
		From(_Image).
		WithEnvVariable("KOSLI_OWNER", owner).
		WithSecretVariable("KOSLI_API_TOKEN", token).
		WithEnvVariable("KOSLI_PIPELINE", PipelineName)
}

func kosliGitHub(ctx context.Context, dag *dagger.Client) *dagger.Container {
	githubServerURL := env.Val(ctx, env.HostEnv(ctx, dag.Host(), "GITHUB_SERVER_URL"))
	githubRepository := env.Val(ctx, env.HostEnv(ctx, dag.Host(), "GITHUB_REPOSITORY"))
	githubRunID := env.Val(ctx, env.HostEnv(ctx, dag.Host(), "GITHUB_RUN_ID"))
	buildURL := fmt.Sprintf("%s/%s/actions/runs/%s", githubServerURL, githubRepository, githubRunID)

	src := dag.Host().Directory(".", dagger.HostDirectoryOpts{
		Include: []string{
			".git",
		},
	})

	return kosli(ctx, dag).
		WithEnvVariable("KOSLI_BUILD_URL", buildURL).
		WithMountedDirectory("/workdir", src).
		WithWorkdir("/workdir")
}

// ⏳ waiting for Service Accounts to be enabled for  changelog.1password.com
//
// func op(ctx context.Context, dag *dagger.Client) {
// 	opAddress := envVal(ctx, hostEnv(ctx, dag.Host(), "OP_ADDRESS"))
// 	opEmail := envVal(ctx, hostEnv(ctx, dag.Host(), "OP_EMAIL"))
// 	opSecretKey := hostEnv(ctx, dag.Host(), "OP_SECRET_KEY").Secret()

// 	output, _ := dag.Container().
// 		From("1password/op:2").
// 		WithSecretVariable("OP_SECRET_KEY", opSecretKey).
// 		WithExec([]string{
// 			"echo",
// 			"2023-01-15.2",
// 		}).
// 		WithExec([]string{
// 			"env",
// 		}).
// 		WithExec([]string{
// 			"op", "account", "add",
// 			"--address", opAddress,
// 			"--email", opEmail,
// 			"--signin",
// 		}).
// 		Stdout(ctx)

// 	panic(output)
// }
