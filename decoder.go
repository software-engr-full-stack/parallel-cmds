package main

import (
    "fmt"

    "github.com/go-errors/errors"

    yaml "gopkg.in/yaml.v3"
)

type commandsDecoder struct {
    Meta     Meta        `yaml:"meta"`
    Series   []yaml.Node `yaml:"series"`
    Parallel Parallel    `yaml:"parallel"`
}

type commandDecoder struct {
    Meta   Meta   `yaml:"meta"`
    CmdStr string `yaml:"cmd"`
}

func (commands *Commands) UnmarshalYAML(v *yaml.Node) error {
    var commandsDec commandsDecoder
    if err := v.Decode(&commandsDec); err != nil {
        return errors.New(err)
    }

    commands.Meta = commandsDec.Meta

    for _, item := range commandsDec.Series { //nolint:gocritic,rangeValCopy
        var str string
        if err := item.Decode(&str); err == nil {
            si := SeriesItem{CmdStr: str}
            si.Meta.update(&commands.Meta)
            commands.Series = append(commands.Series, &si)
            continue
        }

        var serItem SeriesItem
        if err := item.Decode(&serItem); err == nil {
            commands.Series = append(commands.Series, &serItem)
            continue
        }

        var list []*Command
        if err := item.Decode(&list); err == nil {
            for _, pCommand := range list {
                if pCommand.UpdateMeta {
                    pCommand.Meta.update(&commands.Meta)
                }
            }
            commands.Series = append(commands.Series, &SeriesItem{Parallel: list})
            continue
        }

        return errors.New(fmt.Errorf("unsupported object: %#v", item))
    }

    commands.Parallel = commandsDec.Parallel

    return nil
}

func (command *Command) UnmarshalYAML(v *yaml.Node) error {
    var str string
    if err := v.Decode(&str); err == nil {
        command.CmdStr = str
        command.UpdateMeta = true
        return nil
    }

    var commandDec commandDecoder
    if err := v.Decode(&commandDec); err == nil {
        command.Meta = commandDec.Meta
        command.CmdStr = commandDec.CmdStr
        return nil
    }

    return errors.New(fmt.Errorf("unsupported object: %#v", v))
}
