package main

import (
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	nats "github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
	"github.com/spf13/viper"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
	"yuriy_test/models"
)

func ReadConfig() error {
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	viper.SetConfigType("json")

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	return nil
}

func parseMsg(m *stan.Msg) (*models.Order, error) {
	order := &models.Order{}
	err := json.Unmarshal(m.Data, order)
	if err != nil {
		log.Printf("err = %s\n", err.Error())
	}
	//Дебажный вывод сообщения
	//log.Printf("%v\n\n", order)

	return order, err
}

const search = `<!DOCTYPE html>
<html>
<head>
 <meta charset="utf-8">
 <title>Поиск заказа</title>
</head>
<body>
 <form>
  <p><input type="search" name="uuid" placeholder="Поиск по айди">
  <input type="submit"  formaction="/search?q" value="Найти заказ"></p>
 </form>
</body>
</html>`

func searchHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, search)
}

func loadCache(orderMap map[string]models.Order, mu *sync.Mutex, db *sqlx.DB) error {
	orders, err := LoadOrders(db)
	if err != nil {
		log.Fatal("Error in LoadOrders")
	}
	for _, order := range orders {
		mu.Lock()
		orderMap[order.OrderUid] = order
		mu.Unlock()
	}
	return err
}

func main() {
	err := ReadConfig()
	if err != nil {
		log.Fatal("Failed to read config")
	}
	// Connect to DB
	db, err := DbConnect()
	if err != nil {
		log.Fatal("Failed to connect to DB")
	}
	//кэш
	ordersMap := make(map[string]models.Order)
	mu := sync.Mutex{}

	err = loadCache(ordersMap, &mu, db)
	if err != nil {
		log.Fatal("Failed to load Cache")
	}

	//Загружаем шаблон
	tmpl, err := template.New("").ParseFiles("index.html")
	if err != nil {
		panic(err)
	}
	//Хэндлеры сервера
	http.HandleFunc("/", searchHandler)
	http.HandleFunc("/search",
		func(w http.ResponseWriter, r *http.Request) {
			qmap := r.URL.Query()
			uuid := qmap["uuid"]
			order, ok := ordersMap[uuid[0]]
			if !ok {
				fmt.Fprintln(w, "<html lang=\"en\"><a href=\"/\">на главную</a></html>")
				fmt.Fprintln(w, "<div>Нет такого заказа</div>")
				for key, _ := range ordersMap {
					fmt.Fprintf(w, "<div>%q is the order uuid</div></html>", key)
				}
				return
			}
			err = tmpl.ExecuteTemplate(w, "index.html", struct{ models.Order }{order})
			if err != nil {
				panic(err)
			}
		})

	//Запускаем сервер
	log.Println("starting server at :8080")
	go http.ListenAndServe(":8080", nil)

	//Получаем параметры из конфига
	name := viper.GetString("nats.name")
	clusterID := viper.GetString("nats.cluster")
	clientID := viper.GetString("nats.client")
	durable := viper.GetString("nats.durable")
	subj := viper.GetString("nats.subject")

	opts := []nats.Option{nats.Name(name)}
	URL := stan.DefaultNatsURL

	// Connect to NATS
	nc, err := nats.Connect(URL, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	sc, err := stan.Connect(clusterID, clientID, stan.NatsConn(nc),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatalf("Connection lost, reason: %v", reason)
		}))
	if err != nil {
		log.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, URL)
	}

	// Process Subscriber Options.
	mcb := func(msg *stan.Msg) {
		order, err := parseMsg(msg)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		mu.Lock()
		ordersMap[order.OrderUid] = *order
		mu.Unlock()
		err = StoreOrder(db, order)
		//manual Ack
		err = msg.Ack()
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
	}
	aw, _ := time.ParseDuration("15m")
	_, err = sc.Subscribe(subj, mcb, stan.DurableName(durable), stan.SetManualAckMode(), stan.AckWait(aw))
	if err != nil {
		sc.Close()
		log.Fatal(err)
	}

	// Wait for a SIGINT (perhaps triggered by user with CTRL-C)
	// Run cleanup when signal is received
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for range signalChan {
			fmt.Printf("\nReceived an interrupt, closing connection...\n\n")
			sc.Close()
			cleanupDone <- true
		}
	}()
	<-cleanupDone
}
