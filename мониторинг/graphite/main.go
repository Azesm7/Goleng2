package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"time"

	_ "expvar"
)

func hendler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello word"))
}

var carbonAddr = flag.String("graphite", "192.168.99.100:2003", "The address of carbon receiver")

func main() {
	flag.Parse()
	go sendStat()
	http.HandleFunc("/", hendler)
	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}

func sendStat() {
	m := &runtime.MemStats{}
	conn, err := net.Dial("tcp", *carbonAddr)
	if err != nil {
		fmt.Println("not connect tcp at sendStat")
		return
	}
	c := time.Tick(time.Minute)
	for tickTime := range c {
		runtime.ReadMemStats(m)                                                                      // читаем статистику по памяти
		buf := bytes.NewBuffer([]byte{})                                                             // создаём новый буфер и туда записываем значенме
		fmt.Fprintf(buf, "coursera.mem_heap %d %d\n", m.HeapInuse, tickTime.Unix())                  // количество памяти в хипи
		fmt.Fprintf(buf, "coursera.mem_Stack %d %d\n", m.StackInuse, tickTime.Unix())                // количество памяти в стеки
		fmt.Fprintf(buf, "coursera.goruntines_num %d %d\n", runtime.NumGoroutine(), tickTime.Unix()) //количество гоурутин
		conn.Write(buf.Bytes())
		fmt.Println(buf.String())
	}
}
