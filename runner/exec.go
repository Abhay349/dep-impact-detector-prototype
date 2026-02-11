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

func (e *Executor) CloneRepo(repoURL, branch, targetDir string) error {
    args := []string{"clone", "--depth", "1", "--branch", branch, repoURL, targetDir}
    return e.run("git", args...)
}

func (e *Executor) ModReplace(targetModule, localPath string) error {
    absPath, err := filepath.Abs(localPath)
    if err != nil {
        return err
    }
    return e.run("go", "mod", "edit", "-replace", fmt.Sprintf("%s=%s", targetModule, absPath))
}

func (e *Executor) ModTidy() error {
    return e.run("go", "mod", "tidy")
}
func (e *Executor) RunTests(packages []string) ([]byte, error) {
    args := append([]string{"test", "-json"}, packages...)
    var outBuf bytes.Buffer
    e.Stdout = &outBuf 
    err := e.run("go", args...)
    e.Stdout = io.Discard 
    return outBuf.Bytes(), err
}

// helper to run commands
func (e *Executor) run(cmdName string, args ...string) error {
    cmd := exec.Command(cmdName, args...)
    cmd.Dir = e.WorkDir
    if e.Verbose {
        fmt.Printf("[EXEC] %s %s\n", cmdName, strings.Join(args, " "))
    }
   
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("command failed: %s %s: %v\nOutput: %s", cmdName, strings.Join(args, " "), err, string(output))
    }
    return nil
}
