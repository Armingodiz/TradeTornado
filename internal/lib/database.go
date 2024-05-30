package lib

import (
	"context"
)

type IMigrator interface {
	Migrate(ctx context.Context) error
}

type ISeeder interface {
	ApplySeeds(ctx context.Context) error
}
