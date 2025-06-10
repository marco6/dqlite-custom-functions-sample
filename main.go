package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/marco6/dqlite-custom-functions-sample/shell"
	"github.com/peterh/liner"
	"github.com/spf13/cobra"

	"github.com/canonical/go-dqlite/v3/app"
)

func main() {
	var address string
	var cluster []string
	var dataDir string
	var timeoutMsec uint

	cmd := &cobra.Command{
		Use:   "dqlite-custom-functions-sample [options] <database>",
		Short: "Standard dqlite shell with custom functions",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := app.New(dataDir,
				app.WithAddress(address),
				app.WithCluster(cluster),
			)
			if err != nil {
				return err
			}
			if err := app.Ready(context.Background()); err != nil {
				return err
			}
			sh, err := shell.New(&shell.ShellConfig{
				App:      app,
				Database: args[0],
				Timeout:  time.Duration(timeoutMsec) * time.Millisecond,
			})
			if err != nil {
				return err
			}

			line := liner.NewLiner()
			defer line.Close()
			for {
				input, err := line.Prompt("dqlite> ")
				if err != nil {
					if err == io.EOF {
						break
					}
					return err
				}

				result, err := sh.Process(context.Background(), input)
				if err != nil {
					fmt.Println("Error: ", err)
				} else {
					line.AppendHistory(input)
					if result != "" {
						fmt.Println(result)
					}
				}
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&address, "address", "a", "127.0.0.1:9001", "node address")
	flags.StringSliceVarP(&cluster, "join", "j", nil, "comma-separated list of db servers")
	flags.StringVarP(&dataDir, "data-dir", "d", "data", "data directory")
	flags.UintVar(&timeoutMsec, "timeout", 2000, "timeout of each request (msec)")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
