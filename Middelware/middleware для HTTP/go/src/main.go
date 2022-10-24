package main

import (
	"fmt"
	"net/http"
	"time"
)

func mainPage(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session_id")
	//учебный пример ! учебный пример это не проверка авторизации
	loggedIn := (err != http.ErrNoCookie)
	if loggedIn {
		fmt.Fprintf(w, `<a href="/logout">logout</a>`)
		fmt.Fprintln(w, "Welcome,"+session.Value)
	} else {
		fmt.Fprintf(w, `<a href="/login">login</a>`)
		fmt.Fprintln(w, "You need to login")
	}
}
func loginPage(w http.ResponseWriter, r *http.Request) {
	expiration := time.Now().Add(10 * time.Hour)
	cookie := http.Cookie{
		Name:    "session_id",
		Value:   "Roman",
		Expires: expiration,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}
func logoutPage(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session_id")
	if err != http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	session.Expires = time.Now().AddDate(0, 0, -1)
	http.SetCookie(w, session)
	http.Redirect(w, r, "/", http.StatusFound)

}
func adminIndex(w http.ResponseWriter, r *http.Request) { //админский список
	fmt.Fprintf(w, `<a href="/">site index</a>`)
	fmt.Fprintln(w, "Admin main page")
}

func PanicPage(w http.ResponseWriter, r *http.Request) { //вызов паники
	panic("this must me recovered")
}

func adminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("adminAuthMiddleware", r.URL.Path)
		_, err := r.Cookie("session_id")
		//учебный пример! это не проверка аунтификации
		if err != nil {
			fmt.Println("no auth at", r.URL.Path)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r) // вызов следущей цепочки
	})
}

func accessLogMiddleware(next http.Handler) http.Handler { // передаём http хендер возвращаем
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // ворощаем функцию
		fmt.Println("accessLogMiddleware", r.URL.Path)                                        // вывод какой миделвер отработал
		start := time.Now()                                                                   // засекаем время начала запроса
		next.ServeHTTP(w, r)                                                                  // вызов следущей цепочки
		fmt.Printf("[%s] %s, %s %s\n", r.Method, r.RemoteAddr, r.URL.Path, time.Since(start)) // вывод
	})
}

func panicMiddleware(next http.Handler) http.Handler { // передаём http хендер возвращаем
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // ворощаем функцию
		fmt.Println("panicMiddlevare", r.URL.Path)
		defer func() { //обработка паники
			if err := recover(); err != nil {
				fmt.Println("recovered", err)
				http.Error(w, "Internal server error", 500)
			}
		}()
		next.ServeHTTP(w, r) // вызов следущей цепочки
	})
}

func main() {
	adminNux := http.NewServeMux() //админский мультитексмат
	adminNux.HandleFunc("/admin", adminIndex)
	adminNux.HandleFunc("/panic", PanicPage)

	// set middleware
	adminHadler := adminAuthMiddleware(adminNux) //аторизация
	seteMux := http.NewServeMux()                // мултиплекстар
	seteMux.Handle("/admin", adminHadler)        //админом передаём хендер
	seteMux.HandleFunc("/login", loginPage)
	seteMux.HandleFunc("/logout", logoutPage)
	seteMux.HandleFunc("/", mainPage)
	// set middleware
	siteHandler := accessLogMiddleware(seteMux)
	siteHandler = panicMiddleware(siteHandler)

	fmt.Print("starting server at :8080")
	http.ListenAndServe(":8080", siteHandler)

}
