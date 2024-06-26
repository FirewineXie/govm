package cmd

import (
	"fmt"
	"github.com/FirewineXie/envm/internal/config"

	"github.com/urfave/cli"
	"os"
)

// Execute adds all child goCommands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	app := cli.NewApp()
	app.Name = "envm"
	app.Usage = "Any More Version Manager"
	app.Version = "v1.0.2"
	app.Description = `
			java & go  & node  version manager
     `

	app.Authors = []cli.Author{
		cli.Author{
			Name: "Firewine",
		},
	}
	app.Before = func(context *cli.Context) error {
		return config.VerifyEnv()
	}

	app.Commands = baseCommands

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "[g] %s\n", err.Error())
		os.Exit(1)
	}
}
