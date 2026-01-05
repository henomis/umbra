package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/henomis/umbra/config"
	"github.com/henomis/umbra/internal/ghost"
	"github.com/henomis/umbra/internal/provider"
	"github.com/henomis/umbra/umbra"
)

const version = "0.0.1"

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "umbra",
	Short: "Umbra securely split, encrypt, and redundantly store files across pluggable providers via CLI.",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(version)
	},
}

var providersCmd = &cobra.Command{
	Use:     "providers",
	Aliases: []string{"p"},
	Short:   "List available storage providers",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("Available providers:")
		for _, p := range provider.DefaultProviders {
			fmt.Printf("  - %s\n", p)
		}
	},
}

/*
 * =====================
 * Upload Command
 * =====================
 */

var (
	uploadFile string
	password   string
	chunkSize  int64
	chunks     int
	copies     int
	providers  []string
	// rawOptions   []string // for future use.
	// options      map[string]string // for future use.
	outputFile   string
	manifestPath string
	quiet        bool
	ghostMode    string
)

var infoCmd = &cobra.Command{
	Use:     "info",
	Aliases: []string{"i"},
	Short:   "Display manifest information",
	Run: func(_ *cobra.Command, _ []string) {
		cfg := &config.Config{
			ManifestPath: manifestPath,
			Password:     password,
			GhostMode:    ghostMode,
		}

		umbraInstance, err := umbra.New(cfg)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err := umbraInstance.Info(context.Background()); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var uploadCmd = &cobra.Command{
	Use:     "upload",
	Aliases: []string{"u"},
	Short:   "Upload a file",
	PreRunE: func(_ *cobra.Command, _ []string) error {
		// For future use
		// parsed, err := parseKeyValueOptions(rawOptions)
		// if err != nil {
		// 	return err
		// }

		// options = parsed

		// Validate ghost mode
		if ghostMode != "" && !ghost.IsValidGhostMode(ghostMode) {
			return fmt.Errorf("invalid ghost mode %q: must be one of %s", ghostMode, strings.Join(ghost.GhostModes(), ", "))
		}

		return nil
	},
	Run: func(_ *cobra.Command, _ []string) {
		cfg := &config.Config{
			ManifestPath: manifestPath,
			Password:     password,
			Quiet:        quiet,
			Providers:    providers,
			// Options:      options, // for future use
			GhostMode: ghostMode,
			Upload: &config.Upload{
				InputFilePath: uploadFile,
				ChunkSize:     chunkSize,
				Chunks:        chunks,
				Copies:        copies,
			},
		}

		umbraInstance, err := umbra.New(cfg)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err := umbraInstance.Upload(context.Background()); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

// For future use
// func parseKeyValueOptions(input []string) (map[string]string, error) {
// 	result := make(map[string]string)

// 	for _, item := range input {
// 		parts := strings.SplitN(item, "=", 2)
// 		if len(parts) != 2 || parts[0] == "" {
// 			return nil, fmt.Errorf("invalid option %q, expected key=value", item)
// 		}
// 		result[parts[0]] = parts[1]
// 	}

// 	return result, nil
// }

/*
 * =====================
 * Download Command
 * =====================
 */

var downloadCmd = &cobra.Command{
	Use:     "download",
	Aliases: []string{"d"},
	Short:   "Download a file using a manifest",
	PreRunE: func(_ *cobra.Command, _ []string) error {
		// For future use
		// parsed, err := parseKeyValueOptions(rawOptions)
		// if err != nil {
		// 	return err
		// }
		// options = parsed

		// Validate ghost mode
		if ghostMode != "" && !ghost.IsValidGhostMode(ghostMode) {
			return fmt.Errorf("invalid ghost mode %q: must be one of %s", ghostMode, strings.Join(ghost.GhostModes(), ", "))
		}

		return nil
	},
	Run: func(_ *cobra.Command, _ []string) {
		cfg := &config.Config{
			ManifestPath: manifestPath,
			Password:     password,
			Quiet:        quiet,
			// Options:      options, // for future use
			GhostMode: ghostMode,
			Download: &config.Download{
				OutputFilePath: outputFile,
			},
		}

		umbraInstance, err := umbra.New(cfg)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err := umbraInstance.Download(context.Background()); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	/*
	 * Upload flags
	 */
	uploadCmd.Flags().StringVarP(&uploadFile, "file", "f", "", "specify file to upload")
	uploadCmd.Flags().StringVarP(&password, "password", "p", "", "specify password")
	uploadCmd.Flags().Int64VarP(&chunkSize, "chunk-size", "s", 0, "specify chunk size in bytes")
	uploadCmd.Flags().IntVarP(&chunks, "chunks", "c", 3, "specify number of chunks to process")
	uploadCmd.Flags().IntVarP(&copies, "copies", "n", 1, "specify number of copies per chunk")
	uploadCmd.Flags().StringSliceVarP(&providers, "providers", "P", []string{}, "specify list of providers to use")
	uploadCmd.Flags().StringVarP(&manifestPath, "manifest", "m", "", "specify manifest file to save or provider:<provider> to upload manifest")
	uploadCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "enable quiet output")
	uploadCmd.Flags().StringVarP(&ghostMode, "ghost", "g", "", fmt.Sprintf("embed manifest using ghost mode. (%s)", strings.Join(ghost.GhostModes(), ", ")))

	// Generic provider options - for future use
	// uploadCmd.Flags().StringSliceVarP(
	// 	&rawOptions,
	// 	"option",
	// 	"o",
	// 	[]string{},
	// 	"provider option in key=value form (repeatable)",
	// )

	//nolint:errcheck // MarkFlagRequired only errors if flag doesn't exist, which is impossible here
	uploadCmd.MarkFlagRequired("file")
	//nolint:errcheck // MarkFlagRequired only errors if flag doesn't exist, which is impossible here
	uploadCmd.MarkFlagRequired("password")
	//nolint:errcheck // MarkFlagRequired only errors if flag doesn't exist, which is impossible here
	uploadCmd.MarkFlagRequired("manifest")
	uploadCmd.MarkFlagsMutuallyExclusive("chunk-size", "chunks")

	/*
	 * Download flags
	 */
	downloadCmd.Flags().StringVarP(&manifestPath, "manifest", "m", "", "specify manifest file to read or provider<provider>:<hash> to download from provider")
	downloadCmd.Flags().StringVarP(&password, "password", "p", "", "specify password")
	downloadCmd.Flags().StringVarP(&outputFile, "file", "f", "", "specify output file path")
	downloadCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "enable quiet output")
	downloadCmd.Flags().StringVarP(&ghostMode, "ghost", "g", "", fmt.Sprintf("decode manifest from ghost mode. (%s)", strings.Join(ghost.GhostModes(), ", ")))

	// Generic provider options - for future use
	// downloadCmd.Flags().StringSliceVarP(
	// 	&rawOptions,
	// 	"option",
	// 	"o",
	// 	[]string{},
	// 	"provider option in key=value form (repeatable)",
	// )

	//nolint:errcheck // MarkFlagRequired only errors if flag doesn't exist, which is impossible here
	downloadCmd.MarkFlagRequired("manifest")
	//nolint:errcheck // MarkFlagRequired only errors if flag doesn't exist, which is impossible here
	downloadCmd.MarkFlagRequired("password")
	//nolint:errcheck // MarkFlagRequired only errors if flag doesn't exist, which is impossible here
	downloadCmd.MarkFlagRequired("file")

	infoCmd.Flags().StringVarP(&manifestPath, "manifest", "m", "", "specify manifest file to read or provider<provider>:<hash> to download from provider")
	infoCmd.Flags().StringVarP(&password, "password", "p", "", "specify password")
	infoCmd.Flags().StringVarP(&ghostMode, "ghost", "g", "", fmt.Sprintf("decode manifest from ghost mode. (%s)", strings.Join(ghost.GhostModes(), ", ")))

	//nolint:errcheck // MarkFlagRequired only errors if flag doesn't exist, which is impossible here
	infoCmd.MarkFlagRequired("manifest")
	//nolint:errcheck // MarkFlagRequired only errors if flag doesn't exist, which is impossible here
	infoCmd.MarkFlagRequired("password")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(providersCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(uploadCmd)
	rootCmd.AddCommand(downloadCmd)
}
