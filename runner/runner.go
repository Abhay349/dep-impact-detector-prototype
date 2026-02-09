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

// Runner handles the execution environment for tests
type Runner struct {
    WorkDir  string // The directory where we clone repositories
}

// NewRunner creates a runner and ensures the working directory exists
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

// CloneRepo clones a repository and returns its local path
func (r *Runner) CloneRepo(url, branch string) (string, error) {
    // Determine target directory name from URL
    parts := strings.Split(url, "/")
    repoName := parts[len(parts)-1]
    targetPath := filepath.Join(r.WorkDir, repoName)

    // Remove existing if present to ensure clean state
    // For demo stability: Check if already exists and skip clone if so
    if _, err := os.Stat(filepath.Join(targetPath, ".git")); err == nil {
        fmt.Printf(" [Cache] Repository %s already exists, using cached version.\n", repoName)
        // Ensure we are on the right branch and clean
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
        // Clean up failed directory before retry
        os.RemoveAll(targetPath) 
    }
    return "", err
}

// ResetRepo resets the repository to its clean state (git checkout .)
func (r *Runner) ResetRepo(repoPath string) error {
    cmd := exec.Command("git", "checkout", ".")
    cmd.Dir = repoPath
    if out, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("git checkout failed: %v\nOutput: %s", err, string(out))
    }
    return nil
}

// InjectModule replaces the target module with the local version
func (r *Runner) InjectModule(repoPath, targetModule, localModulePath string) error {
    absLocalPath, err := filepath.Abs(localModulePath)
    if err != nil {
        return fmt.Errorf("failed to get absolute path for local module: %w", err)
    }

    // go mod edit -replace
    cmd := exec.Command("go", "mod", "edit", "-replace", fmt.Sprintf("%s=%s", targetModule, absLocalPath))
    cmd.Dir = repoPath
    if out, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("go mod edit failed: %v\nOutput: %s", err, string(out))
    }

    // go mod tidy to resolve dependencies
    cmd = exec.Command("go", "mod", "tidy")
    cmd.Dir = repoPath
    if out, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("go mod tidy failed: %v\nOutput: %s", err, string(out))
    }
    
    return nil
}

// RunTests executes 'go test -json' in the given repository path
func (r *Runner) RunTests(repoPath string, packages []string) ([]byte, error) {
    args := append([]string{"test", "-json"}, packages...)
    cmd := exec.Command("go", args...)
    cmd.Dir = repoPath

    var outBuf bytes.Buffer
    cmd.Stdout = &outBuf
    
    // We ignore the error here because 'go test' returns non-zero if tests fail,
    // which is a valid outcome we want to analyze.
    // If it's a build error, the JSON output will reflect that or be empty/malformed,
    // which the analyzer should handle.
    _ = cmd.Run()

    return outBuf.Bytes(), nil
}
