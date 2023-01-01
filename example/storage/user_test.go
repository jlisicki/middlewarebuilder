package storage

import (
	"context"
	"fmt"
	"os"
)

// ExampleUserRepositoryGet presents usage of middlewares to inject debug middlewares
// that allows to inspect cache and storage calls.
func ExampleUserRepositoryGet() {
	output := os.Stdout
	repo, err := NewUserRepository(output)
	if err != nil {
		panic(err)
	}
	ctx := ContextWithEnabledDebug(context.Background())
	fmt.Println("Create user")
	_ = repo.Set(ctx, User{
		ID:   "10",
		Name: "John",
	})
	fmt.Println("Populate cache")
	_, _ = repo.Get(ctx, "10")
	fmt.Println("Fetch from cache")
	_, _ = repo.Get(ctx, "10")
	// Output: Create user
	// [DEBUG][CacheCall] PreSet
	// [DEBUG][StorageCall] PreSet
	// Populate cache
	// [DEBUG][CacheCall] PreGet
	// [DEBUG][StorageCall] PreGet
	// Fetch from cache
	// [DEBUG][CacheCall] PreGet
}
