// Copyright 2020 Google LLC
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
package acceptance

import (
	"testing"

	"github.com/GoogleCloudPlatform/buildpacks/internal/acceptance"
)

func init() {
	acceptance.DefineFlags()
}

const (
	npm       = "google.nodejs.npm"
	yarn      = "google.nodejs.yarn"
	appEngine = "google.nodejs.appengine"
)

func TestAcceptance(t *testing.T) {
	builderImage, runImage, cleanup := acceptance.ProvisionImages(t)
	t.Cleanup(cleanup)

	testCases := []struct {
		test                       acceptance.Test
		versionInclusionConstraint string
	}{
		{
			test: acceptance.Test{
				App:        "no_package_json",
				MustUse:    []string{appEngine},
				MustNotUse: []string{npm, yarn},
			},
		},
		{
			test: acceptance.Test{
				App:        "package_json_no_dependencies",
				MustUse:    []string{npm},
				MustNotUse: []string{yarn},
			},
		},
		{
			test: acceptance.Test{
				App:        "package_json",
				MustUse:    []string{npm},
				MustNotUse: []string{yarn},
			},
		},
		{
			test: acceptance.Test{
				App:        "package_lock",
				MustUse:    []string{npm},
				MustNotUse: []string{yarn},
			},
		},
		{
			test: acceptance.Test{
				App:        "npm_shrinkwrap",
				MustUse:    []string{npm},
				MustNotUse: []string{yarn},
			},
		},
		{
			test: acceptance.Test{
				App:        "yarn_lock",
				MustUse:    []string{yarn},
				MustNotUse: []string{npm},
			},
		},
		{
			test: acceptance.Test{
				App:        "yarn_lock_before_package_lock",
				MustUse:    []string{yarn},
				MustNotUse: []string{npm},
			},
		},
		{
			test: acceptance.Test{
				App:        "postinstall_npm",
				MustUse:    []string{npm},
				MustNotUse: []string{yarn},
				MustNotOutputCached: []string{
					"removing existing node_modules/ before installation",
					"added 1 package",
				},
			},
		},
		{
			test: acceptance.Test{
				App:              "postinstall_yarn",
				MustUse:          []string{yarn},
				MustNotUse:       []string{npm},
				MustOutputCached: []string{"Already up-to-date."},
			},
		},
		{
			test: acceptance.Test{
				App:        "gcp_build_npm",
				MustUse:    []string{npm},
				MustNotUse: []string{yarn},
			},
		},
		{
			test: acceptance.Test{
				App:        "gcp_build_yarn",
				MustUse:    []string{yarn},
				MustNotUse: []string{npm},
			},
		},
		{
			test: acceptance.Test{
				App:        "gcp_build_npm_no_dependencies",
				MustUse:    []string{npm},
				MustNotUse: []string{yarn},
			},
		},
		{
			test: acceptance.Test{
				App:        "gcp_build_yarn_no_dependencies",
				MustUse:    []string{yarn},
				MustNotUse: []string{npm},
			},
		},
		{
			test: acceptance.Test{
				App:        "local_dependency_npm",
				MustUse:    []string{npm},
				MustNotUse: []string{yarn},
			},
		},
		{
			test: acceptance.Test{
				App:        "local_dependency_yarn",
				MustUse:    []string{yarn},
				MustNotUse: []string{npm},
			},
		},
		{
			test: acceptance.Test{
				App:        "dev_dependency_npm",
				Env:        []string{"NODE_ENV=development"},
				MustUse:    []string{npm},
				MustNotUse: []string{yarn},
			},
		},
		{
			test: acceptance.Test{
				App:        "dev_dependency_yarn",
				Env:        []string{"NODE_ENV=development"},
				MustUse:    []string{yarn},
				MustNotUse: []string{npm},
			},
		},
		{
			test: acceptance.Test{
				App:        "node_modules_bin_npm",
				Path:       "/index.txt",
				MustUse:    []string{npm},
				MustNotUse: []string{yarn},
			},
		},
		{
			test: acceptance.Test{
				Name:       "node_modules/.bin custom entrypoint",
				App:        "node_modules_bin_npm",
				Path:       "/index.txt",
				Env:        []string{"GOOGLE_ENTRYPOINT=http-server -p $PORT"},
				MustUse:    []string{npm},
				MustNotUse: []string{yarn},
			},
		},
		{
			test: acceptance.Test{
				App:        "node_modules_bin_yarn",
				Path:       "/index.txt",
				MustUse:    []string{yarn},
				MustNotUse: []string{npm},
			},
		},
		{
			test: acceptance.Test{
				Name:       "node_modules/.bin yarn custom entrypoint",
				App:        "node_modules_bin_yarn",
				Path:       "/index.txt",
				Env:        []string{"GOOGLE_ENTRYPOINT=http-server -p $PORT"},
				MustUse:    []string{yarn},
				MustNotUse: []string{npm},
			},
		},
		{
			test: acceptance.Test{
				App:        "custom_entrypoint",
				Env:        []string{"GOOGLE_ENTRYPOINT=node custom.js"},
				MustUse:    []string{npm},
				MustNotUse: []string{yarn},
			},
		},
		{
			// yarn_two uses worker_threads which require node.js@12+
			versionInclusionConstraint: ">= 12.0.0",
			test: acceptance.Test{
				App:        "yarn_two",
				MustUse:    []string{yarn},
				MustNotUse: []string{npm},
			},
		},
		{
			// yarn_two_pnp uses worker_threads which require node.js@12+
			versionInclusionConstraint: ">= 12.0.0",
			test: acceptance.Test{
				App:        "yarn_two_pnp",
				MustUse:    []string{yarn},
				MustNotUse: []string{npm},
				Env:        []string{"GOOGLE_ENTRYPOINT=yarn start"},
			},
		},
	}
	for _, tc := range testCases {
		if !acceptance.ShouldTestVersion(t, tc.versionInclusionConstraint) {
			continue
		}
		tc := applyStaticTestOptions(tc.test)
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			acceptance.TestApp(t, builderImage, runImage, tc)
		})
	}
}

func applyStaticTestOptions(tc acceptance.Test) acceptance.Test {
	tc.Env = append(tc.Env, "X_GOOGLE_TARGET_PLATFORM=gae")
	if tc.Name == "" {
		tc.Name = tc.App
	}
	return tc
}