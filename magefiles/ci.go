//go:build mage
// +build mage

package main

import (
	"context"
	"os"

	"dagger.io/dagger"

	"github.com/thechangelog/changelog.com/magefiles/dockerhub"
	"github.com/thechangelog/changelog.com/magefiles/kosli"
	"github.com/thechangelog/changelog.com/magefiles/legacy"
)

// Run the entire CI pipeline
func CI(ctx context.Context) error {
	dag, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer dag.Close()

	err = kosli.Environment(ctx, dag)
	if err != nil {
		return err
	}

	err = kosli.Pipeline(ctx, dag)
	if err != nil {
		return err
	}

	err = legacy.ShipIt(ctx, dag)
	if err != nil {
		return err
	}

	_, err = dockerhub.Publish(ctx, dag)
	if err != nil {
		return err
	}

	//  err = kosli.Image(ctx, dag)
	// if err != nil {
	// 	return err
	// }

	//  err = kosli.Deployment(ctx, dag)
	// if err != nil {
	// 	return err
	// }

	return nil
}
