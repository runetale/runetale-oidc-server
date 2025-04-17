package cmd

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/peterbourgon/ff/v2/ffcli"
	"github.com/runetale/runetale-oidc-server/database"
	"github.com/runetale/runetale-oidc-server/utility"
)

var upArgs struct {
	dbUrl    string
	logLevel string
	logPath  string
	logFmt   string
	debug    bool
}

var upCmd = &ffcli.Command{
	Name:       "up",
	ShortUsage: "up [flags]",
	FlagSet: (func() *flag.FlagSet {
		fs := flag.NewFlagSet("up", flag.ExitOnError)
		fs.StringVar(&upArgs.dbUrl, "dburl", "", "set postgresql db url")
		fs.StringVar(&upArgs.logLevel, "loglevel", utility.DebugLevelStr, "set log level")
		fs.StringVar(&upArgs.logPath, "logpath", utility.StdErrFilePath, "set log output path")
		fs.StringVar(&upArgs.logFmt, "logfmt", "json", "set log format")
		fs.BoolVar(&upArgs.debug, "debug", false, "is debug")
		return fs
	})(),
	Exec: execUp,
}

func execUp(ctx context.Context, args []string) error {
	f, err := os.Create(upArgs.logPath)
	if err != nil {
		return err
	}

	logger, err := utility.NewLogger(f, upArgs.logFmt, upArgs.logLevel)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	logger.Info(upArgs.dbUrl)

	db, err := database.NewPostgres(upArgs.dbUrl)
	if err != nil {
		log.Fatalf("failed to open postgresql: %v", err)
		return nil
	}

	err = db.MigrateUp("migrations")
	if err != nil {
		logger.Error(fmt.Sprintf("migrate failed"), err)
		return err
	}
	logger.Info("migration succeded")

	logger.Info("finished migratation up")

	return nil
}
