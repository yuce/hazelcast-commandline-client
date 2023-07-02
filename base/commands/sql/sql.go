//go:build base || sql

package sql

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client/sql"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
	. "github.com/hazelcast/hazelcast-commandline-client/prv/check"
	"github.com/hazelcast/hazelcast-commandline-client/prv/plug"
)

const (
	propertyUseMappingSuggestion = "use-mapping-suggestion"
	minServerVersion             = "5.0.0"
)

type SQLCommand struct{}

func (cm *SQLCommand) Augment(ec plug.ExecContext, props *plug.Properties) error {
	// set the default format to table in the interactive mode
	if ec.CommandName() == "clc shell" && len(ec.Args()) == 0 {
		props.Set(clc.PropertyFormat, base.PrinterTable)
	}
	return nil
}

func (cm *SQLCommand) Init(cc plug.InitContext) error {
	if cc.Interactive() {
		return errors.ErrNotAvailable
	}
	cc.SetCommandUsage("sql [query] [flags]")
	cc.SetPositionalArgCount(1, 1)
	cc.AddCommandGroup("sql", "SQL")
	cc.SetCommandGroup("sql")
	long := fmt.Sprintf(`Runs the given SQL query or starts the SQL shell

If QUERY is not given, then the SQL shell is started.
	
This command requires a Viridian or a Hazelcast cluster
having version %s or better.
`, minServerVersion)
	cc.SetCommandHelp(long, "Run SQL")
	cc.AddBoolFlag(propertyUseMappingSuggestion, "", false, false, "execute the proposed CREATE MAPPING suggestion and retry the query")
	return nil
}

func (cm *SQLCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	// this method is only for the non-interactive mode
	if len(ec.Args()) < 1 {
		return nil
	}
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	if sv, ok := cmd.CheckServerCompatible(ci, minServerVersion); !ok {
		return fmt.Errorf("server (%s) does not support this command, at least %s is expected", sv, minServerVersion)
	}
	query := ec.Args()[0]
	res, stop, err := cm.execQuery(ctx, query, ec)
	if err != nil {
		return err
	}
	// this should be deferred because UpdateOutput will iterate on the result
	defer stop()
	verbose := ec.Props().GetBool(clc.PropertyVerbose)
	return UpdateOutput(ctx, ec, res, verbose)
}

func (cm *SQLCommand) execQuery(ctx context.Context, query string, ec plug.ExecContext) (sql.Result, context.CancelFunc, error) {
	return ExecSQL(ctx, ec, query)
}

func init() {
	plug.Registry.RegisterAugmentor("20-sql", &SQLCommand{})
	Must(plug.Registry.RegisterCommand("sql", &SQLCommand{}))
}
