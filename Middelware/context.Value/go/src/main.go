package main

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// сколько в среднем спим при эмуляции работы
const AvgSleep = 50

func trackContextTimings(ctx context.Context, metricName string, start time.Time) {
	// получаем тайминги из контекста
	// поскольку там пустой интерфейс, то нам надо преобразовать к нужному типу
	timings, ok := ctx.Value(timingsKey).(*ctxTimings)
	if !ok { //ели нет контекст тайминга
		return
	}
	elapsed := time.Since(start) // просмотор сколько прошло времени
	// лочимся на случай конкурентной записи в мапку
	timings.Lock()
	defer timings.Unlock()
	// если меткри ещё нет - мы её создадим, если есть - допишем в существующую
	if metric, metricExist := timings.Data[metricName]; !metricExist {
		timings.Data[metricName] = &Timing{
			Count:    1,
			Duration: elapsed,
		}
	} else {
		metric.Count++             // плюсуем время
		metric.Duration += elapsed // плюсуем количество
	}
}

type Timing struct {
	Count    int
	Duration time.Duration
}

type ctxTimings struct {
	sync.Mutex
	Data map[string]*Timing
}

// линтер ругается если используем базовые типы в Value контекста
// типа так безопаснее разграничивать
type key int

const timingsKey key = 1

func logContextTimings(ctx context.Context, path string, start time.Time) {
	// получаем тайминги из контекста
	// поскольку там пустой интерфейс, то нам надо преобразовать к нужному типу
	timings, ok := ctx.Value(timingsKey).(*ctxTimings) // проверяем тайминги
	if !ok {
		return
	}
	totalReal := time.Since(start) // смотрим время начало запроса
	buf := bytes.NewBufferString(path)
	var total time.Duration
	for timing, value := range timings.Data { //начинаем интератироваться по всем таймингам
		total += value.Duration
		buf.WriteString(fmt.Sprintf("\n\t%s(%d): %s", timing, value.Count, value.Duration)) // пишем что это был за тайминг ,сколько его количество было , сколько времени занимал
	}
	buf.WriteString(fmt.Sprintf("\n\ttotal: %s", totalReal))      // сколько времени всего прошло
	buf.WriteString(fmt.Sprintf("\n\ttracked: %s", total))        // сколько времени учтено
	buf.WriteString(fmt.Sprintf("\n\tunkn: %s", totalReal-total)) // сколько времени не учтено

	fmt.Println(buf.String())
}

func timingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//ctx := r.Context()// получаем контекст
		ctx := context.WithValue(r.Context(),
			timingsKey, //указываем ключ
			&ctxTimings{ //структура
				Data: make(map[string]*Timing),
			})
		defer logContextTimings(ctx, r.URL.Path, time.Now()) //внутри функция куда передаём контекст  Url  и время запроса
		next.ServeHTTP(w, r.WithContext(ctx))                //тобслуживаем запрос

	})
}

func emulateWork(ctx context.Context, workName string) {
	defer trackContextTimings(ctx, workName, time.Now()) //внутри функция куда передаём контекст имя работы , время начала работы

	rnd := time.Duration(rand.Intn(AvgSleep))
	time.Sleep(time.Millisecond * rnd) //выполнится запрос после слипа
}

func loadPostsHandle(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()             //получаем контекст
	emulateWork(ctx, "checkCache")   //функция для эмулатции запросов в неё передаём контекст и имя работы
	emulateWork(ctx, "loadPosts")    //функция для эмулатции запросов в неё передаём контекст и имя работы
	emulateWork(ctx, "loadPosts")    //функция для эмулатции запросов в неё передаём контекст и имя работы
	emulateWork(ctx, "loadPosts")    //функция для эмулатции запросов в неё передаём контекст и имя работы
	time.Sleep(1 * time.Microsecond) // создание таймера
	emulateWork(ctx, "loadSidebar")  //функция для эмулатции запросов в неё передаём контекст и имя работы
	emulateWork(ctx, "loadComments") //функция для эмулатции запросов в неё передаём контекст и имя работы

	fmt.Fprintln(w, "Request done") //вывод

}
func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	siteMux := http.NewServeMux()            //мултитекст запросов
	siteMux.HandleFunc("/", loadPostsHandle) //регистрируем обработчик

	siteHandle := timingMiddleware(siteMux) // регистрируем делевер

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", siteHandle)
}
