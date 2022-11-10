package main

import (
	"fmt"
)

var SessManager *SessionManager

func main() {
	SessManager = NewSessionManager()
	// создаем сессию
	sessId, err := SessManager.Create(
		&Session{
			Login:     "Roman",
			Useragent: "index",
		})
	fmt.Println("sessId", sessId, err)

	// проверяем сессию
	sess := SessManager.Check(
		&SessionID{
			ID: sessId.ID,
		})
	fmt.Println("sess", sess)
	// удаляем сессию
	SessManager.Delete(
		&SessionID{
			ID: sessId.ID,
		})

	sess = SessManager.Check(
		&SessionID{
			ID: sessId.ID,
		})
	fmt.Println("sess", sess)

}
