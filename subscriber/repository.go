package main

import (
	"fmt"
	"log"
	"yuriy_test/models"

	//_ "database/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

const itemCols = "chrt_id,track_number,price,rid,name," +
	"sale,item_size,total_price,nm_id,brand,status"

const paymentCols = "transaction, request_id, currency, provider, amount, payment_dt," +
	" bank, delivery_cost, goods_total, custom_fee "

const deliveryCols = "name,phone,zip,city,address,region,email"

const orderCols = "order_uid,track_number,entry," +
	"locale,internal_signature,customer_id," +
	"delivery_service,shardkey,sm_id,date_created,oof_shard"

func DbConfig() (string, error) {
	user := viper.GetString("postgres.user")
	dbname := viper.GetString("postgres.dbname")
	password := viper.GetString("postgres.password")
	host := viper.GetString("postgres.host")
	port := viper.GetInt("postgres.port")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	return psqlInfo, nil
}

func DbConnect() (*sqlx.DB, error) {
	connStr, err := DbConfig()
	if err != nil {
		return nil, err
	}
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return db, err
	}
	err = db.Ping()
	if err != nil {
		return db, err
	}
	return db, err
}

func SelectItems(db *sqlx.DB, id string) ([]models.Item, error) {
	items := []models.Item{}
	schema := "SELECT " +
		itemCols +
		" FROM items WHERE orderid = $1"
	err := db.Select(&items, schema, id)
	if err != nil {
		log.Println(err.Error())
		return items, err
	}
	return items, err
}

func GetPayment(db *sqlx.DB, id string) (*models.Payment, error) {
	payment := models.Payment{}
	schema := "SELECT " + paymentCols +
		" FROM payment WHERE orderid = $1"
	err := db.Get(&payment, schema, id)
	if err != nil {
		log.Println(err.Error())
		return &payment, err
	}
	return &payment, err
}

func GetDelivery(db *sqlx.DB, id string) (*models.Delivery, error) {
	delivery := models.Delivery{}
	schema := "SELECT " + deliveryCols +
		" FROM delivery WHERE orderid = $1"
	err := db.Get(&delivery, schema, id)
	if err != nil {
		log.Println(err.Error())
		return &delivery, err
	}
	return &delivery, err
}

func GetOrder(db *sqlx.DB, uuid string) (*models.Order, error) {
	order := models.Order{}
	schema := "SELECT " + orderCols +
		" FROM orders WHERE order_uid = $1"
	err := db.Get(&order, schema, uuid)
	if err != nil {
		log.Println(err.Error())
		return &order, err
	}
	return &order, err
}

func GetOrderFull(db *sqlx.DB, uuid string) (*models.Order, error) {
	order, err := GetOrder(db, uuid)
	if err != nil {
		log.Println("error in GetOrder")
		return order, err
	}

	delivery, err := GetDelivery(db, uuid)
	if err != nil {
		log.Println("error in GetDelivery")
		return order, err
	}
	order.Delivery = *delivery

	payment, err := GetPayment(db, uuid)
	if err != nil {
		log.Println("error in GetPayment")
		return order, err
	}
	order.Payment = *payment

	order.Items, err = SelectItems(db, uuid)
	if err != nil {
		log.Println("error in SelectItems")
		return order, err
	}

	return order, nil
}

