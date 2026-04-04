package gitutil

import (
	"fmt"
	"os/exec"
	"strings"
)

type Repo struct {
	Dir string
}

func New(dir string) *Repo {
	return &Repo{Dir: dir}
}

func (r *Repo) run(args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", r.Dir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (r *Repo) CurrentTag() string {
	out, err := r.run("describe", "--tags", "--exact-match")
	if err != nil {
		return "none"
	}
	return out
}

func (r *Repo) CurrentSHA() string {
	out, err := r.run("rev-parse", "--short", "HEAD")
	if err != nil {
		return "unknown"
	}
	return out
}

func (r *Repo) CurrentBranch() string {
	out, err := r.run("branch", "--show-current")
	if err != nil {
		return "unknown"
	}
	return out
}

func (r *Repo) VersionExists(version string) bool {
	_, err := r.run("rev-parse", version)
	return err == nil
}

func (r *Repo) Checkout(version string) error {
	_, err := r.run("checkout", version, "--quiet")
	return err
}

func (r *Repo) DiffStat(fromSHA string, paths ...string) (string, error) {
	args := []string{"diff", "--stat", fromSHA + "..HEAD", "--"}
	args = append(args, paths...)
	return r.run(args...)
}

func (r *Repo) Tags() (string, error) {
	return r.run("tag", "--sort=-v:refname")
}

func (r *Repo) TagDate(tag string) string {
	out, err := r.run("log", "-1", "--format=%ai", tag)
	if err != nil {
		return ""
	}
	parts := strings.Fields(out)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func (r *Repo) TagMessage(tag string) string {
	out, err := r.run("tag", "-l", "--format=%(subject)", tag)
	if err != nil {
		return ""
	}
	return out
}

func (r *Repo) CatFileType(ref, path string) (string, error) {
	return r.run("cat-file", "-t", ref+":"+path)
}

func (r *Repo) Archive(ref, path, destDir string) error {
	archive := exec.Command("git", "-C", r.Dir, "archive", ref, "--", path)
	tar := exec.Command("tar", "-x", "-C", destDir)

	pipe, err := archive.StdoutPipe()
	if err != nil {
		return fmt.Errorf("pipe: %w", err)
	}
	tar.Stdin = pipe

	if err := archive.Start(); err != nil {
		return fmt.Errorf("git archive: %w", err)
	}
	if err := tar.Start(); err != nil {
		return fmt.Errorf("tar: %w", err)
	}
	if err := archive.Wait(); err != nil {
		return fmt.Errorf("git archive wait: %w", err)
	}
	return tar.Wait()
}

func (r *Repo) ShowFile(ref, path string) (string, error) {
	return r.run("show", ref+":"+path)
}
