package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/icrowley/fake"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024, // ограничение на чтение
	WriteBufferSize: 1024, // ограничение на запись
	CheckOrigin: func(r *http.Request) bool { //функция позволяющая стучаться
		return true
	},
}

func sendNewMsgNotifications(client *websocket.Conn) {
	ticker := time.NewTicker(3 * time.Second) // создание тикера
	for {
		w, err := client.NextWriter(websocket.TextMessage) //получаем новый пакет с данными
		if err != nil {                                    // обработка ошибки
			ticker.Stop() // выключение  тикера
			break
		}
		msg := newMassage() // запуск функции по генерации  фейковых сообщений и запись их в переменую
		w.Write(msg)        // передача сообщения
		w.Close()           // закрытие

		<-ticker.C // ждём следущего сообщения
	}
}
func main() {
	tmpl := template.Must(template.ParseFiles("index.html")) // парсим html код

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { // страница в корне
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/notifications", func(w http.ResponseWriter, r *http.Request) { // сраница  с уведомлениями
		ws, err := upgrader.Upgrade(w, r, nil) // сообщаем что мжно стучаться
		if err != nil {                        // обработка ошибки
			log.Fatal(err)
		}
		go sendNewMsgNotifications(ws) // вызов функции в одельной go рутине и предача ей значения
	})
	fmt.Println("starting server at :8080") // вывод
	http.ListenAndServe(":8080", nil)       // запуск сервера

}
func newMassage() []byte { // создание фейковых сообщений
	data, _ := json.Marshal(map[string]string{
		"email":   fake.EmailAddress(),
		"name":    fake.FirstName() + " " + fake.LastName(),
		"subject": fake.Product() + " " + fake.Model(),
	})
	return data

}
