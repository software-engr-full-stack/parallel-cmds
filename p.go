package main

// "Generic" parallel types and functions

import (
    "sync"

    "github.com/go-errors/errors"
)

type pTaskIface interface {
    process()
    end()
    getError() error
}

type pFactoryIface interface {
    getMaxThreads() (int, error)
    getTasks() []pTaskIface
}

func pRun(f pFactoryIface) ([]pTaskIface, error) {
    var wg sync.WaitGroup

    in := make(chan pTaskIface)

    wg.Add(1)
    go func() {
        for _, task := range f.getTasks() {
            in <- task
        }
        close(in)
        wg.Done()
    }()

    out := make(chan pTaskIface)

    maxThreads, err := f.getMaxThreads()
    if err != nil {
        return nil, errors.New(err)
    }

    for i := 0; i < maxThreads; i++ {
        wg.Add(1)
        go func() {
            for t := range in {
                t.process()
                out <- t
            }
            wg.Done()
        }()
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    tasks := []pTaskIface{}
    for t := range out {
        t.end()
        tasks = append(tasks, t)
    }

    return tasks, nil
}
