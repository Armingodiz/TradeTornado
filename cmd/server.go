package cmd

import (
	configs "tradeTornado/config"
	"tradeTornado/internal/lib"
	"tradeTornado/internal/service/wiring"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Server() {
	cnf := configs.ConfigFromEnv()
	c := wiring.NewContainer(cnf)
	if err := c.Run(lib.Terminable()); err != nil {
		logrus.Errorln(err)
	}
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "run",
	Short: "Run",
	Run: func(cmd *cobra.Command, args []string) {
		Server()
	},
}
