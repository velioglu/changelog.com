package image

import (
	"fmt"

	"dagger.io/dagger"
)

const (
	// https://hub.docker.com/r/flyio/flyctl/tags
	flyctlVersion = "0.0.492"
)

func (image *Image) Deploy() *Image {
	githubRepo := image.Env("GITHUB_REPOSITORY")
	githubRef := image.Env("GITHUB_REF_NAME")

	image = image.flyctl()

	if githubRepo.Value() != RootRepository && githubRef.Value() != MainBranch {
		fmt.Printf("\n👮 Deploy only runs on %s repo, %s branch\n", RootRepository, MainBranch)
		return image
	}

	image.container = image.container.
		WithExec([]string{
			"status",
		}).
		WithExec([]string{
			"deploy",
			"--image", image.ProductionImageRef(),
		})

	return image
}

func (image *Image) flyctl() *Image {
	image.container = image.NewContainer().
		From(flyctlImageRef()).
		WithSecretVariable("FLY_API_TOKEN", image.Env("FLY_API_TOKEN").Secret()).
		WithMountedFile("fly.toml", image.flyConfig()).
		WithExec([]string{
			"version",
		})

	return image
}

func flyctlImageRef() string {
	return fmt.Sprintf("flyio/flyctl:v%s", flyctlVersion)
}

func (image *Image) flyConfig() *dagger.File {
	return image.dag.Host().Directory("2022.fly").File("fly.toml")
}
