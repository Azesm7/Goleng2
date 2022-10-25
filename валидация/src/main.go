package main

import (
	"fmt"
	"net/http"

	// валидатор
	"github.com/asaskevich/govalidator"
	// парсинг параметров в структуру
	"github.com/gorilla/schema"
)

// http://127.0.0.1:8080/?to=v.romanov@corp.mail.ru&priority=low&subject=Hello!&inner=ignored&id=12&flag=23
type SendMessage struct {
	ID        int    `valid:",optional"`             //оно может быть или не быть
	Priority  string `	valid:"in(low|normal|hight)"` //валидатор ин и сами значения которые могут быть
	Recipient string `schema:"to"	valid:"email"`     //схемой указываем что поле должно называться ту и валидатором будет емэил
	Subject   string `valid:"msgSubject"`            //собственый валидатор который зарегистрирован для этих целей
	Inner     string `schema:"-"	valid:"-"`          // нечего не парсит нечего не валидирует
	flag      int
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("request" + r.URL.String() + "\n\n")) // вывод точки запроса

	msg := &SendMessage{} // создаём соообщение

	decoder := schema.NewDecoder()  // создаём декодер для того чтобы распарсить из входящих параметров в мою структуру
	decoder.IgnoreUnknownKeys(true) // на не известные поля не надо ругаться
	err := decoder.Decode(msg, r.URL.Query())
	if err != nil {
		fmt.Println(err)
		http.Error(w, "internal", 500)
		return
	}

	_, err = govalidator.ValidateStruct(msg) // вызываем валидейт струтуру

	if err != nil {
		if allErrs, ok := err.(govalidator.Errors); ok { // пытаемся достучаться до интерфейса  и до реальных ошибок
			for _, fld := range allErrs.Errors() { // перебор ошибок
				data := []byte(fmt.Sprintf("field:	%#v\n\n", fld))
				w.Write(data)
			}
		}
		// просто вывод ошибки
		w.Write([]byte(fmt.Sprintf("error:	%s\n\n", err)))
	} else {
		w.Write([]byte(fmt.Sprintf("msg is correct\n\n")))

	}
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("starting server	:8080")
	http.ListenAndServe(":8080", nil)
}

func init() { // функция для того чтобы показать что был заригистрирован валидатор
	govalidator.CustomTypeTagMap.Set("msgSubject", govalidator.CustomTypeValidator(func(i interface{}, o interface{}) bool {
		sibject, ok := i.(string) //регистация  собственого
		if !ok {                  // попытка преобразование в строку
			return false
		}
		if len(sibject) == 0 || len(sibject) > 10 {
			return false
		}
		return false
	}))
}
