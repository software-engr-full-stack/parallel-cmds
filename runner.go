package main

// Shell command runner functions

import (
    "fmt"
    "io/fs"
    "os"
    "os/exec"
    "path/filepath"

    "github.com/go-errors/errors"
    "golang.org/x/sys/unix"
)

type customOut struct {
    File string
}

var filePerm = 0644

func (c customOut) Write(p []byte) (int, error) {
    fmt.Print(string(p))

    err := writeToCustomFile(c.File, p)
    if err != nil {
        return -1, errors.New(err)
    }

    return len(p), nil
}

type customErr struct {
    File string
}

func (c customErr) Write(p []byte) (int, error) {
    fmt.Fprint(os.Stderr, string(p))

    err := writeToCustomFile(c.File, p)
    if err != nil {
        return -1, errors.New(err)
    }

    return len(p), nil
}

func writeToCustomFile(file string, data []byte) error {
    if file == "" {
        return nil
    }

    f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fs.FileMode(filePerm))
    if err != nil {
        return errors.New(err)
    }
    defer f.Close()

    if _, err := f.WriteString(string(data)); err != nil {
        return errors.New(err)
    }

    return nil
}

type runnerType struct {
    Cmd *exec.Cmd
}

func (rt *runnerType) run() error {
    fmt.Println("# ... START", rt.Cmd.Args, rt.Cmd.Dir)
    if rt.Cmd.Dir == "" {
        if err := rt.Cmd.Run(); err != nil {
            return errors.New(err)
        }
        return nil
    }

    cwd, err := os.Getwd()
    if err != nil {
        return errors.New(err)
    }
    err = os.Chdir(rt.Cmd.Dir)
    if err != nil {
        return errors.New(err)
    }
    if err = rt.Cmd.Run(); err != nil {
        return errors.New(err)
    }
    err = os.Chdir(cwd)
    if err != nil {
        return errors.New(err)
    }
    fmt.Println("# ... END", rt.Cmd.Args, rt.Cmd.Dir)

    return nil
}

func runnerNew(cmdStr string, meta *Meta) (*runnerType, error) {
    // Apparently, this form is safer over breaking the command string into fields.
    // See "gosec G204".
    cmd := exec.Command("/bin/bash", "-c", cmdStr)

    if meta.WorkingDir != "" {
        cmd.Dir = meta.WorkingDir
    }

    if meta.OutFile != "" {
        err := prepareCustomFile(meta.WorkingDir, meta.OutFile)
        if err != nil {
            return nil, errors.New(err)
        }
        cmd.Stdout = customOut{File: meta.OutFile}
    } else {
        cmd.Stdout = os.Stdout
    }

    if meta.ErrFile != "" {
        err := prepareCustomFile(meta.WorkingDir, meta.ErrFile)
        if err != nil {
            return nil, errors.New(err)
        }
        cmd.Stderr = customErr{File: meta.ErrFile}
    } else {
        cmd.Stderr = os.Stderr
    }

    return &runnerType{Cmd: cmd}, nil
}

func prepareCustomFile(workingDir, file string) error {
    if err := unix.Access(workingDir, unix.W_OK); err != nil {
        return errors.New(fmt.Errorf(
            "write access check for %#v failed: %#v", workingDir, err.Error(),
        ))
    }

    fullPath := filepath.Join(workingDir, file)
    if _, err := os.Stat(fullPath); err == nil {
        if err := os.Truncate(fullPath, 0); err != nil {
            return errors.New(err)
        }
    } else if errors.Is(err, os.ErrNotExist) {
    } else {
        // Schrodinger: file may or may not exist.
        panic(fmt.Errorf("weird error in checking existence of file %#v: %#v", fullPath, err.Error()))
    }

    return nil
}
