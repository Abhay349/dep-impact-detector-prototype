package runner

import (
    "bytes"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"
)

type Runner struct {
    WorkDir  string 
}

func NewRunner(workDir string) (*Runner, error) {
    absPath, err := filepath.Abs(workDir)
    if err != nil {
        return nil, err
    }
    if err := os.MkdirAll(absPath, 0755); err != nil {
        return nil, err
    }
    return &Runner{WorkDir: absPath}, nil
}

func (r *Runner) CloneRepo(url, branch string) (string, error) {
    parts := strings.Split(url, "/")
    repoName := parts[len(parts)-1]
    targetPath := filepath.Join(r.WorkDir, repoName)

    if _, err := os.Stat(filepath.Join(targetPath, ".git")); err == nil {
        fmt.Printf(" [Cache] Repository %s already exists, using cached version.\n", repoName)
        cmd := exec.Command("git", "checkout", ".")
        cmd.Dir = targetPath
        cmd.Run()
        return targetPath, nil
    }

    var err error
    for i := 0; i < 3; i++ {
        cmd := exec.Command("git", "clone", "--depth", "1", "--branch", branch, url, targetPath)
        out, cmdErr := cmd.CombinedOutput()
        if cmdErr == nil {
            return targetPath, nil
        }
        err = fmt.Errorf("git clone failed (attempt %d/3): %v\nOutput: %s", i+1, cmdErr, string(out))
        fmt.Printf("Clone attempt %d failed, retrying... (%v)\n", i+1, cmdErr)
        time.Sleep(2 * time.Second)
        os.RemoveAll(targetPath) 
    }
    return "", err
}

func (r *Runner) ResetRepo(repoPath string) error {
    cmd := exec.Command("git", "checkout", ".")
    cmd.Dir = repoPath
    if out, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("git checkout failed: %v\nOutput: %s", err, string(out))
    }
    return nil
}

func (r *Runner) InjectModule(repoPath, targetModule, localModulePath string) error {
    absLocalPath, err := filepath.Abs(localModulePath)
    if err != nil {
        return fmt.Errorf("failed to get absolute path for local module: %w", err)
    }
    
    cmd := exec.Command("go", "mod", "edit", "-replace", fmt.Sprintf("%s=%s", targetModule, absLocalPath))
    cmd.Dir = repoPath
    if out, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("go mod edit failed: %v\nOutput: %s", err, string(out))
    }

    cmd = exec.Command("go", "mod", "tidy")
    cmd.Dir = repoPath
    if out, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("go mod tidy failed: %v\nOutput: %s", err, string(out))
    }
    
    return nil
}

func (r *Runner) RunTests(repoPath string, packages []string) ([]byte, error) {
    args := append([]string{"test", "-json"}, packages...)
    cmd := exec.Command("go", args...)
    cmd.Dir = repoPath

    var outBuf bytes.Buffer
    cmd.Stdout = &outBuf
    
    _ = cmd.Run()

    return outBuf.Bytes(), nil
}
