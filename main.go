package main

import (
    "os"
    "strings"

    "github.com/go-errors/errors"

    yaml "gopkg.in/yaml.v3"
)

func main() {
    args := os.Args[1:]

    cmdsFile := "./cmds.yml"
    if len(args) == 1 {
        cmdsFile = args[0]
    }

    data, err := os.ReadFile(cmdsFile)
    if err != nil {
        panic(err.(*errors.Error).ErrorStack())
    }

    commands := Commands{}

    err = yaml.Unmarshal(data, &commands)
    if err != nil {
        panic(err.(*errors.Error).ErrorStack())
    }

    for _, si := range commands.Series {
        if err = si.run(); err != nil {
            panic(err.(*errors.Error).ErrorStack())
        }
    }

    err = runParallel(commands.Parallel)
    if err != nil {
        panic(err.(*errors.Error).ErrorStack())
    }
}

type Commands struct {
    Meta     Meta          `yaml:"meta"`
    Series   []*SeriesItem `yaml:"series"`
    Parallel Parallel      `yaml:"parallel"`
}

type Meta struct {
    WorkingDir string `yaml:"working_dir"`
    OutFile    string `yaml:"out_file"`
    ErrFile    string `yaml:"err_file"`
}

type Command struct {
    Meta       Meta   `yaml:"meta"`
    CmdStr     string `yaml:"cmd"`
    UpdateMeta bool   `yaml:"update_meta"`
}

type SeriesItem struct {
    Meta     Meta     `yaml:"meta"`
    CmdStr   string   `yaml:"cmd"`
    Parallel Parallel `yaml:"parallel"`
}

type Parallel []*Command

func (m *Meta) update(topMeta *Meta) {
    if m.WorkingDir == "" {
        m.WorkingDir = topMeta.WorkingDir
    }

    if m.OutFile == "" {
        m.OutFile = topMeta.OutFile
    }

    if m.ErrFile == "" {
        m.ErrFile = topMeta.ErrFile
    }
}

func (si *SeriesItem) run() error {
    if si.CmdStr != "" {
        runner, err := runnerNew(si.CmdStr, &si.Meta)
        if err != nil {
            return errors.New(err)
        }
        return runner.run()
    }

    return runParallel(si.Parallel)
}

func runParallel(pList Parallel) error {
    runners := []*runnerType{}
    for _, pi := range pList {
        runner, err := runnerNew(pi.CmdStr, &pi.Meta)
        if err != nil {
            return errors.New(err)
        }
        runners = append(runners, runner)
    }

    maxThreads := 1000

    given := strings.TrimSpace(os.Getenv("DEBUG"))
    isDebugSerial := false
    if given == "serial" {
        isDebugSerial = true
    }

    if isDebugSerial {
        // Serial, for debugging
        for _, runner := range runners {
            if err := runner.run(); err != nil {
                return errors.New(err)
            }
        }

        return nil
    }

    // Parallel
    pcft, err := pCmdsFactoryNew(maxThreads, runners)
    if err != nil {
        return errors.New(err)
    }
    tasks, err := pRun(pcft)
    if err != nil {
        return errors.New(err)
    }

    errList := []error{}
    for _, task := range tasks {
        err = task.getError()
        if err != nil {
            errList = append(errList, err)
        }
    }
    if len(errList) > 0 {
        // Lint disabled because editor is running an older version of Go where errors.Join
        // is not available.
        err = errors.Join(errList...)
        return errors.New(err)
    }

    return nil
}
