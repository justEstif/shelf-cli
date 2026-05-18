package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	baseURL string
	token   string
)

func Execute() error {
	rootCmd := &cobra.Command{
		Use:   "shelf",
		Short: "Upload and manage files on Shelf",
	}

	rootCmd.PersistentFlags().StringVar(&baseURL, "url", envDefault("SHELF_URL", "https://shelf.estifanos.cc"), "Shelf server URL")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "API token (or SHELF_API_TOKEN env)")
	if token == "" {
		token = os.Getenv("SHELF_API_TOKEN")
	}

	rootCmd.AddCommand(uploadCmd())
	rootCmd.AddCommand(rmCmd())
	rootCmd.AddCommand(lsCmd())
	rootCmd.AddCommand(openCmd())

	return rootCmd.Execute()
}

func envDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func requireToken() error {
	if token == "" {
		return fmt.Errorf("set SHELF_API_TOKEN or pass --token")
	}
	return nil
}

func apiURL(path string) string {
	return strings.TrimRight(baseURL, "/") + "/admin/api" + path
}

// --- upload ---

func uploadCmd() *cobra.Command {
	var folder string
	var visibility string

	cmd := &cobra.Command{
		Use:     "upload <file...>",
		Short:   "Upload files",
		Aliases: []string{"up"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(); err != nil {
				return err
			}

			var body bytes.Buffer
			w := multipart.NewWriter(&body)

			for _, path := range args {
				f, err := os.Open(path)
				if err != nil {
					return fmt.Errorf("open %s: %w", path, err)
				}
				part, err := w.CreateFormFile("files", filepath.Base(path))
				if err != nil {
					f.Close()
					return err
				}
				if _, err := io.Copy(part, f); err != nil {
					f.Close()
					return fmt.Errorf("read %s: %w", path, err)
				}
				f.Close()
			}

			if folder != "" {
				w.WriteField("folder", folder)
			}

			w.Close()

			req, _ := http.NewRequest("POST", apiURL("/upload"), &body)
			req.Header.Set("Content-Type", w.FormDataContentType())
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("upload failed: %w", err)
			}
			defer resp.Body.Close()

			var result struct {
				Uploaded []struct {
					Path string `json:"path"`
					URL  string `json:"url"`
				} `json:"uploaded"`
				Errors []string `json:"errors"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("parse response: %w", err)
			}

			for _, e := range result.Errors {
				fmt.Fprintf(os.Stderr, "\x1b[31m✗ %s\x1b[0m\n", e)
			}

			for _, f := range result.Uploaded {
				url := strings.TrimRight(baseURL, "/") + "/" + f.Path
				fmt.Printf("\x1b[32m✓\x1b[0m %s\n", url)
			}

			// Set visibility if requested
			if visibility != "" && len(result.Uploaded) > 0 {
				first := result.Uploaded[0].Path
				if err := setVisibility(first, visibility); err != nil {
					fmt.Fprintf(os.Stderr, "\x1b[31m✗ %v\x1b[0m\n", err)
				} else {
					color := map[string]string{
						"public":    "\x1b[32m",
						"private":   "\x1b[35m",
						"protected": "\x1b[33m",
					}[visibility]
					fmt.Printf("%s◆ %s\x1b[0m\n", color, visibility)
				}
			}

			// Copy first URL to clipboard
			if len(result.Uploaded) > 0 {
				firstURL := strings.TrimRight(baseURL, "/") + "/" + result.Uploaded[0].Path
				copyClipboard(firstURL)
				fmt.Printf("\x1b[2m  copied to clipboard\x1b[0m\n")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&folder, "folder", "f", "", "Upload into subfolder")
	cmd.Flags().StringVarP(&visibility, "visibility", "v", "", "Set visibility (public, private, protected)")

	return cmd
}

// --- rm ---

func rmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <path...>",
		Short: "Delete files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(); err != nil {
				return err
			}

			for _, path := range args {
				req, _ := http.NewRequest("DELETE", apiURL("/files/"+path), nil)
				req.Header.Set("Authorization", "Bearer "+token)

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					fmt.Fprintf(os.Stderr, "\x1b[31m✗ %s: %v\x1b[0m\n", path, err)
					continue
				}
				resp.Body.Close()

				if resp.StatusCode == 200 {
					fmt.Printf("\x1b[32m✓\x1b[0m deleted %s\n", path)
				} else {
					fmt.Fprintf(os.Stderr, "\x1b[31m✗ %s: HTTP %d\x1b[0m\n", path, resp.StatusCode)
				}
			}
			return nil
		},
	}
}

// --- ls ---

func lsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List files",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(); err != nil {
				return err
			}

			req, _ := http.NewRequest("GET", apiURL("/files"), nil)
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("list failed: %w", err)
			}
			defer resp.Body.Close()

			var result struct {
				Files []struct {
					Path  string `json:"path"`
					Size  int64  `json:"size"`
					IsDir bool   `json:"IsDir"`
				} `json:"files"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("parse response: %w", err)
			}

			for _, f := range result.Files {
				icon := "📄"
				if f.IsDir {
					icon = "📁"
				}
				fmt.Printf("%s %s\n", icon, f.Path)
			}
			return nil
		},
	}
}

// --- open ---

func openCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open [path]",
		Short: "Open file in browser",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ""
			if len(args) > 0 {
				path = args[0]
			}

			url := strings.TrimRight(baseURL, "/") + "/" + path
			return openBrowser(url)
		},
	}
}

// --- helpers ---

func setVisibility(path, visibility string) error {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	w.WriteField("path", path)
	w.WriteField("visibility", visibility)
	w.Close()

	req, _ := http.NewRequest("POST", apiURL("/visibility"), &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}
