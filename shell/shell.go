package shell

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/canonical/go-dqlite/v3/app"
)

// Shell can be used to implement interactive prompts for inspecting a dqlite
// database.
type Shell struct {
	DB   *sql.DB
	conf *ShellConfig
}

type ShellConfig struct {
	App      *app.App
	Database string
	Timeout  time.Duration
}

// New creates a new Shell connected to the given database.
func New(config *ShellConfig) (*Shell, error) {
	db, err := config.App.Open(context.Background(), config.Database)
	if err != nil {
		return nil, err
	}

	return &Shell{
		DB:   db,
		conf: config,
	}, nil
}

// Process a single input line.
func (s *Shell) Process(ctx context.Context, line string) (string, error) {
	if s.conf.Timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.conf.Timeout)
		defer cancel()
	}

	switch line {
	case ".cluster":
		return s.processCluster(ctx)
	case ".leader":
		return s.processLeader(ctx)
	case ".help":
		return s.processHelp(), nil
	}
	return s.processQuery(ctx, line)
}

func (s *Shell) processHelp() string {
	return `
Dqlite shell is a simple interactive prompt for inspecting a dqlite database.
Enter a SQL statement to execute it, or one of the following built-in commands:

  .cluster                          Show the cluster membership
  .leader                           Show the current leader
`[1:]
}

func (s *Shell) processCluster(ctx context.Context) (string, error) {
	cli, err := s.conf.App.FindLeader(ctx)
	if err != nil {
		return "", err
	}
	cluster, err := cli.Cluster(ctx)
	if err != nil {
		return "", err
	}
	result := ""
	for i, server := range cluster {
		if i > 0 {
			result += "\n"
		}
		result += fmt.Sprintf("%x|%s|%s", server.ID, server.Address, server.Role)
	}

	return result, nil
}

func (s *Shell) processLeader(ctx context.Context) (string, error) {
	cli, err := s.conf.App.FindLeader(ctx)
	if err != nil {
		return "", err
	}
	leader, err := cli.Leader(ctx)
	if err != nil {
		return "", err
	}
	if leader == nil {
		return "", nil
	}
	return leader.Address, nil
}

func (s *Shell) processQuery(ctx context.Context, line string) (string, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin transaction: %w", err)
	}

	rows, err := tx.Query(line)
	if err != nil {
		err = fmt.Errorf("query: %w", err)
		if rbErr := tx.Rollback(); rbErr != nil {
			return "", fmt.Errorf("unable to rollback: %v", err)
		}
		return "", err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		err = fmt.Errorf("columns: %w", err)
		if rbErr := tx.Rollback(); rbErr != nil {
			return "", fmt.Errorf("unable to rollback: %v", err)
		}
		return "", err
	}
	n := len(columns)

	var sb strings.Builder
	writer := tabwriter.NewWriter(&sb, 0, 8, 1, '\t', 0)
	for _, col := range columns {
		fmt.Fprintf(writer, "%s\t", col)
	}
	fmt.Fprintln(writer)

	for rows.Next() {
		row := make([]interface{}, n)
		rowPointers := make([]interface{}, n)
		for i := range row {
			rowPointers[i] = &row[i]
		}

		if err := rows.Scan(rowPointers...); err != nil {
			err = fmt.Errorf("scan: %w", err)
			if rbErr := tx.Rollback(); rbErr != nil {
				return "", fmt.Errorf("unable to rollback: %v", err)
			}
			return "", err
		}

		for _, column := range row {
			fmt.Fprintf(writer, "%v\t", column)
		}
		fmt.Fprintln(writer)
	}

	if err := rows.Err(); err != nil {
		err = fmt.Errorf("rows: %w", err)
		if rbErr := tx.Rollback(); rbErr != nil {
			return "", fmt.Errorf("unable to rollback: %v", err)
		}
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return "", fmt.Errorf("flush: %w", err)
	}

	return strings.TrimRight(sb.String(), "\n"), nil
}
