package main

import (
	"github.com/stretchr/testify/assert"
	sqlxmock "github.com/zhashkevych/go-sqlxmock"
	"regexp"
	"strings"
	"testing"
	"time"
	"yuriy_test/models"
)

func TestGetOrders(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	or := models.Order{
		OrderUid:    "1d58768b-0e69-40a4-bd66-b5988187e39b",
		TrackNumber: "odit",
		Entry:       "ea",
		Delivery: models.Delivery{
			Name:    "Lula Swift",
			Phone:   "7342199469",
			Zip:     "26425",
			City:    "Santa Ana",
			Address: "Fork, 34",
			Region:  "United States Minor Outlying Islands",
			Email:   "clintshields@paucek.name",
		},
		Payment: models.Payment{
			Transaction:  "2b03d6c6-8d52-4cd2-a367-2fd7eed4358c",
			RequestId:    "ae9e3e9a-42b1-44c2-83c7-06cacc4eda0f",
			Currency:     "SVC",
			Provider:     "People Power",
			Amount:       96,
			PaymentDt:    18,
			Bank:         "Investormill",
			DeliveryCost: 22,
			GoodsTotal:   98,
			CustomFee:    34,
		},
		Items:             nil,
		Locale:            "lb",
		InternalSignature: "explicabo",
		CustomerId:        "99239695-e48a-418d-b215-a913fe77ea76",
		DeliveryService:   "Abt Associates",
		Shardkey:          "reiciendis",
		SmId:              98,
		DateCreated:       time.Time{},
		OofShard:          "officiis",
	}

	rows0 := sqlxmock.NewRows(strings.Split(orderCols, ",")).
		AddRow(or.OrderUid, or.TrackNumber, or.Entry,
			or.Locale, or.InternalSignature, or.CustomerId,
			or.DeliveryService, or.Shardkey, or.SmId, or.DateCreated,
			or.OofShard)

	query0 := "SELECT " + orderCols +
		" FROM orders ORDER BY order_uid"

	mock.ExpectQuery(regexp.QuoteMeta(query0)).WillReturnRows(rows0)

	d := or.Delivery
	rows1 := sqlxmock.NewRows(strings.Split(deliveryCols, ",")).
		AddRow(d.Name, d.Phone, d.Zip, d.City,
			d.Address, d.Region, d.Email)

	query1 := "SELECT " + deliveryCols +
		" FROM delivery ORDER BY orderid"

	mock.ExpectQuery(regexp.QuoteMeta(query1)).WillReturnRows(rows1)

	p := or.Payment
	rows2 := sqlxmock.NewRows(strings.Split(paymentCols, ",")).
		AddRow(p.Transaction, p.RequestId,
			p.Currency, p.Provider, p.Amount, p.PaymentDt,
			p.Bank, p.DeliveryCost, p.GoodsTotal, p.CustomFee)

	query2 := "SELECT " + paymentCols +
		" FROM payment ORDER BY orderid"

	mock.ExpectQuery(regexp.QuoteMeta(query2)).WillReturnRows(rows2)

	rows3 := sqlxmock.NewRows(strings.Split(itemCols, ","))

	query3 := "SELECT " + itemCols +
		" FROM items ORDER BY orderid"

	mock.ExpectQuery(regexp.QuoteMeta(query3)).WillReturnRows(rows3)

	rows4 := sqlxmock.NewRows([]string{"orderid", "count(*)"})

	query4 := "SELECT orderid, count(*)" +
		" FROM items GROUP BY orderid ORDER BY orderid"

	mock.ExpectQuery(regexp.QuoteMeta(query4)).WillReturnRows(rows4)

	orders, err := LoadOrders(db)

	assert.NoError(t, err)
	assert.NotNil(t, orders)
	assert.Equal(t, orders[0], or)
}
