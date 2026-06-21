package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCLIVersionSubprocess(t *testing.T) {
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(t.TempDir(), "wpci")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/wpci")
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building wpci: %v\n%s", err, out)
	}

	run := exec.Command(bin, "version", "--json")
	run.Dir = repoRoot
	out, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("running wpci version: %v\n%s", err, out)
	}
	if len(out) == 0 {
		t.Fatal("expected version output")
	}
	_ = os.Remove(bin)
}

func TestCLIAccountSubprocess(t *testing.T) {
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(t.TempDir(), "wpci")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/wpci")
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building wpci: %v\n%s", err, out)
	}

	configDir := t.TempDir()
	run := exec.Command(bin, "account", "add", "test", "--server", "https://ci.example.com", "--token", "tok")
	run.Dir = repoRoot
	run.Env = append(os.Environ(), "XDG_CONFIG_HOME="+configDir)
	out, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("running wpci account add: %v\n%s", err, out)
	}
	if len(out) == 0 {
		t.Fatal("expected account add output")
	}

	run2 := exec.Command(bin, "account", "token", "status", "test")
	run2.Dir = repoRoot
	run2.Env = append(os.Environ(), "XDG_CONFIG_HOME="+configDir)
	out2, err := run2.CombinedOutput()
	if err != nil {
		t.Fatalf("running wpci account token status: %v\n%s", err, out2)
	}
	if len(out2) == 0 {
		t.Fatal("expected token status output")
	}
	_ = os.Remove(bin)
}
