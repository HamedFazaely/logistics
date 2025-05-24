package db

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"gitlab.com/Hamed1984/logistics/pkg/conf"
)

func TestShit(t *testing.T) {
	t.Setenv("DB_PASSWORD", "hamed1984")
	cnfg := conf.GetConffiguration()
	db, err := sql.Open("mysql", cnfg.GetDBDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		t.Fatal(err)
	}
	res, err := db.Exec("INSERT INTO customers (first_name, last_name, email) VALUES ('bangi','faz','hfaza@gmail.com')")
	if err != nil {
		t.Fatal(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(id)

}

func TestGetAllCustomers(t *testing.T) {
	t.Setenv("DB_PASSWORD", "hamed1984")
	cnfg := conf.GetConffiguration()
	db, err := sql.Open("mysql", cnfg.GetDBDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewCustomerRepoImpl(db)
	data, err := repo.GetAllCustomers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range data {
		t.Log(x)
	}
}

func TestCreateCustomer(t *testing.T) {
	t.Setenv("DB_PASSWORD", "hamed1984")
	cnfg := conf.GetConffiguration()
	db, err := sql.Open("mysql", cnfg.GetDBDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	c := Customer{
		FirstName: "ali",
		LastName:  "banci",
		Email:     "x@gmail.com",
	}
	repo := NewCustomerRepoImpl(db)
	res, err := repo.CreateCustomer(context.Background(), &c)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res.ID)
}

func TestCreateProvider(t *testing.T) {
	t.Setenv("DB_PASSWORD", "hamed1984")
	cnfg := conf.GetConffiguration()
	db, err := sql.Open("mysql", cnfg.GetDBDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	respo := NewProverRepoImpl(db)
	prov := Provider{
		Name:     "nnn",
		StatusEP: "http://localhost/status",
		PickupEP: "http://localhost/pickup",
	}
	res, err := respo.CreateProvider(context.Background(), &prov)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res.ID)
}

func TestGetAllProvidors(t *testing.T) {
	t.Setenv("DB_PASSWORD", "hamed1984")
	cnfg := conf.GetConffiguration()
	db, err := sql.Open("mysql", cnfg.GetDBDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewProverRepoImpl(db)
	res, err := repo.GetAllProviders(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range res {
		t.Log(x)
	}
}

func TestCreateOrder(t *testing.T) {
	t.Setenv("DB_PASSWORD", "hamed1984")
	cnfg := conf.GetConffiguration()
	db, err := sql.Open("mysql", cnfg.GetDBDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewOrderRepoImpl(db, cnfg)
	ord := &Order{
		CutomerID:     1,
		ProviderID:    1,
		SenderPhone:   "9197459057",
		ReceiverPhone: "9197459057",
		Address:       "my address",
	}
	res, err := repo.CreateOrder(context.Background(), ord)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res.ID, res.Status)
}

func TestGetOrderByID(t *testing.T) {
	t.Setenv("DB_PASSWORD", "hamed1984")
	cnfg := conf.GetConffiguration()
	db, err := sql.Open("mysql", cnfg.GetDBDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewOrderRepoImpl(db, cnfg)
	ord, err := repo.GetOrderByID(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ord)
}

func TestSetOrderPickedUp(t *testing.T) {
	t.Setenv("DB_PASSWORD", "hamed1984")
	cnfg := conf.GetConffiguration()
	db, err := sql.Open("mysql", cnfg.GetDBDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewOrderRepoImpl(db, cnfg)
	res, err := repo.SetOrderPickedUp(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}

func TestSetOrderStatus(t *testing.T) {
	t.Setenv("DB_PASSWORD", "hamed1984")
	cnfg := conf.GetConffiguration()
	db, err := sql.Open("mysql", cnfg.GetDBDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewOrderRepoImpl(db, cnfg)
	err = repo.UpdateOrderStatus(context.Background(), 2, InProgress)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetSMSSent(t *testing.T) {
	t.Setenv("DB_PASSWORD", "hamed1984")
	cnfg := conf.GetConffiguration()
	db, err := sql.Open("mysql", cnfg.GetDBDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewOrderRepoImpl(db, cnfg)
	err = repo.SetPickupSMSSent(context.Background(), 1)
	if err!=nil {
		t.Fatal(err)
	}
}
