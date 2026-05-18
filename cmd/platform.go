package cmd

import (
	"io"
	"os/exec"
	"runtime"
)

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return nil
	}
	return cmd.Start()
}

func copyClipboard(text string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		// Try wayland first, fall back to x11
		if p, _ := exec.LookPath("wl-copy"); p != "" {
			cmd = exec.Command("wl-copy", text)
		} else if p, _ := exec.LookPath("xclip"); p != "" {
			cmd = exec.Command("xclip", "-selection", "clipboard")
			cmd.Stdin = stringsReader(text)
		}
	case "darwin":
		cmd = exec.Command("pbcopy")
		cmd.Stdin = stringsReader(text)
	case "windows":
		cmd = exec.Command("cmd", "/c", "clip")
		cmd.Stdin = stringsReader(text)
	}
	if cmd != nil {
		cmd.Run()
	}
}

type stringReader struct {
	s string
	i int
}

func stringsReader(s string) *stringReader { return &stringReader{s: s} }
func (r *stringReader) Read(b []byte) (n int, err error) {
	if r.i >= len(r.s) {
		return 0, io.EOF
	}
	n = copy(b, r.s[r.i:])
	r.i += n
	return
}
