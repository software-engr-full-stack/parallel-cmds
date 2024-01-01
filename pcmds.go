package main

// Parallel commands types and functions for use by "generic" parallel functions
// Coupling between the app to the generic parallel runner.

import (
    "fmt"
    "os"

    "github.com/go-errors/errors"
)

type pCmdsFactoryType struct {
    MaxThreads int
    Runners    []*runnerType
}

func (pcft *pCmdsFactoryType) getMaxThreads() (int, error) {
    if pcft.MaxThreads < 1 {
        return -1, errors.New("max threads must be greater than zero")
    }
    return pcft.MaxThreads, nil
}

func (pcft *pCmdsFactoryType) getTasks() []pTaskIface {
    tasks := []pTaskIface{}
    for _, runner := range pcft.Runners {
        var task pTaskIface = &pCmdsTask{
            Runner: runner,
        }
        tasks = append(tasks, task)
    }
    return tasks
}

func pCmdsFactoryNew(maxThreads int, runners []*runnerType) (*pCmdsFactoryType, error) {
    if maxThreads < 1 {
        return nil, errors.New("max threads must be greater than zero")
    }

    return &pCmdsFactoryType{MaxThreads: maxThreads, Runners: runners}, nil
}

type pCmdsTask struct {
    Runner           *runnerType
    ErrorWithContext error
}

func (pct *pCmdsTask) process() {
    cmdWords := pct.Runner.Cmd.Args
    fmt.Printf("# ... START: %#v %#v\n", cmdWords, pct.Runner.Cmd.Dir)

    errFmtStr := "ERROR: %v failed => %#v"

    if pct.Runner.Cmd.Dir == "" {
        if err := pct.Runner.Cmd.Run(); err != nil {
            pct.ErrorWithContext = errors.New(
                fmt.Errorf(errFmtStr, cmdWords, err.Error()),
            )
        }
        return
    }

    cwd, err := os.Getwd()
    if err != nil {
        pct.ErrorWithContext = errors.New(
            fmt.Errorf(errFmtStr, cmdWords, err.Error()),
        )
    }

    err = os.Chdir(pct.Runner.Cmd.Dir)
    if err != nil {
        pct.ErrorWithContext = errors.New(
            fmt.Errorf(errFmtStr, cmdWords, err.Error()),
        )
    }
    if err = pct.Runner.Cmd.Run(); err != nil {
        pct.ErrorWithContext = errors.New(
            fmt.Errorf(errFmtStr, cmdWords, err.Error()),
        )
    }
    err = os.Chdir(cwd)
    if err != nil {
        pct.ErrorWithContext = errors.New(
            fmt.Errorf(errFmtStr, cmdWords, err.Error()),
        )
    }
}

func (pct *pCmdsTask) end() {
    fmt.Printf("# ... END: %#v %#v\n", pct.Runner.Cmd.Args, pct.Runner.Cmd.Dir)
}

func (pct *pCmdsTask) getError() error {
    return pct.ErrorWithContext
}
