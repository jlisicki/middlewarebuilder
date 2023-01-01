package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
)

type (
	Identifier interface {
		comparable
	}
	Entity[K Identifier] interface {
		Identifier() K
	}
	Repository[T Entity[K], K Identifier] interface {
		Get(ctx context.Context, id K) (T, error)
		Set(ctx context.Context, entity T) error
		Delete(ctx context.Context, id K) error
	}

	serializer[T any] interface {
		Serialize(T) ([]byte, error)
		UnSerialize([]byte) (T, error)
	}
	// InMemoryRepository stores entities in local memory.
	InMemoryRepository[T Entity[K], K Identifier] struct {
		lock                 sync.Mutex
		entities             map[string][]byte
		identifierSerializer serializer[K]
		entitySerializer     serializer[T]
	}
	// Cache for repository in local memory.
	Cache[T Entity[K], K Identifier] struct {
		Next   Repository[T, K]
		cached map[K]T
		lock   sync.Mutex
	}
	// Telemetry for repository.
	Telemetry[T Entity[K], K Identifier] struct {
		Next Repository[T, K]
	}
	Debug[T Entity[K], K Identifier] struct {
		Next   Repository[T, K]
		Output io.Writer
		Label  string
	}
)

type debugEnablerCtxKey string

var debugEnabler debugEnablerCtxKey = "debug"

func ContextWithEnabledDebug(ctx context.Context) context.Context {
	return context.WithValue(ctx, debugEnabler, "enabled")
}

func (d Debug[T, K]) Get(ctx context.Context, id K) (T, error) {
	if _, ok := ctx.Value(debugEnabler).(string); ok {
		_, _ = fmt.Fprintf(d.Output, "[DEBUG][%s] PreGet\n", d.Label)
	}
	return d.Next.Get(ctx, id)
}

func (d Debug[T, K]) Set(ctx context.Context, entity T) error {
	if _, ok := ctx.Value(debugEnabler).(string); ok {
		_, _ = fmt.Fprintf(d.Output, "[DEBUG][%s] PreSet\n", d.Label)
	}
	return d.Next.Set(ctx, entity)
}

func (d Debug[T, K]) Delete(ctx context.Context, id K) error {
	if _, ok := ctx.Value(debugEnabler).(string); ok {
		_, _ = fmt.Fprintf(d.Output, "[DEBUG][%s] PreDelete\n", d.Label)
	}
	return d.Next.Delete(ctx, id)
}

func (c *Cache[T, K]) Get(ctx context.Context, id K) (T, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	entity, isCached := c.cached[id]
	if isCached {
		return entity, nil
	}
	entity, err := c.Next.Get(ctx, id)
	if err != nil {
		return entity, err
	}
	c.cached[entity.Identifier()] = entity
	return entity, nil
}

func (c *Cache[T, K]) Set(ctx context.Context, entity T) error {
	c.lock.Lock()
	delete(c.cached, entity.Identifier())
	c.lock.Unlock()
	return c.Next.Set(ctx, entity)
}

func (c *Cache[T, K]) Delete(ctx context.Context, id K) error {
	c.lock.Lock()
	delete(c.cached, id)
	c.lock.Unlock()
	return c.Next.Delete(ctx, id)
}

func (t Telemetry[T, K]) Get(ctx context.Context, id K) (T, error) {
	sT := time.Now()
	defer func() {
		// For now log values instead of applying changes to metrics.
		log.Printf("Get: %s", time.Since(sT))
	}()
	return t.Next.Get(ctx, id)
}

func (t Telemetry[T, K]) Set(ctx context.Context, entity T) error {
	sT := time.Now()
	defer func() {
		// For now log values instead of applying changes to metrics.
		log.Printf("Set: %s", time.Since(sT))
	}()
	return t.Next.Set(ctx, entity)
}

func (t Telemetry[T, K]) Delete(ctx context.Context, id K) error {
	sT := time.Now()
	defer func() {
		// For now log values instead of applying changes to metrics.
		log.Printf("Delete: %s", time.Since(sT))
	}()
	return t.Next.Delete(ctx, id)
}

func NewInMemoryRepository[T Entity[K], K Identifier](identitySerializer serializer[K], entitySerializer serializer[T]) *InMemoryRepository[T, K] {
	return &InMemoryRepository[T, K]{
		entities:             make(map[string][]byte),
		identifierSerializer: identitySerializer,
		entitySerializer:     entitySerializer,
	}
}

var errNotFound = errors.New("not found")

func (i *InMemoryRepository[T, K]) Get(ctx context.Context, id K) (T, error) {
	i.lock.Lock()
	defer i.lock.Unlock()
	var entity T
	key, err := i.identifierSerializer.Serialize(id)
	if err != nil {
		return entity, fmt.Errorf("unable to serialize identifier: %w", err)
	}
	raw, exists := i.entities[string(key)]
	if !exists {
		return entity, errNotFound
	}
	entity, err = i.entitySerializer.UnSerialize(raw)
	if err != nil {
		return entity, fmt.Errorf("unable to unserialize entity: %w", err)
	}
	return entity, nil
}

func (i *InMemoryRepository[T, K]) Set(ctx context.Context, entity T) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	key, err := i.identifierSerializer.Serialize(entity.Identifier())
	if err != nil {
		return fmt.Errorf("unable to serialize identifier: %w", err)
	}
	raw, err := i.entitySerializer.Serialize(entity)
	if err != nil {
		return fmt.Errorf("unable to serialize entity: %w", err)
	}
	i.entities[string(key)] = raw
	return nil
}

func (i *InMemoryRepository[T, K]) Delete(ctx context.Context, id K) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	key, err := i.identifierSerializer.Serialize(id)
	if err != nil {
		return fmt.Errorf("unable to serialize identifier: %w", err)
	}
	delete(i.entities, string(key))
	return nil
}
