// pkg/stdlib/websocket.go
// WebSocket utilities for the Xxlang standard library.
// These functions are designed for WebSocket server mode.
package stdlib

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/topxeq/xxlang/pkg/objects"
)

// websocketUpgrader is the default WebSocket upgrader
var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func init() {
	Register(&Module{
		Name: "websocket",
		Exports: map[string]objects.Object{
			// upgrade upgrades an HTTP connection to WebSocket.
			// Usage: websocket.upgrade(request, response)
			"upgrade": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("upgrade() takes exactly 2 arguments")
				}

				req, ok := args[0].(*objects.HttpReq)
				if !ok {
					return Error("first argument to upgrade must be HTTP_REQ")
				}

				resp, ok := args[1].(*objects.HttpResp)
				if !ok {
					return Error("second argument to upgrade must be HTTP_RESP")
				}

				if req.Value == nil || resp.Value == nil {
					return Error("http request or response is nil")
				}

				conn, err := websocketUpgrader.Upgrade(resp.Value, req.Value, nil)
				if err != nil {
					return Error("websocket upgrade failed: " + err.Error())
				}

				return objects.NewWebSocket(conn)
			}),

			// readMsg reads a message from a WebSocket connection.
			// Usage: websocket.readMsg(conn)
			"readMsg": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("readMsg() takes exactly 1 argument")
				}

				ws, ok := args[0].(*objects.WebSocket)
				if !ok {
					return Error("argument to readMsg must be WEBSOCKET")
				}

				return ws.ReadMessage()
			}),

			// sendText sends a text message over a WebSocket connection.
			// Usage: websocket.sendText(conn, text)
			"sendText": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("sendText() takes exactly 2 arguments")
				}

				ws, ok := args[0].(*objects.WebSocket)
				if !ok {
					return Error("first argument to sendText must be WEBSOCKET")
				}

				text, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument to sendText must be STRING")
				}

				return ws.SendTextMessage(text.Value)
			}),

			// sendBinary sends a binary message over a WebSocket connection.
			// Usage: websocket.sendBinary(conn, data)
			"sendBinary": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("sendBinary() takes exactly 2 arguments")
				}

				ws, ok := args[0].(*objects.WebSocket)
				if !ok {
					return Error("first argument to sendBinary must be WEBSOCKET")
				}

				data, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument to sendBinary must be STRING")
				}

				return ws.SendBinaryMessage(data.Value)
			}),

			// sendClose sends a close message over a WebSocket connection.
			// Usage: websocket.sendClose(conn)
			"sendClose": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("sendClose() takes exactly 1 argument")
				}

				ws, ok := args[0].(*objects.WebSocket)
				if !ok {
					return Error("argument to sendClose must be WEBSOCKET")
				}

				return ws.SendCloseMessage()
			}),

			// close closes a WebSocket connection.
			// Usage: websocket.close(conn)
			"close": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("close() takes exactly 1 argument")
				}

				ws, ok := args[0].(*objects.WebSocket)
				if !ok {
					return Error("argument to close must be WEBSOCKET")
				}

				return ws.Close()
			}),

			// isWebSocket checks if a value is a WebSocket connection.
			// Usage: websocket.isWebSocket(value)
			"isWebSocket": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isWebSocket() takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.WebSocket)
				return Bool(ok)
			}),
		},
	})
}
