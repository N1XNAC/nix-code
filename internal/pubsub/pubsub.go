package pubsub

import (
	"sync"
)

type EventType string

const (
	EventSessionCreated EventType = "session.created"
	EventMessageAdded   EventType = "message.added"
	EventMessageDelta   EventType = "message.delta"
	EventToolStarted    EventType = "tool.started"
	EventToolCompleted  EventType = "tool.completed"
	EventModeChanged    EventType = "mode.changed"
)

type Event struct {
	Type    EventType
	Data    map[string]any
	Session string
}

type Bus struct {
	mu         sync.RWMutex
	subscribers map[EventType][]chan Event
}

func NewBus() *Bus {
	return &Bus{
		subscribers: make(map[EventType][]chan Event),
	}
}

func (b *Bus) Subscribe(eventType EventType, ch chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[eventType] = append(b.subscribers[eventType], ch)
}

func (b *Bus) Unsubscribe(eventType EventType, ch chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	subs := b.subscribers[eventType]
	for i, sub := range subs {
		if sub == ch {
			b.subscribers[eventType] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
}

func (b *Bus) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	subs := b.subscribers[event.Type]
	for _, ch := range subs {
		select {
		case ch <- event:
		default:
		}
	}
}
