// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Implements nodejs/firebaseangular buildpack.
// The nodejs/firebaseangular buildpack does some prep work for angular and runs the build script.
package main

import (
	"encoding/json"

	"github.com/GoogleCloudPlatform/buildpacks/pkg/nodejs"
	// "google3/third_party/golang/hashicorp/version/version"
	"github.com/Masterminds/semver"

	gcp "github.com/GoogleCloudPlatform/buildpacks/pkg/gcpbuildpack"
)

var (
	// minAngularVersion is the lowest version of angular supported by the firebase angular buildpack.
	minAngularVersion = semver.MustParse("17.0.0")
)

func main() {
	gcp.Main(detectFn, buildFn)
}

func detectFn(ctx *gcp.Context) (gcp.DetectResult, error) {
	// b/319754948
	// In monorepo scenarios, we'll probably need to support environment variable that can be used to
	// know where the application directory is located within the repository.
	angularJSONExists, err := ctx.FileExists("angular.json")
	if err != nil {
		return nil, err
	}
	if angularJSONExists {
		return gcp.OptInFileFound("angular.json"), nil
	}
	return gcp.OptOut("angular config not found"), nil
}

func buildFn(ctx *gcp.Context) error {
	pjs, err := nodejs.ReadPackageJSONIfExists(ctx.ApplicationRoot())
	if err != nil {
		return err
	}
	version := version(ctx, pjs)

	err = validateVersion(ctx, version)
	if err != nil {
		return err
	}

	buildScript, exists := pjs.Scripts["build"]
	if exists && buildScript != "ng build" && buildScript != "apphosting-adapter-angular-build" {
		ctx.Warnf("*** You are using a custom build command (your build command is NOT 'ng build'), we will accept it as is but will error if output structure is not as expected ***")
	}

	njsl, err := ctx.Layer("npm_modules", gcp.BuildLayer, gcp.CacheLayer)
	if err != nil {
		return err
	}
	err = nodejs.InstallAngularBuildAdaptor(ctx, njsl)
	if err != nil {
		return err
	}
	// This env var indicates to the package manager buildpack that a different command needs to be run
	nodejs.OverrideAngularBuildScript(njsl)

	return nil
}

func validateVersion(ctx *gcp.Context, depVersion string) error {
	version, err := semver.NewVersion(depVersion)
	if err != nil {
		ctx.Warnf("Unrecognized version of angular: %s", depVersion)
		ctx.Warnf("Consider updating your angular dependencies to >=%s", minAngularVersion.String())
		return nil
	}
	if version.LessThan(minAngularVersion) {
		ctx.Warnf("Unsupported version of angular: %s", depVersion)
		ctx.Warnf("Update the angular dependencies to >=%s", minAngularVersion.String())
		return gcp.UserErrorf("unsupported version of angular %s", depVersion)
	}
	return nil
}

type angularVersionJSON struct {
	Version      string `json:"version"`
	Name         string `json:"name"`
	Dependencies struct {
		Angular struct {
			Version    string `json:"version"`
			Resolved   string `json:"resolved"`
			Overridden bool   `json:"overridden"`
		} `json:"angular"`
	} `json:"dependencies"`
}

// tries to get the concrete angular version used, otherwise falls back on package.json
func version(ctx *gcp.Context, pjs *nodejs.PackageJSON) string {
	result, err := ctx.Exec([]string{"npm", "list", "@angular/cli", "--json"})
	if err != nil {
		return pjs.Dependencies["@angular/cli"]
	}
	var angularVersionnpm angularVersionJSON
	if err := json.Unmarshal([]byte(result.Stdout), &angularVersionnpm); err != nil {
		return pjs.Dependencies["@angular/cli"]
	}
	return angularVersionnpm.Dependencies.Angular.Version
}
