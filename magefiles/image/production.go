package image

import (
	"fmt"
	"os"
	"time"

	"dagger.io/dagger"
	"github.com/magefile/mage/sh"
)

func (image *Image) Production() *Image {
	productionImage := image.Runtime().
		WithAppSrc().
		WithAppDeps().
		WithProdEnv().
		WithAppCompiled().
		WithAppStaticAssets().
		WithAppLegacyAssets().
		WithGitAuthor().
		WithGitSHA().
		WithBuildURL().
		OK()

	if os.Getenv("DEBUG") != "" {
		_, err := productionImage.container.Export(image.ctx, "tmp/app.test.prod.tar")
		mustCreate(err)
	}

	return productionImage
}

// ProductionClean() is a TEMPORARY function, until Elixir releases are implemented
// === WHY?
// Jerod Santo - 8:35 PM - Mar. 15th, 2023
// @gerhard hmm 5 failed ship it runs in a row…
// I don’t think this is sufferable much longer
// https://github.com/thechangelog/changelog.com/actions/runs/4430462525
func (image *Image) ProductionClean() *Image {
	app := image.Production().container.
		WithExec([]string{
			"cp", "--recursive", "--preserve=mode,ownership,timestamps", "/app", "/app.prod",
		}).
		// 👉 TODO: multiple arguments to rm -fr do not work as expected
		WithExec([]string{
			"rm", "-fr", "/app.prod/_build/test",
		}).
		WithExec([]string{
			"rm", "-fr", "/app.prod/assets",
		}).
		WithExec([]string{
			"rm", "-fr", "/app.prod/test",
		}).
		Directory("/app.prod")

	image = image.
		Elixir().
		WithAptPackages().
		WithGit().
		WithImagemagick().
		WithProdEnv()

	if os.Getenv("R2_ACCESS_KEY_ID") != "" {
		fmt.Printf("⚡️ Uploading static assets...")
		image = image.UploadStaticAssets()
	}

	image.container = image.container.
		WithDirectory("/app", app).
		WithWorkdir("/app").
		WithExec([]string{
			"sh", "-c", "echo \"Ensure $MIX_ENV bytecode is present & OK...\"",
		}).
		WithExec([]string{
			"sh", "-c", "ls -lahd /app/_build/$MIX_ENV/lib/*/ebin",
		})

	image = image.OK()

	if os.Getenv("DEBUG") != "" {
		_, err := image.container.Export(image.ctx, "tmp/app.prod.tar")
		mustCreate(err)
	}

	return image
}

func (image *Image) PublishProduction() *Image {
	return image.
		WithProductionLabels().
		Publish(image.ProductionImageRef())
}

func (image *Image) WithProductionLabels() *Image {
	description := fmt.Sprintf(
		"🖥 %s",
		buildURL(),
	)

	image.container = image.container.
		WithLabel("org.opencontainers.image.description", description).
		WithLabel("org.opencontainers.image.source", RepositoryURL)

	return image
}

func (image *Image) ProductionImageRef() string {
	imageOwner := os.Getenv("IMAGE_OWNER")
	if imageOwner == "" {
		imageOwner = os.Getenv("GITHUB_ACTOR")
	}

	return fmt.Sprintf(
		"ghcr.io/%s/changelog-prod:%s",
		imageOwner,
		gitSHA(),
	)
}

// 👉 TODO:
// 1. Upload legacy assets
// 2. /wp-content/** redirect
func (image *Image) UploadStaticAssets() *Image {
	R2_API_HOST := image.dag.SetSecret("R2_API_HOST", os.Getenv("R2_API_HOST"))
	R2_ACCESS_KEY_ID := image.dag.SetSecret("R2_ACCESS_KEY_ID", os.Getenv("R2_ACCESS_KEY_ID"))
	R2_SECRET_ACCESS_KEY := image.dag.SetSecret("R2_SECRET_ACCESS_KEY", os.Getenv("R2_SECRET_ACCESS_KEY"))

	_, err := image.Production().
		// 🤔 Why do we need to start the app - and therefore require the DB - to upload static assets?
		// 🚨👇 This leaves behind envs such as DB_HOST, DB_NAME, DB_USER, DB_PASS
		WithPostgreSQL("changelog_prod").
		container.
		// AT busts the cache so that...
		WithEnvVariable("AT", time.Now().String()).
		// ... we always run the db migration - the DB container doesn't keep data
		WithExec([]string{
			"mix", "ecto.migrate",
		}).
		WithSecretVariable("R2_API_HOST", R2_API_HOST).
		WithEnvVariable("R2_ASSETS_BUCKET", "changelog-assets").
		WithSecretVariable("R2_ACCESS_KEY_ID", R2_ACCESS_KEY_ID).
		WithSecretVariable("R2_SECRET_ACCESS_KEY", R2_SECRET_ACCESS_KEY).
		WithExec([]string{
			"mix", "changelog.static.upload",
		}).
		Sync(image.ctx)

	mustCreate(err)

	return image
}

func (image *Image) WithProdEnv() *Image {
	image.container = image.container.
		WithEnvVariable("MIX_ENV", "prod")

	return image
}

func (image *Image) WithGitAuthor() *Image {
	image.container = image.container.
		WithNewFile("COMMIT_USER", dagger.ContainerWithNewFileOpts{
			Contents:    gitAuthor(),
			Permissions: 444,
		})

	return image
}

func gitAuthor() string {
	author := os.Getenv("GITHUB_ACTOR")
	if author == "" {
		author = os.Getenv("USER")
	}

	return author
}

func (image *Image) WithGitSHA() *Image {
	image.container = image.container.
		WithNewFile("priv/static/version.txt", dagger.ContainerWithNewFileOpts{
			Contents:    gitSHA(),
			Permissions: 444,
		})

	return image
}

func gitSHA() string {
	gitSHA := os.Getenv("GITHUB_SHA")
	if gitSHA == "" {
		if gitHEAD, err := sh.Output("git", "rev-parse", "HEAD"); err == nil {
			gitSHA = fmt.Sprintf("%s.", gitHEAD)
		}
		gitSHA = fmt.Sprintf("%sdev", gitSHA)
	}

	return gitSHA
}

func (image *Image) WithBuildURL() *Image {
	image.container = image.container.
		WithNewFile("priv/static/build.txt", dagger.ContainerWithNewFileOpts{
			Contents:    buildURL(),
			Permissions: 444,
		})

	return image
}

func buildURL() string {
	githubServerURL := os.Getenv("GITHUB_SERVER_URL")
	githubRepository := os.Getenv("GITHUB_REPOSITORY")
	githubRunID := os.Getenv("GITHUB_RUN_ID")
	buildURL := fmt.Sprintf("%s/%s/actions/runs/%s", githubServerURL, githubRepository, githubRunID)

	if githubRunID == "" {
		if hostname, err := os.Hostname(); err == nil {
			buildURL = hostname
		}
		if cwd, err := os.Getwd(); err == nil {
			buildURL = fmt.Sprintf("%s:%s", buildURL, cwd)
		}
	}

	return buildURL
}
