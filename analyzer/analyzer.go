package analyzer

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io"
    "time"
)

type TestEvent struct {
    Time    time.Time `json:"Time"`
    Action  string    `json:"Action"`
    Package string    `json:"Package"`
    Test    string    `json:"Test"`
    Output  string    `json:"Output"`
    Elapsed float64   `json:"Elapsed"`
}

type TestResult struct {
    Name    string
    Status  string // pass, fail, skip
    Elapsed float64
}

type ResultMap map[string]map[string]TestResult

func ParseTestOutput(r io.Reader) (ResultMap, error) {
    scanner := bufio.NewScanner(r)
    results := make(ResultMap)

    for scanner.Scan() {
        var event TestEvent
        if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
            continue
        }

        if event.Package == "" {
            continue
        }

        if _, ok := results[event.Package]; !ok {
            results[event.Package] = make(map[string]TestResult)
        }

        if event.Test != "" {
            if event.Action == "pass" || event.Action == "fail" || event.Action == "skip" {
                results[event.Package][event.Test] = TestResult{
                    Name:    event.Test,
                    Status:  event.Action,
                    Elapsed: event.Elapsed,
                }
            }
        } else if event.Test == "" {
            if event.Action == "fail" || event.Action == "pass" {
            }
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }
    return results, nil
}


func CompareResults(baseline, experiment ResultMap) string {
    var diff string
    
    regressions := 0
    for pkg, tests := range baseline {
        expTests, ok := experiment[pkg]
        if !ok {
            diff += fmt.Sprintf("Missing Package in Experiment: %s\n", pkg)
            continue
        }
        for name, res := range tests {
            expRes, ok := expTests[name]
            if !ok {
                diff += fmt.Sprintf("[%s] Test Missing in Experiment: %s\n", pkg, name)
                continue
            }
            if res.Status == "pass" && expRes.Status == "fail" {
                diff += fmt.Sprintf("REGRESSION [%s] %s: Pass -> Fail\n", pkg, name)
                regressions++
            }
        }
    }

    fixes := 0
    for pkg, tests := range baseline {
        expTests, ok := experiment[pkg]
        if !ok { continue }
        for name, res := range tests {
            expRes, ok := expTests[name]
            if !ok { continue }
            if res.Status == "fail" && expRes.Status == "pass" {
                diff += fmt.Sprintf("FIX [%s] %s: Fail -> Pass\n", pkg, name)
                fixes++
            }
        }
    }

    summary := fmt.Sprintf("\nSummary: %d Regressions, %d Fixes\n", regressions, fixes)
    return diff + summary
}
