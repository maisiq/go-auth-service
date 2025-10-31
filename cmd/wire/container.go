package wire

import (
	"context"
	"reflect"
	"sync"
	"syscall"

	"github.com/maisiq/go-auth-service/pkg/closer"
)

type Constructor[T any] func(*DIContainer) T
type ConfigPath string

type DIContainer struct {
	services      sync.Map
	constructors  sync.Map
	closer        *closer.Closer
	mu            sync.RWMutex
	watcher       <-chan interface{}
	cancelWatcher context.CancelFunc
	rebuildch     chan interface{}
	wait          <-chan interface{}
	once          sync.Once
}

func New(configPath string) *DIContainer {
	closer := closer.New(syscall.SIGTERM, syscall.SIGINT)
	c := &DIContainer{
		closer: closer,
	}

	Provide(c, func(d *DIContainer) ConfigPath {
		return ConfigPath(configPath)
	})

	return c
}

func Provide[T any](c *DIContainer, constructor Constructor[T]) {
	var zero T
	key := getKey(zero)
	c.constructors.Store(key, constructor)
}

func Get[T any](c *DIContainer) T {
	var zero T
	key := getKey(zero)

	if service, exists := c.services.Load(key); exists {
		return service.(T)
	}

	constructor, exists := c.constructors.Load(key)
	if !exists {
		panic("service not registered: " + key)
	}

	fn := constructor.(Constructor[T])
	service := fn(c)
	c.services.Store(key, service)

	return service
}

func ProvideNamed[T any](c *DIContainer, name string, constructor Constructor[T]) {
	key := getKey(*new(T)) + ":" + name
	c.constructors.Store(key, constructor)
}

func GetNamed[T any](c *DIContainer, name string) T {
	key := getKey(*new(T)) + ":" + name

	if service, exists := c.services.Load(key); exists {
		return service.(T)
	}

	constructor, exists := c.constructors.Load(key)
	if !exists {
		panic("named service not registered: " + key)
	}

	fn := constructor.(Constructor[T])
	service := fn(c)
	c.services.Store(key, service)

	return service
}

func getKey[T any](_ T) string {
	key := reflect.TypeOf((*T)(nil)).Elem().String()
	return key
}

func (c *DIContainer) AddToCloser(fn func() error) {
	c.closer.Add(fn)
}

func (c *DIContainer) Rebuild() <-chan interface{} {
	return c.rebuildch
}

func (c *DIContainer) ShutdownResources() {
	c.services.Clear()
	c.closer.CloseAll()
	c.closer.Wait()
}

func (c *DIContainer) RebuildOn(ctx context.Context, ch <-chan interface{}) {
	if c.cancelWatcher != nil {
		c.cancelWatcher() // there is no goroutine atm, but for safety
	}

	dctx, cancel := context.WithCancel(ctx) // not used yet
	c.cancelWatcher = cancel
	c.watcher = ch

	rebuildch := make(chan interface{}, 1)
	c.rebuildch = rebuildch

	go func() {
		for {
			select {
			case _, ok := <-c.watcher:
				if !ok {
					return
				}
				c.rebuildch <- "call"
				c.reload()
			case <-dctx.Done():
				return
			}
		}
	}()
}

func (c *DIContainer) reload() {
	c.ShutdownResources()
	c.mu.Lock()
	c.closer = closer.New(syscall.SIGTERM, syscall.SIGINT)
	c.mu.Unlock()
}
