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

package main

import (
	"testing"

	bpt "github.com/GoogleCloudPlatform/buildpacks/internal/buildpacktest"
	"github.com/GoogleCloudPlatform/buildpacks/internal/mockprocess"
)

func TestDetect(t *testing.T) {
	testCases := []struct {
		name  string
		files map[string]string
		want  int
	}{
		{
			name: "with angular config",
			files: map[string]string{
				"index.js":     "",
				"angular.json": "",
			},
			want: 0,
		},
		{
			name: "without angular config",
			files: map[string]string{
				"index.js": "",
			},
			want: 100,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bpt.TestDetect(t, detectFn, tc.name, tc.files, []string{}, tc.want)
		})
	}
}

func TestBuild(t *testing.T) {
	testCases := []struct {
		name         string
		wantExitCode int
		wantCommands []string
		// opts          []bpt.Option
		mocks         []*mockprocess.Mock
		files         map[string]string
		filesExpected map[string]string
	}{
		{
			name: "replace build script",
			files: map[string]string{
				"package.json": `{
				"scripts": {
					"build": "ng build"
				},
				"dependencies": {
					"angular": "17.0.0"
				}
			}`,
			},
			mocks: []*mockprocess.Mock{
				mockprocess.New(`npm install --prefix npm_modules @apphosting/adapter-angular@latest`, mockprocess.WithStdout("installed adaptor")),
			},
			wantCommands: []string{
				"npm install --prefix npm_modules @apphosting/adapter-angular@latest",
			},
		},
		{
			name: "build script doesnt exist",
			files: map[string]string{
				"package.json": `{
					"dependencies": {
						"angular": "17.0.0"
					}
				}`,
			},
			mocks: []*mockprocess.Mock{
				mockprocess.New(`npm install --prefix npm_modules @apphosting/adapter-angular@latest`, mockprocess.WithStdout("installed adaptor")),
			},
		},
		{
			name: "build script already set",
			files: map[string]string{
				"package.json": `{
					"scripts": {
						"build": "apphosting-adapter-angular-build"
					},
					"dependencies": {
						"angular": "17.0.0"
					}
				}`,
			},
			mocks: []*mockprocess.Mock{
				mockprocess.New(`npm install --prefix npm_modules @apphosting/adapter-angular@latest`, mockprocess.WithStdout("installed adaptor")),
			},
		},
		{
			name: "error out if the version is below 17.0.0",
			files: map[string]string{
				"package.json": `{
				"dependencies": {
					"angular": "16.0.0"
				}
			}`,
			},
			wantExitCode: 1,
		},
		{
			name: "should work if the version is exactly 17.0.0",
			files: map[string]string{
				"package.json": `{
				"dependencies": {
					"angular": "17.0.0"
				}
			}`,
			},
			mocks: []*mockprocess.Mock{
				mockprocess.New(`npm install --prefix npm_modules @apphosting/adapter-angular@latest`, mockprocess.WithStdout("installed adaptor")),
			},
			wantExitCode: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := []bpt.Option{
				bpt.WithTestName(tc.name),
				bpt.WithFiles(tc.files),
				bpt.WithExecMocks(tc.mocks...),
			}
			result, err := bpt.RunBuild(t, buildFn, opts...)
			if err != nil && tc.wantExitCode == 0 {
				t.Fatalf("error running build: %v, logs: %s", err, result.Output)
			}

			if result.ExitCode != tc.wantExitCode {
				t.Errorf("build exit code mismatch, got: %d, want: %d", result.ExitCode, tc.wantExitCode)
			}

			for _, cmd := range tc.wantCommands {
				if !result.CommandExecuted(cmd) {
					t.Errorf("expected command %q to be executed, but it was not, build output: %s", cmd, result.Output)
				}
			}
		})
	}
}
