package wsLobby

import (
	"github.com/holdennekt/sgame/api/ws"
	"github.com/holdennekt/sgame/api/ws/lobby/events"
)

func handleRdsMessage(wsConn *ws.WsConn, msg ws.InternalMessage) {
	switch msg.Event {
	case events.CHAT:
		handleRdsChatMessage(wsConn, msg)
	case events.LOBBY_ROOM:
		handleRdsLobbyRoomMessage(wsConn, msg)
	case events.ROOM_DELETED:
		handleRdsRoomDeletedMessage(wsConn, msg)
	}
}

func handleRdsChatMessage(wsConn *ws.WsConn, msg ws.InternalMessage) {
	chatMessage, err := events.NewChatMessage(msg)
	if err != nil {
		wsConn.PublishError(err)
	}
	wsConn.Publish(chatMessage.ToMessage())
}

func handleRdsLobbyRoomMessage(wsConn *ws.WsConn, msg ws.InternalMessage) {
	wsConn.Publish(ws.Message{Event: events.LOBBY_ROOM, Payload: msg.Payload})
}

func handleRdsRoomDeletedMessage(wsConn *ws.WsConn, msg ws.InternalMessage) {
	wsConn.Publish(ws.Message{Event: events.ROOM_DELETED, Payload: msg.Payload})
}
