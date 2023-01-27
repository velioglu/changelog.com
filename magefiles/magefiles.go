//go:build mage
// +build mage

package main

import (
	//mage:import
	_ "github.com/thechangelog/changelog.com/magefiles/kosli"
	//mage:import
	_ "github.com/thechangelog/changelog.com/magefiles/legacy"
	//mage:import
	_ "github.com/thechangelog/changelog.com/magefiles/dockerhub"
)
