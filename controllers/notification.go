package controllers

import (
	// . "creativelab/ecleave-dev/models"

	"github.com/gorilla/websocket"
	// "io/ioutil"
	// "net/http"
	// "strings"
)

type NotificationController struct {
	*BaseController
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// func (c *NotificationController) GetNotification(k *knot.WebContext) interface{} {
// 	k.Config.OutputType = knot.OutputJson

// 	conn, err := upgrader.Upgrade(k.Writer, k.Request, nil)

// 	if err != nil {
// 		fmt.Println(err)
// 		return err
// 	}

// 	for {
// 		_, msg, err := conn.ReadMessage()
// 		if err != nil {
// 			fmt.Println(err)
// 			return err
// 		}
// 		if string(msg) == "ping" {
// 			dashboard := DashboardController(*c)
// 			data, err := dashboard.PendingApproveDeclineRequest(k)
// 			// fmt.Println("ping")
// 			if err != nil {
// 				conn.Close()
// 				return err
// 			}
// 			time.Sleep(2 * time.Second)
// 			err = conn.WriteJSON(data)
// 			if err != nil {
// 				fmt.Println(err)
// 				return err
// 			}
// 		} else {
// 			conn.Close()
// 			fmt.Println(string(msg))
// 			return err
// 		}
// 	}

// 	return "hshshshshs"
// }

// func handleIO(currentConn *WebSocketConnection, connections []*WebSocketConnection) {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			log.Println("ERROR", fmt.Sprintf("%v", r))
// 		}
// 	}()

// 	for {
// 		fmt.Println("----------- masuk notif")
// 	}
// }
