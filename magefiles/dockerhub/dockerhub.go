package dockerhub

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
	"github.com/containers/image/v5/docker/reference"
	"github.com/magefile/mage/mg"
	"github.com/thechangelog/changelog.com/magefiles/docker"
	"github.com/thechangelog/changelog.com/magefiles/env"
)

type Dockerhub mg.Namespace

const (
	RootRepo = "thechangelog/changelog.com"
)

// Push container image to https://hub.docker.com/r/thechangelog/changelog.com/tags
func (t Dockerhub) Publish(ctx context.Context) error {
	dag, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer dag.Close()

	err = Auth(ctx, dag)
	if err != nil {
		return err
	}

	dagWithAuth, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer dagWithAuth.Close()
	_, err = Publish(ctx, dagWithAuth)
	if err != nil {
		return err
	}

	return nil
}

// Authenticates host Docker
func (t Dockerhub) Auth(ctx context.Context) error {
	dag, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer dag.Close()

	return Auth(ctx, dag)
}

func Publish(ctx context.Context, dag *dagger.Client) (string, error) {
	githubRepo := env.Val(ctx, env.HostEnv(ctx, dag.Host(), "GITHUB_REPOSITORY"))
	githubBranch := env.Val(ctx, env.HostEnv(ctx, dag.Host(), "GITHUB_REF_NAME"))
	githubBranchSafe := strings.ReplaceAll(githubBranch, "/", "-")

	if githubRepo == RootRepo {
		image := fmt.Sprintf("%s:%s", RootRepo, githubBranchSafe)

		refWithSHA, err := dag.Container().From(image).Publish(ctx, image)
		if err != nil {
			return "", err
		}
		imageRefFlyValid, err := reference.ParseDockerRef(refWithSHA)
		if err != nil {
			return "", err
		}

		return imageRefFlyValid.String(), nil
	}

	return "", errors.New(fmt.Sprintf("\n📦 Publishing runs only in CI, %s repo", RootRepo))
}

func Auth(ctx context.Context, dag *dagger.Client) error {
	hostDockerConfigDir := os.Getenv("DOCKER_CONFIG")
	if hostDockerConfigDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		hostDockerConfigDir = filepath.Join(home, ".docker")
	}
	hostDockerClientConfig := filepath.Join(hostDockerConfigDir, "config.json")

	dockerhubUsername := env.Val(ctx, env.HostEnv(ctx, dag.Host(), "DOCKERHUB_USERNAME"))
	dockerhubPassword := env.HostEnv(ctx, dag.Host(), "DOCKERHUB_PASSWORD").Secret()

	dockerAuthContainer := docker.Container(ctx, dag).
		WithEnvVariable("DOCKERHUB_USERNAME", dockerhubUsername).
		WithMountedSecret("/var/run/secret/dockerhub/password", dockerhubPassword).
		WithExec([]string{
			"sh", "-c",
			"docker login --username $DOCKERHUB_USERNAME --password $(cat /var/run/secret/dockerhub/password)",
		})

	_, err := dockerAuthContainer.File("/root/.docker/config.json").Export(ctx, hostDockerClientConfig)
	if err != nil {
		return err
	}

	return nil
}
