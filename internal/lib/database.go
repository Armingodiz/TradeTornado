package lib

import (
	"context"
)

type IMigrator interface {
	Migrate(ctx context.Context) error
}
