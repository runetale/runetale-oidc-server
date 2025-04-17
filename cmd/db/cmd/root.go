package cmd

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v2/ffcli"
)

func Run(args []string) error {
	fs := flag.NewFlagSet("db", flag.ExitOnError)
	cmd := &ffcli.Command{
		Name:       "db",
		ShortUsage: "db <subcommands> [command flags]",
		ShortHelp:  "db command is a command to be executed against the runetale-server db",
		Subcommands: []*ffcli.Command{
			pretreatmentCmd,
			upCmd,
		},
		FlagSet: fs,
		Exec:    func(context.Context, []string) error { return flag.ErrHelp },
	}

	if err := cmd.Parse(args); err != nil {
		return err
	}

	if err := cmd.Run(context.Background()); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	return nil
}
