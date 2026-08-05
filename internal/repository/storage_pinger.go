package repository

import "context"

type StoragePinger interface {
	Ping(ctx context.Context) error
}
