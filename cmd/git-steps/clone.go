//  Copyright 2024 Google LLC
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/GoogleCloudBuild/cicd-images/internal/logger"
	"github.com/spf13/cobra"
)

var (
	subDirectory   string
	deleteExisting string
	depth          string
	revision       string
	submodules     string
	verbose        bool
)

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone a public or private git repository.",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.SetupLogger(verbose)
		slog.Info("Executing clone command")

		// extract url from urlPath
		url, err := os.ReadFile(urlPath)
		if err != nil {
			return fmt.Errorf("error reading urlPath: %v", err)
		}
		slog.Debug("Extracted URL from urlPath", "url", string(url))

		checkoutDir := ""
		if (len(subDirectory) != 0) && (deleteExisting == "true") {
			slog.Debug("Deleting existing repo directory")
			checkoutDir = subDirectory
			// 	Delete any existing contents of the repo directory if it exists.
			//  We don't just "rm -rf ${CHECKOUT_DIR}" because ${CHECKOUT_DIR} might be "/" or the root of a mounted volume.
			if stat, err := os.Stat(checkoutDir); err == nil && stat.IsDir() {
				// Delete non-hidden files and directories
				removeNonHiddenFilesCmd := fmt.Sprintf("-rf %v/*", checkoutDir)
				rmNonHiddenFiles := exec.Command("rm", strings.Fields(removeNonHiddenFilesCmd)...)
				if err := rmNonHiddenFiles.Run(); err != nil {
					return fmt.Errorf("error deleting non-hidden files and directories: %v", err)
				}
				// Delete files and directories starting with . but excluding ..
				removeHiddenFilesCmd := fmt.Sprintf("-rf %v/.[!.]*", checkoutDir)
				rmHiddenFiles := exec.Command("rm", strings.Fields(removeHiddenFilesCmd)...)
				if err := rmHiddenFiles.Run(); err != nil {
					return fmt.Errorf("error deleting files and directories starting with . but excluding ..: %v", err)
				}
				// Delete files and directories starting with .. plus any other character
				removeOtherFilesCmd := fmt.Sprintf("-rf %v/..?*", checkoutDir)
				rmOtherFiles := exec.Command("rm", strings.Fields(removeOtherFilesCmd)...)
				if err := rmOtherFiles.Run(); err != nil {
					return fmt.Errorf("error deleting files and directories starting with .. plus any other character: %v", err)
				}
			}
		}

		// git init "${CHECKOUT_DIR}"
		initCommand := fmt.Sprintf("init %v", checkoutDir)
		gitInit := exec.Command("git", strings.Fields(initCommand)...)
		if err := gitInit.Run(); err != nil {
			return fmt.Errorf("error running 'git %v': %v", initCommand, err)
		}
		slog.Info("Initialized git repository")

		if checkoutDir != "" {
			// change to initialized git directory
			err = os.Chdir(checkoutDir)
			if err != nil {
				return fmt.Errorf("error running 'cd %v': %v", checkoutDir, err)
			}
			slog.Info("Changed to directory", "directory", checkoutDir)
		}

		// git remote add origin "${URL}"
		remoteAddCommand := fmt.Sprintf("remote add origin %v", string(url))
		gitRemoteAdd := exec.Command("git", strings.Fields(remoteAddCommand)...)
		if err = gitRemoteAdd.Run(); err != nil {
			return fmt.Errorf("error running 'git %v': %v", remoteAddCommand, err)
		}
		slog.Debug("Added remote origin", "url", string(url))

		// fetch branch/tag/sha
		fetchCommand := "fetch --all"
		if len(depth) != 0 { // if depth specified
			fetchCommand = fmt.Sprintf("fetch --depth %v --all", depth)
		}
		gitFetch := exec.Command("git", strings.Fields(fetchCommand)...)
		if err = gitFetch.Run(); err != nil {
			return fmt.Errorf("error running 'git %v': %v", fetchCommand, err)
		}
		slog.Info("Fetched all branches/tags/shas")

		// git checkout -f "${REVISION}"
		checkoutCommand := fmt.Sprintf("checkout -f %v", revision)
		gitCheckout := exec.Command("git", strings.Fields(checkoutCommand)...)
		if err = gitCheckout.Run(); err != nil {
			return fmt.Errorf("error running 'git %v': %v", checkoutCommand, err)
		}
		slog.Info("Checked out revision", "revision", revision)

		// update submodules if true: git submodule update --init --recursive
		if submodules == "true" {
			submodulesCommand := "submodule update --init --recursive"
			gitUpdateSubmodules := exec.Command("git", strings.Fields(submodulesCommand)...)
			if err = gitUpdateSubmodules.Run(); err != nil {
				return fmt.Errorf("error running 'git %v': %v", submodulesCommand, err)
			}
			slog.Info("Updated submodules")
		}

		slog.Info("Successfully cloned git repository")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)

	cloneCmd.Flags().StringVar(&subDirectory, "subDirectory", "", "The subdirectory to clone the git repo into.")
	cloneCmd.Flags().StringVar(&deleteExisting, "deleteExisting", "", "Bool string to delete the existing repo or not.")
	cloneCmd.Flags().StringVar(&urlPath, "urlPath", "", "The path to store the extracted URL of the repository.")
	cloneCmd.Flags().StringVar(&depth, "depth", "", "Depth level to perform git clones. If unspecified, a full clone will be performed.")
	cloneCmd.Flags().StringVar(&revision, "revision", "", "Revision (branch, tag, or commit sha) to checkout.")
	cloneCmd.Flags().StringVar(&submodules, "submodules", "", "Initialize and fetch git submodules.")
}
