package subpub

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type MessageHandler func(msg interface{})

type Subscription interface {
	Unsubscribe()
}

type SubPub interface {
	Subscribe(subject string, cb MessageHandler) (Subscription, error)
	Publish(subject string, msg interface{}) error
	Close(ctx context.Context) error
}

type subscription struct {
	subject string
	id      uuid.UUID
	ps      *PubSub
}

func (s *subscription) Unsubscribe() {
	s.ps.mu.Lock()
	defer s.ps.mu.Unlock()

	subscribers, ok := s.ps.subscribers[s.subject]
	if !ok {
		return
	}

	// Удаляем подписчика по UUID
	delete(subscribers, s.id)

	// Если подписчиков больше нет, удаляем тему
	if len(subscribers) == 0 {
		delete(s.ps.subscribers, s.subject)
	}
}

// PubSub - конкретная реализация SubPub интерфейса
type PubSub struct {
	subscribers map[string]map[uuid.UUID]MessageHandler
	mu          sync.Mutex
	wg          sync.WaitGroup
	closed      bool
}

func NewSubPub() *PubSub {
	return &PubSub{
		subscribers: make(map[string]map[uuid.UUID]MessageHandler),
	}
}

func (ps *PubSub) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return nil, context.Canceled
	}

	if _, ok := ps.subscribers[subject]; !ok {
		ps.subscribers[subject] = make(map[uuid.UUID]MessageHandler)
	}

	id := uuid.New()

	ps.subscribers[subject][id] = cb

	return &subscription{
		subject: subject,
		id:      id,
		ps:      ps,
	}, nil
}

func (ps *PubSub) Publish(subject string, msg interface{}) error {
	ps.mu.Lock()

	if ps.closed {
		ps.mu.Unlock()
		return context.Canceled
	}

	var handlers []MessageHandler
	if subs, ok := ps.subscribers[subject]; ok {
		handlers = make([]MessageHandler, 0, len(subs))
		for _, handler := range subs {
			handlers = append(handlers, handler)
		}
	}
	ps.mu.Unlock()

	if len(handlers) > 0 {
		ps.wg.Add(len(handlers)) // Увеличиваем счетчик для каждого обработчика
		for _, handler := range handlers {
			go func(h MessageHandler) {
				defer ps.wg.Done()
				h(msg) // Вызываем обработчик в отдельной горутине
			}(handler)
		}
	}

	return nil
}

func (ps *PubSub) Close(ctx context.Context) error {
	ps.mu.Lock()
	ps.closed = true
	ps.mu.Unlock()

	done := make(chan struct{})

	go func() {
		ps.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
