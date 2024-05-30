package service

import (
	"context"
	"errors"
	"log"

	"tradeTornado/internal/lib"
	"tradeTornado/internal/service/provider"
)

type IMigrationRegistry interface {
	RegisterMigration(string, lib.IMigrator)
	Run(ctx context.Context, name string) error
}

type MigrationRegistry struct {
	migrators map[string]lib.IMigrator
	gs        *provider.GormSession
}

func NewMigrationRegistry(gs *provider.GormSession) *MigrationRegistry {
	return &MigrationRegistry{
		migrators: map[string]lib.IMigrator{},
		gs:        gs,
	}
}

func (r *MigrationRegistry) RegisterMigration(name string, migrator lib.IMigrator) {
	r.migrators[name] = migrator
}

func (r *MigrationRegistry) Run(ctx context.Context, name string) error {
	return r.gs.RunTx(ctx, func() error {
		if name == "all" {
			for s, _ := range r.migrators {
				if err := r.migrators[s].Migrate(ctx); err != nil {
					return err
				} else {
					log.Println("#####migration " + s + " done")
				}
			}
			return nil
		} else {
			if _, ok := r.migrators[name]; !ok {
				return errors.New("migration not Found")
			}
			err := r.migrators[name].Migrate(ctx)
			if err != nil {
				return err
			} else {
				log.Println("#####migration " + name + " done")
				return nil
			}
		}
	})
}
