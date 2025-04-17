package cmd

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/peterbourgon/ff/v2/ffcli"
	"github.com/runetale/runetale-oidc-server/database"
	"github.com/runetale/runetale-oidc-server/utility"
)

var pretreatmnetArgs struct {
	dbUrl    string
	dbName   string
	logLevel string
	logPath  string
	logFmt   string
	debug    bool
}

var pretreatmentCmd = &ffcli.Command{
	Name:       "pretreatment",
	ShortUsage: "pretreatment [args]",
	FlagSet: (func() *flag.FlagSet {
		fs := flag.NewFlagSet("pretreatment", flag.ExitOnError)
		fs.StringVar(&pretreatmnetArgs.dbUrl, "dburl", "", "set postgresql db url")
		fs.StringVar(&pretreatmnetArgs.dbName, "dbname", "runetale-oidc-server", "set postgresql db name")
		fs.StringVar(&pretreatmnetArgs.logLevel, "loglevel", utility.DebugLevelStr, "set log level")
		fs.StringVar(&pretreatmnetArgs.logPath, "logpath", utility.StdErrFilePath, "set log output path")
		fs.StringVar(&pretreatmnetArgs.logFmt, "logfmt", "json", "set log format")
		fs.BoolVar(&pretreatmnetArgs.debug, "debug", false, "is debug")
		return fs
	})(),
	Exec: execPretreatment,
}

func execPretreatment(ctx context.Context, args []string) error {
	f, err := os.Create(pretreatmnetArgs.logPath)
	if err != nil {
		return err
	}

	logger, err := utility.NewLogger(f, pretreatmnetArgs.logFmt, pretreatmnetArgs.logLevel)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	db, err := database.NewPostgres(pretreatmnetArgs.dbUrl)
	if err != nil {
		return err
	}

	err = db.CreateDB(pretreatmnetArgs.dbName)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	logger.Info("finished pretreatment")

	return nil
}
