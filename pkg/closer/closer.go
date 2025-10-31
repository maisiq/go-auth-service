package closer

import (
	"os"
	"os/signal"
	"sync"

	"github.com/maisiq/go-auth-service/internal/logger"
)

type Closer struct {
	done chan interface{}
	wait []func() error
	mu   sync.Mutex
	once sync.Once
}

func (c *Closer) Add(fn ...func() error) {
	c.mu.Lock()
	c.wait = append(c.wait, fn...)
	c.mu.Unlock()
}

func New(sig ...os.Signal) *Closer {
	c := &Closer{done: make(chan interface{}, 1)}

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, sig...)
		<-ch
		signal.Stop(ch)
		c.CloseAll()
	}()
	return c
}

func (c *Closer) CloseAll() {
	c.once.Do(func() {
		defer close(c.done)
		c.mu.Lock()
		funcs := c.wait
		c.wait = nil
		c.mu.Unlock()

		log := logger.GetLogger()

		errs := make(chan error, len(funcs))

		for _, fn := range funcs {
			go func(fn func() error) {
				errs <- fn()
			}(fn)
		}

		for i := 0; i < cap(errs); i++ {
			if err := <-errs; err != nil {
				log.Errorf("Error while closing: %w", err)
			}
		}
	})
}

func (c *Closer) Wait() {
	<-c.done
}