func InsertOrder(db *sqlx.DB, order *models.Order) error {
	schema := "INSERT INTO orders (" + orderCols +
		") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)" +
		"RETURNING id;"

	args := make([]interface{}, 11)
	args[0] = order.OrderUid
	args[1] = order.TrackNumber
	args[2] = order.Entry
	args[3] = order.Locale
	args[4] = order.InternalSignature
	args[5] = order.CustomerId
	args[6] = order.DeliveryService
	args[7] = order.Shardkey
	args[8] = order.SmId
	args[9] = order.DateCreated
	args[10] = order.OofShard
	_, err := db.Exec(schema, args...)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func InsertPayment(db *sqlx.DB, order *models.Order) error {
	schema := "INSERT INTO payment (orderid," + paymentCols +
		") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);"
	args := make([]interface{}, 11)
	pay := order.Payment
	args[0] = order.OrderUid
	args[1] = pay.Transaction
	args[2] = pay.RequestId
	args[3] = pay.Currency
	args[4] = pay.Provider
	args[5] = pay.Amount
	args[6] = pay.PaymentDt
	args[7] = pay.Bank
	args[8] = pay.DeliveryCost
	args[9] = pay.GoodsTotal
	args[10] = pay.CustomFee
	_, err := db.Exec(schema, args...)
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func InsertDelivery(db *sqlx.DB, order *models.Order) error {
	schema := "INSERT INTO delivery (orderid," + deliveryCols +
		") VALUES ($1, $2, $3, $4, $5, $6, $7, $8);"

	args := make([]interface{}, 8)
	delivery := order.Delivery
	args[0] = order.OrderUid
	args[1] = delivery.Name
	args[2] = delivery.Phone
	args[3] = delivery.Zip
	args[4] = delivery.City
	args[5] = delivery.Address
	args[6] = delivery.Region
	args[7] = delivery.Email
	_, err := db.Exec(schema, args...)
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func InsertItems(db *sqlx.DB, order *models.Order) error {
	schema := "INSERT INTO items (orderid," +
		itemCols + ") VALUES"

	items := order.Items
	args := make([]interface{}, 1+len(items)*11)
	args[0] = order.OrderUid
	for i, item := range items {
		values := "($1, "
		args[i*11+1] = item.ChrtId
		values += `$` + fmt.Sprint(i*11+2) + `, `
		args[i*11+2] = item.TrackNumber
		values += `$` + fmt.Sprint(i*11+3) + `, `
		args[i*11+3] = item.Price
		values += `$` + fmt.Sprint(i*11+4) + `, `
		args[i*11+4] = item.Rid
		values += `$` + fmt.Sprint(i*11+5) + `, `
		args[i*11+5] = item.Name
		values += `$` + fmt.Sprint(i*11+6) + `, `
		args[i*11+6] = item.Sale
		values += `$` + fmt.Sprint(i*11+7) + `, `
		args[i*11+7] = item.Size
		values += `$` + fmt.Sprint(i*11+8) + `, `
		args[i*11+8] = item.TotalPrice
		values += `$` + fmt.Sprint(i*11+9) + `, `
		args[i*11+9] = item.NmId
		values += `$` + fmt.Sprint(i*11+10) + `, `
		args[i*11+10] = item.Brand
		values += `$` + fmt.Sprint(i*11+11) + `, `
		args[i*11+11] = item.Status
		values += `$` + fmt.Sprint(i*11+12)
		if i < len(items)-1 {
			values = values + "),"
		}
		if i == len(items)-1 {
			values = values + ");"
		}
		schema = schema + values
	}
	_, err := db.Exec(schema, args...)
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func StoreOrder(db *sqlx.DB, order *models.Order) error {
	err := InsertOrder(db, order)
	if err != nil {
		log.Println("Error in InsertOrder")
		return err
	}
	err = InsertPayment(db, order)
	if err != nil {
		log.Println("Error in InsertPayment")
		return err
	}

	err = InsertDelivery(db, order)
	if err != nil {
		log.Println("Error in InsertDelivery")
		return err
	}

	err = InsertItems(db, order)
	if err != nil {
		log.Println("Error in InsertItems")
		return err
	}
	return nil
}

//лучше переписать на дженериках, передавая название таблицы
func GetOrders(db *sqlx.DB) ([]models.Order, error) {
	orders := []models.Order{}
	//в продакшне загружать чанками, через limit и offset
	schema := "SELECT " + orderCols +
		" FROM orders ORDER BY order_uid"
	err := db.Select(&orders, schema)
	if err != nil {
		log.Println(err.Error())
		return orders, err
	}
	return orders, err
}

func GetItems(db *sqlx.DB) ([]models.Item, error) {
	items := []models.Item{}
	//в продакшне загружать чанками, через limit и offset
	schema := "SELECT " + itemCols +
		" FROM items ORDER BY orderid"
	err := db.Select(&items, schema)
	if err != nil {
		log.Println(err.Error())
		return items, err
	}
	return items, err
}

func CountItems(db *sqlx.DB) (map[string]int, error) {
	countMap := make(map[string]int)
	//в продакшне загружать чанками, через limit и offset
	schema := "SELECT orderid, count(*)" +
		" FROM items GROUP BY orderid ORDER BY orderid"
	rows, err := db.Query(schema)
	if err != nil {
		log.Println(err.Error())
		return countMap, err
	}
	for rows.Next() {
		var uid string
		var count int
		err = rows.Scan(&uid, &count)
		countMap[uid] = count
	}
	err = rows.Err()
	return countMap, err
}

func GetPayments(db *sqlx.DB) ([]models.Payment, error) {
	payments := []models.Payment{}
	//в продакшне загружать чанками, через limit и offset
	schema := "SELECT " + paymentCols +
		" FROM payment ORDER BY orderid"
	err := db.Select(&payments, schema)
	if err != nil {
		log.Println(err.Error())
		return payments, err
	}
	return payments, err
}

func GetDeliveries(db *sqlx.DB) ([]models.Delivery, error) {
	deliveries := []models.Delivery{}
	//в продакшне загружать чанками, через limit и offset
	schema := "SELECT " + deliveryCols +
		" FROM delivery ORDER BY orderid"
	err := db.Select(&deliveries, schema)
	if err != nil {
		log.Println(err.Error())
		return deliveries, err
	}
	return deliveries, err
}

func LoadOrders(db *sqlx.DB) ([]models.Order, error) {
	orders, err := GetOrders(db)
	if err != nil {
		log.Println("Error in GetOrders")
		return orders, err
	}

	deliveries, err := GetDeliveries(db)
	if err != nil {
		log.Println("Error in GetDeliveries")
		return orders, err
	}

	payments, err := GetPayments(db)
	if err != nil {
		log.Println("Error in GetPayments")
		return orders, err
	}

	items, err := GetItems(db)
	if err != nil {
		log.Println("Error in GetItems")
		return orders, err
	}

	cMap, err := CountItems(db)
	if err != nil {
		log.Println("Error in CountItems")
		return orders, err
	}
	k := 0
	for i, order := range orders {
		orders[i].Payment = payments[i]
		orders[i].Delivery = deliveries[i]
		count := cMap[order.OrderUid]
		for count > 0 {
			orders[i].Items = append(order.Items, items[k])
			k++
			count--
		}
	}
	return orders, err
}
