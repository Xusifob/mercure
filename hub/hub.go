package hub

import (
	"log"
)

type serializedUpdate struct {
	update *Update
	event  string
}

func newSerializedUpdate(u *Update) *serializedUpdate {
	return &serializedUpdate{u, u.String()}
}

// Hub stores channels with clients currently subcribed
type Hub struct {
	options            *Options
	subscribers        map[chan *serializedUpdate]struct{}
	newSubscribers     chan chan *serializedUpdate
	removedSubscribers chan chan *serializedUpdate
	updates            chan *serializedUpdate
}

// NewHubFromEnv creates a hub fusing the configuration set in env vars
func NewHubFromEnv() (*Hub, error) {
	options, err := NewOptionsFromEnv()
	if err != nil {
		return nil, err
	}

	return NewHub(options), nil
}

// NewHub creates a hub
func NewHub(options *Options) *Hub {
	return &Hub{
		options,
		make(map[chan *serializedUpdate]struct{}),
		make(chan (chan *serializedUpdate)),
		make(chan (chan *serializedUpdate)),
		make(chan *serializedUpdate),
	}
}

// Start starts the hub
func (h *Hub) Start() {
	go func() {
		for {
			select {

			case s := <-h.newSubscribers:
				h.subscribers[s] = struct{}{}

			case s := <-h.removedSubscribers:
				delete(h.subscribers, s)
				close(s)

			case serializedUpdate, ok := <-h.updates:
				for s := range h.subscribers {
					if ok {
						s <- serializedUpdate
					} else {
						close(s)
					}
				}
				if ok {
					log.Printf("Broadcasted topics %s", serializedUpdate.update.Topics)
				} else {
					return
				}
			}
		}
	}()
}

// Stop stops disconnect all connected clients
func (h *Hub) Stop() {
	close(h.updates)
}
