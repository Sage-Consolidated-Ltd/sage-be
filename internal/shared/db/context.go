package db

import "context"

type executorKey struct{}

func WithExecutor(ctx context.Context, exec Executor) context.Context {
	return context.WithValue(ctx, executorKey{}, exec)
}

func (db *DB) Executor(ctx context.Context) Executor {
	exec, ok := ctx.Value(executorKey{}).(Executor)
	if ok {
		return exec
	}
	return db.DB
}
