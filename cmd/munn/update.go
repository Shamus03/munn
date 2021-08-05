package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Download the latest release from GitHub",
	Long:  `Download the latest release from GitHub and install it in-place.`,
	RunE: func(cmd *cobra.Command, args []string) (cmdError error) {
		ctx := context.Background()
		owner := "Shamus03"
		repo := "munn"

		client := github.NewClient(nil)

		release, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
		if err != nil {
			return fmt.Errorf("getting latest release: %v", err)
		}
		currentVersion := cmd.Root().Version

		if release.GetName() == currentVersion {
			cmd.Printf("Already up to date\n")
			return nil
		}

		cmd.Printf("Updating to %s\n", release.GetName())

		var assetID int64
		for _, asset := range release.Assets {
			if strings.Contains(asset.GetName(), runtime.GOOS) {
				assetID = asset.GetID()
				break
			}
		}

		if assetID == 0 {
			return fmt.Errorf("could not find a suitable release to download")
		}

		assetReader, _, err := client.Repositories.DownloadReleaseAsset(ctx, owner, repo, assetID, http.DefaultClient)
		if err != nil {
			return fmt.Errorf("download release asset: %v", err)
		}
		defer assetReader.Close()

		outPath := os.Args[0]

		tmpDir, err := os.MkdirTemp("", "munn-bak-")
		if err != nil {
			return fmt.Errorf("make backup dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		bakFile := filepath.Join(tmpDir, filepath.Base(outPath))

		if err := os.Rename(outPath, bakFile); err != nil {
			return fmt.Errorf("rename old executable: %v", err)
		}
		defer func() {
			if cmdError != nil {
				if err := os.Rename(outPath, bakFile); err != nil {
					cmd.Printf("ERROR: error while updating, and failed to restore backup file: %v", err)
				}
			}
		}()

		f, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("create new executable: %v", err)
		}
		defer f.Close()

		if _, err := io.Copy(f, assetReader); err != nil {
			return fmt.Errorf("write new executable: %v", err)
		}

		return nil
	},
}
