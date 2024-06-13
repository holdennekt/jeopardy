package wsLobby

import (
	"github.com/holdennekt/sgame/api/ws"
	"github.com/holdennekt/sgame/api/ws/lobby/events"
)

func handleWsMessage(pubSubConn *ws.PubSubConn, msg ws.InternalMessage) {
	switch msg.Event {
	case events.CHAT:
		events.HandleWsChatMessage(pubSubConn, msg)
	}
}
