package cmd

import (
	configs "tradeTornado/config"
	"tradeTornado/internal/lib"
	"tradeTornado/internal/service/wiring"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Migrate(name string) {
	cnf := configs.ConfigFromEnv()
	cn := wiring.NewContainer(cnf)
	err := cn.RunMigration(lib.Terminable(), name)
	if err != nil {
		log.Errorln(err)
	}
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "run migrate name",
	Run: func(cmd *cobra.Command, args []string) {
		Migrate(args[0])
	},
}
