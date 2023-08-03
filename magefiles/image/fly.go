package image

import (
	"fmt"
	"time"

	"os"
)

func (image *Image) Deploy() *Image {
	githubRepo := os.Getenv("GITHUB_REPOSITORY")
	githubRef := os.Getenv("GITHUB_REF_NAME")

	fmt.Printf("🔍 githubRepo: %s\n", githubRepo)
	fmt.Printf("🔍 githubRef: %s\n", githubRef)

	image = image.flyctl().app()

	if githubRepo != RootRepository {
		fmt.Printf("\n👮 Deploys only run on %s repo\n", RootRepository)
		return image
	}

	if githubRef != MainBranch {
		fmt.Printf("\n👮 Deploys only run on %s branch\n", MainBranch)
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

	return image.OK()
}

func (image *Image) DaggerStart() *Image {
	image = image.flyctl().dagger()
	var err error

	primaryEngineMachineID := os.Getenv("FLY_PRIMARY_DAGGER_ENGINE_MACHINE_ID")
	if primaryEngineMachineID == "" {
		fmt.Printf(
			"👮 Skip starting Dagger Engine, FLY_PRIMARY_ENGINE_MACHINE_ID env var is missing\n",
		)
		return image
	}

	image, err = image.startMachine(primaryEngineMachineID)
	if err != nil {
		secondaryEngineMachineID := os.Getenv("FLY_SECONDARY_DAGGER_ENGINE_MACHINE_ID")
		if secondaryEngineMachineID == "" {
			fmt.Printf(
				"👮 Skip starting Dagger Engine, FLY_SECONDARY_DAGGER_ENGINE_MACHINE_ID env var is missing\n",
			)
			return image
		}

		image, err = image.startMachine(secondaryEngineMachineID)
		mustCreate(err)
	}

	return image
}

func (image *Image) DaggerStop() *Image {
	image = image.flyctl().dagger()
	var err error

	primaryEngineMachineID := os.Getenv("FLY_PRIMARY_DAGGER_ENGINE_MACHINE_ID")
	if primaryEngineMachineID == "" {
		fmt.Printf(
			"👮 Skip stopping Dagger Engine, FLY_PRIMARY_ENGINE_MACHINE_ID env var is missing\n",
		)
		return image
	}

	image, err = image.stopMachine(primaryEngineMachineID)
	mustCreate(err)

	secondaryEngineMachineID := os.Getenv("FLY_SECONDARY_DAGGER_ENGINE_MACHINE_ID")
	if secondaryEngineMachineID == "" {
		fmt.Printf(
			"👮 Skip stopping Dagger Engine, FLY_SECONDARY_DAGGER_ENGINE_MACHINE_ID env var is missing\n",
		)
		return image
	}

	image, err = image.stopMachine(secondaryEngineMachineID)
	mustCreate(err)

	return image
}

func (image *Image) flyctl() *Image {
	FLY_API_TOKEN := image.dag.SetSecret("FLY_API_TOKEN", os.Getenv("FLY_API_TOKEN"))

	image.container = image.NewContainer().
		From(image.flyctlImageRef()).
		WithSecretVariable("FLY_API_TOKEN", FLY_API_TOKEN).
		WithExec([]string{
			"version",
		})

	return image
}

func (image *Image) app() *Image {
	image.container = image.container.
		WithMountedFile("fly.toml", image.dag.Host().Directory("2022.fly").File("fly.toml"))

	return image
}

func (image *Image) dagger() *Image {
	image.container = image.container.
		WithMountedFile("fly.toml", image.dag.Host().Directory("fly.io/dagger-engine-2023-08-12").File("fly.toml"))

	return image
}

func (image *Image) startMachine(id string) (*Image, error) {
	var err error

	image.container, err = image.container.
		WithEnvVariable("CACHE_BUSTED_AT", time.Now().String()).
		WithExec([]string{
			"machine",
			"start", id,
		}).
		WithExec([]string{
			"machine",
			"status", id,
		}).
		Sync(image.ctx)

	return image, err
}

func (image *Image) stopMachine(id string) (*Image, error) {
	var err error

	image.container, err = image.container.
		WithEnvVariable("CACHE_BUSTED_AT", time.Now().String()).
		WithExec([]string{
			"machine",
			"stop", id,
		}).
		WithExec([]string{
			"machine",
			"status", id,
		}).
		Sync(image.ctx)

	return image, err
}

func (image *Image) flyctlImageRef() string {
	return fmt.Sprintf("flyio/flyctl:v%s", image.versions.Flyctl())
}
