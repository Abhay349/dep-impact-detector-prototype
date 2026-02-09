package main

import (
    "bytes"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "github.com/abhaypandey/dep-impact-detector-prototype/analyzer"
    "github.com/abhaypandey/dep-impact-detector-prototype/runner"
)

type Config struct {
    TargetModule string     `json:"target_module"`
    Consumers    []Consumer `json:"consumers"`
}

type Consumer struct {
    Name      string   `json:"name"`
    RepoURL   string   `json:"repo_url"`
    Branch    string   `json:"branch"`
    ModuleDir string   `json:"module_dir"`
    Packages  []string `json:"packages"`
}

func main() {
    localModulePath := flag.String("local", "", "Path to the local module version to test")
    configFile := flag.String("config", "config.json", "Path to configuration file")
    workDir := flag.String("workdir", "./temp_work", "Directory for temporary clones")
    flag.Parse()

    if *localModulePath == "" {
        flag.Usage()
        log.Fatal("Error: -local flag is required")
    }

    // Load Config
    cfgData, err := os.ReadFile(*configFile)
    if err != nil {
        log.Fatalf("Failed to read config file: %v", err)
    }
    var config Config
    if err := json.Unmarshal(cfgData, &config); err != nil {
        log.Fatalf("Failed to parse config: %v", err)
    }

    // Initialize Runner
    r, err := runner.NewRunner(*workDir)
    if err != nil {
        log.Fatalf("Failed to initialize runner: %v", err)
    }
    // defer os.RemoveAll(*workDir) // Clean up disabled for demo/caching stability

    fmt.Printf("Analyzing impact of %s (local: %s) on %d consumers...\n", config.TargetModule, *localModulePath, len(config.Consumers))

    for _, consumer := range config.Consumers {
        fmt.Printf("\n--- Processing Consumer: %s ---\n", consumer.Name)
        
        // 1. Clone
        fmt.Printf("Cloning %s (%s)...\n", consumer.RepoURL, consumer.Branch)
        repoPath, err := r.CloneRepo(consumer.RepoURL, consumer.Branch)
        if err != nil {
            log.Printf("Skipping %s: Clone failed: %v", consumer.Name, err)
            continue
        }

        // Determine module path (root of repo or subdir)
        modulePath := repoPath
        if consumer.ModuleDir != "" {
            modulePath = filepath.Join(repoPath, consumer.ModuleDir)
        }

        // 2. Baseline Run (Current Published Version)
        fmt.Println("Running Baseline Tests...")
        baseOut, err := r.RunTests(modulePath, consumer.Packages)
        if err != nil {
             // Tests might fail normally, continue analysis but warn
             fmt.Printf("Warning: Baseline execution had errors (exit code), check results.\n")
        }
        
        baseResults, err := analyzer.ParseTestOutput(bytes.NewReader(baseOut))
        if err != nil {
            log.Printf("Failed to parse baseline results: %v", err)
        }
        fmt.Printf("Baseline: Found results for %d packages\n", len(baseResults))

        // 3. Reset Repo for clean state
        if err := r.ResetRepo(repoPath); err != nil {
            log.Printf("Failed to reset repo: %v", err)
            continue
        }

        // 4. Inject Local Module
        fmt.Println("Injecting Local Module Version...")
        if err := r.InjectModule(modulePath, config.TargetModule, *localModulePath); err != nil {
            log.Printf("Failed to inject module: %v", err)
            continue
        }

        // 5. Experiment Run (Local Version)
        fmt.Println("Running Experiment Tests...")
        expOut, err := r.RunTests(modulePath, consumer.Packages)
        if err != nil {
             fmt.Printf("Warning: Experiment execution had errors.\n")
        }

        expResults, err := analyzer.ParseTestOutput(bytes.NewReader(expOut))
        if err != nil {
            log.Printf("Failed to parse experiment results: %v", err)
        }
        fmt.Printf("Experiment: Found results for %d packages\n", len(expResults))

        // 6. Compare
        fmt.Println("\n=== Impact Report ===")
        diff := analyzer.CompareResults(baseResults, expResults)
        fmt.Println(diff)
        fmt.Println("=====================")
    }
}
