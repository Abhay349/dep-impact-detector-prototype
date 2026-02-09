package runner

import (
    "bytes"
    "fmt"
    "io"
    "os/exec"
    "path/filepath"
    "strings"
)

type Executor struct {
    WorkDir  string
    Stdout   io.Writer
    Stderr   io.Writer
    Verbose  bool
}

func NewExecutor(workDir string, verbose bool) *Executor {
    return &Executor{
        WorkDir: workDir,
        Stdout:  io.Discard,
        Stderr:  io.Discard,
        Verbose: verbose,
    }
}

// CloneRepo clones a repository to targetDir
func (e *Executor) CloneRepo(repoURL, branch, targetDir string) error {
    // Check local if using git local path protocol, otherwise standard clone
    args := []string{"clone", "--depth", "1", "--branch", branch, repoURL, targetDir}
    return e.run("git", args...)
}

// ModReplace adds a replacement directive to go.mod
func (e *Executor) ModReplace(targetModule, localPath string) error {
    // Abs path needed for go mod edit
    absPath, err := filepath.Abs(localPath)
    if err != nil {
        return err
    }
    return e.run("go", "mod", "edit", "-replace", fmt.Sprintf("%s=%s", targetModule, absPath))
}

// ModTidy ensures dependencies are resolved after replace
func (e *Executor) ModTidy() error {
    return e.run("go", "mod", "tidy")
}

// RunTests executes 'go test -json' and returns the output
func (e *Executor) RunTests(packages []string) ([]byte, error) {
    args := append([]string{"test", "-json"}, packages...)
    var outBuf bytes.Buffer
    e.Stdout = &outBuf // Capture output specifically for this command
    err := e.run("go", args...)
    e.Stdout = io.Discard // Reset
    return outBuf.Bytes(), err
}

// helper to run commands
func (e *Executor) run(cmdName string, args ...string) error {
    cmd := exec.Command(cmdName, args...)
    cmd.Dir = e.WorkDir
    if e.Verbose {
        fmt.Printf("[EXEC] %s %s\n", cmdName, strings.Join(args, " "))
    }
    // We capture stdout/stderr if needed, but for now just pass through if verbose
    // or discard if silent.
    // Actually for cloning, we might want to see output if it fails.
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("command failed: %s %s: %v\nOutput: %s", cmdName, strings.Join(args, " "), err, string(output))
    }
    return nil
}
