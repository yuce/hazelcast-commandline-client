package sql

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/shell"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type SQLShellCommand struct {
	client  *hazelcast.Client
	verbose bool
}

func (cm *SQLShellCommand) Augment(ec plug.ExecContext, props *plug.Properties) error {
	if ec.CommandName() == "clc sql shell" {
		props.Set(clc.PropertyFormat, "table")
	}
	return nil
}

func (cm *SQLShellCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("shell")
	help := "Start the interactive SQL shell"
	cc.SetCommandHelp(help, help)
	return nil
}

func (cm *SQLShellCommand) Exec(ec plug.ExecContext) error {
	ctx := context.Background()
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	cm.client = ci.Client()
	cm.verbose = ec.Props().GetBool(clc.PropertyVerbose)
	return nil
}

func (cm *SQLShellCommand) ExecInteractive(ec plug.ExecInteractiveContext) error {
	endLineFn := func(line string) bool {
		line = strings.TrimSpace(line)
		return strings.HasPrefix(line, "help") || strings.HasPrefix(line, "\\") || strings.HasSuffix(line, ";")
	}
	textFn := func(ctx context.Context, text string) error {
		text, err := convertStatement(text)
		if err != nil {
			return err
		}
		res, err := cm.client.SQL().Execute(ctx, text)
		if err != nil {
			return adaptSQLError(err)
		}
		if err := updateOutput(ec, res, cm.verbose); err != nil {
			return err
		}
		if err := ec.FlushOutput(); err != nil {
			return err
		}
		return nil
	}
	path := paths.Join(paths.Home(), "sql.history")
	sh := shell.New("SQL> ", " ... ", path, ec.Stdout(), ec.Stderr(), endLineFn, textFn)
	sh.SetCommentPrefix("--")
	defer sh.Close()
	return sh.Start(context.Background())
}

func convertStatement(stmt string) (string, error) {
	stmt = strings.TrimSpace(stmt)
	if strings.HasPrefix(stmt, "help") {
		return "", errors.New(help())
	}
	if strings.HasPrefix(stmt, "\\") {
		// this is a shell command
		parts := strings.Fields(stmt)
		switch parts[0] {
		case "\\dm":
			if len(parts) == 1 {
				return "show mappings;", nil
			}
			if len(parts) == 2 {
				// escape single quote
				mn := strings.Replace(parts[1], "'", "''", -1)
				return fmt.Sprintf(`
					SELECT * FROM information_schema.mappings
					WHERE table_name = '%s';
				`, mn), nil
			}
			return "", fmt.Errorf("Usage: \\dm [mapping]")
		case "\\dm+":
			if len(parts) == 1 {
				return "show mappings;", nil
			}
			if len(parts) == 2 {
				// escape single quote
				mn := strings.Replace(parts[1], "'", "''", -1)
				return fmt.Sprintf(`
					SELECT * FROM information_schema.columns
					WHERE table_name = '%s';
				`, mn), nil
			}
			return "", fmt.Errorf("Usage: \\dm+ [mapping]")
		}
		return "", fmt.Errorf("Unknown shell command: %s", stmt)
	}
	return stmt, nil
}

func help() string {
	return `
Commands:
	\dm           list mappings
	\dm  MAPPING  display info about a mapping
	\dm+ MAPPING  describe a mapping
`
}

func init() {
	plug.Registry.RegisterAugmentor("20-sql-shell", &SQLShellCommand{})
	Must(plug.Registry.RegisterCommand("sql:shell", &SQLShellCommand{}))
}
