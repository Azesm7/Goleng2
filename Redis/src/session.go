package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"

	"github.com/garyburd/redigo/redis"
)

type Session struct {
	Login     string
	Useragent string
}

type SessionID struct {
	ID string
}

const sessKeyLen = 10

type SessionManager struct {
	redisConn redis.Conn // соединение до редиса
}

func NewSessionManager(conn redis.Conn) *SessionManager {
	return &SessionManager{
		redisConn: conn,
	}
}

func (sm *SessionManager) Create(in *Session) (*SessionID, error) { // при попытке создать сесию
	id := SessionID{RandStringRunes(sessKeyLen)}                           // мы создаём какойто ключ
	dataSerialized, _ := json.Marshal(in)                                  // стилиризуем данные в json
	mkey := "sessions:" + id.ID                                            //сохранение логина юзер агента
	data, err := sm.redisConn.Do("SET", mkey, dataSerialized, "EX", 86400) //сохранение значения команда do у саединения передаёт все значение ключ данные и опции
	result, err := redis.String(data, err)                                 //преобразование в стринг данных
	if err != nil {
		return nil, err
	}
	if result != "OK" {
		return nil, fmt.Errorf("result not OK")
	}
	return &id, nil
}

func (sm *SessionManager) Check(in *SessionID) *Session {
	mkey := "sessions:" + in.ID                            //сохранение логина юзер агента
	data, err := redis.Bytes(sm.redisConn.Do("GET", mkey)) //получаем слайс байт
	if err != nil {                                        // получаем данные
		log.Println("cant get data:", err)
		return nil
	}
	sess := &Session{}
	err = json.Unmarshal(data, sess) // распоковываем в нём json
	if err != nil {
		log.Println("cant unpack session data:", err)
		return nil
	}
	return sess
}

func (sm *SessionManager) Delete(in *SessionID) {
	mkey := "sessions:" + in.ID
	_, err := redis.Int(sm.redisConn.Do("DEL", mkey)) //команда делит будет говорить сколько записей было удалено
	if err != nil {
		log.Println("redis error:", err)
	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
