package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	_ "github.com/go-sql-driver/mysql"
)

type Item struct {
	Id          int
	Title       string
	Description string
	Updated     sql.NullString
}

type Handler struct {
	DB   *sql.DB
	Tmpl *template.Template
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	items := []*Item{} // создаём слайс объектов  для общего списка

	rows, err := h.DB.Query("SELECT id, title, updated FROM items") //  запрос в бд
	__err_panic(err)
	for rows.Next() {
		post := &Item{}
		err = rows.Scan(&post.Id, &post.Title, &post.Updated) //распоковка данных с бд
		__err_panic(err)
		items = append(items, post) //добавили в общий список
	}
	// надо закрывать соединение, иначе будет течь
	rows.Close()

	err = h.Tmpl.ExecuteTemplate(w, "index.html", struct { //експант шаблона
		Items []*Item
	}{
		Items: items,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h *Handler) AddForm(w http.ResponseWriter, r *http.Request) {
	err := h.Tmpl.ExecuteTemplate(w, "create.html", nil) // экспнсим шаблон
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	// в целям упрощения примера пропущена валидация
	result, err := h.DB.Exec("INSERT INTO items (`title`,`description`) VALUES (?,?)", //запрос в бд
		r.FormValue("title"),       // предаём значение
		r.FormValue("description")) // предаём значение(поле)
	__err_panic(err)

	affected, err := result.RowsAffected() //количество записей с которыми было обновление
	__err_panic(err)

	lastID, err := result.LastInsertId() // номер ставленой записи
	__err_panic(err)

	fmt.Println("Insert - RowsAffected", affected, "LastInsertId: ", lastID)
	http.Redirect(w, r, "/", http.StatusFound)

}

func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)                 //получаем с web
	id, err := strconv.Atoi(vars["id"]) // конвертируем в id
	__err_panic(err)

	post := &Item{}
	// QueryRow сам закрывает коннект
	row := h.DB.QueryRow("SELECT id, title, updated, description FROM items WHERE id =?", id) // выбераем запись

	err = row.Scan(&post.Id, &post.Title, &post.Updated, &post.Description) //отсканировал запись
	__err_panic(err)

	err = h.Tmpl.ExecuteTemplate(w, "edit.html", post) // предал в форму
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)                 //получаем с web
	id, err := strconv.Atoi(vars["id"]) // конвертируем в id
	__err_panic(err)

	// в целям упрощения примера пропущена валидация
	result, err := h.DB.Exec(
		"UPDATE items SET"+
			"`title` = ?"+
			",`description` = ?"+
			",`updated`=?"+
			"WHERE id = ?",
		r.FormValue("title"),
		r.FormValue("description"),
		"rvasily",
		id,
	)
	__err_panic(err)

	affected, err := result.RowsAffected() //количество записей с которыми было обновление
	__err_panic(err)

	fmt.Println("Update - RowsAffected", affected)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	__err_panic(err)

	result, err := h.DB.Exec("DELETE FROM items WHERE id = ?", id)
	__err_panic(err)
	affected, err := result.RowsAffected()
	__err_panic(err)

	fmt.Println("Delete - RowsAffected", affected)

	w.Header().Set("Content-type", "application/json")
	resp := `{"affected": ` + strconv.Itoa(int(affected)) + `}`
	w.Write([]byte(resp))
}

func main() {
	// основные настройки к базе
	dsn := "root:Roman1125./@tcp(localhost:3306)/onli?"
	// указываем кодировку
	dsn += "&charset=utf8"
	// отказываемся от prapared statements
	// параметры подставляются сразу
	dsn += "&interpolateParams=true"

	db, err := sql.Open("mysql", dsn) //подключение к базе данных

	db.SetMaxOpenConns(10) // количество подключений

	err = db.Ping() // вот тут будет первое подключение к базе
	if err != nil {
		panic(err)
	}

	Handlers := &Handler{
		DB:   db,
		Tmpl: template.Must(template.ParseGlob("crud_templates/*")),
	}

	// в целям упрощения примера пропущена авторизация и csrf
	r := mux.NewRouter()
	r.HandleFunc("/", Handlers.List).Methods("GET")
	r.HandleFunc("/items", Handlers.List).Methods("GET")
	r.HandleFunc("/items/new", Handlers.AddForm).Methods("GET")
	r.HandleFunc("/items/new", Handlers.Add).Methods("POST")
	r.HandleFunc("/items/{id}", Handlers.Edit).Methods("GET")
	r.HandleFunc("/items/{id}", Handlers.Update).Methods("POST")
	r.HandleFunc("/items/{id}", Handlers.Delete).Methods("DELETE")

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", r)

}

// не используйте такой код в прошакшене
// ошибка должна всегда явно обрабатываться
func __err_panic(err error) {
	if err != nil {
		panic(err)
	}
}
