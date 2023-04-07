package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/eapache/queue"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var broadcast = make(chan *User) // 用來傳送訊息的廣播頻道
var box *queue.Queue

type User struct {
	uuid string
	conn *websocket.Conn
	send *websocket.Conn
	Message
}

// 訊息結構
type Message struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

var upgrader = websocket.Upgrader{}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/ws", wsHandler)
	box = queue.New() // 創建一個新的佇列
	go handleMessages()
	log.Fatal(http.ListenAndServe(":8087", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// 這裡可以處理聊天室的前端介面
	fmt.Fprint(w, "Welcome to the chat room!")
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	user := &User{
		uuid: uuid.New().String(),
		conn: conn,
	}
	if box.Length() != 0 {
		item := box.Peek().(*User)
		item.send = user.conn
		user.send = item.conn
		box.Remove()
	} else {
		box.Add(user)
	}

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			conn.Close()
			break
		}
		user.Message = msg
		// 將收到的訊息加入廣播頻道
		broadcast <- user
	}

}

func handleMessages() {
	for {
		// 從廣播頻道中接收訊息
		user := <-broadcast

		err := user.send.WriteJSON(user.Message)
		if err != nil {
			log.Println(err)
			user.conn.Close()
		}

	}
}
