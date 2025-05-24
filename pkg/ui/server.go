package ui

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gitlab.com/Hamed1984/logistics/pkg/app"
	"gitlab.com/Hamed1984/logistics/pkg/conf"
	"gitlab.com/Hamed1984/logistics/pkg/db"
	"gitlab.com/Hamed1984/logistics/pkg/logging"
)

func StartServer() error {
	conf := conf.GetConffiguration()
	done := make(chan struct{})
	defer func() {
		close(done)
	}()
	appContext, err := app.NewContext(conf, done)
	if err != nil {
		logging.GetLogger(conf).Error(err.Error())
		return err
	}

	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/healthz", func(c echo.Context) error {
		type HealthOK struct {
			OK string `json:"ok"`
		}
		return c.JSON(http.StatusOK, &HealthOK{"OK"})
	})

	e.POST("/v1/customers", func(c echo.Context) error {
		cu := new(db.Customer)
		err := c.Bind(cu)
		if err != nil {
			resp := &MessageResp{
				Msg: "internal server err",
			}
			return c.JSON(http.StatusInternalServerError, resp)
		}
		ret, err := appContext.CreateCustomer(c.Request().Context(), cu)
		if err != nil {
			logging.GetLogger(conf).Error(err.Error())
			resp := &MessageResp{
				Msg: "internal server errir",
			}
			return c.JSON(http.StatusInternalServerError, resp)
		}
		resp := &MessageResp{
			Msg:  "success",
			Data: ret,
		}
		return c.JSON(http.StatusCreated, resp)
	})

	e.GET("/v1/customers", func(c echo.Context) error {
		ret, err := appContext.GetAllCustomers(c.Request().Context())
		if err != nil {
			logging.GetLogger(conf).Error(err.Error())
			resp := &MessageResp{
				Msg: "internal server error",
			}
			return c.JSON(http.StatusInternalServerError, resp)
		}
		resp := &MessageResp{
			Msg:  "success",
			Data: ret,
		}
		return c.JSON(http.StatusOK, resp)
	})

	e.POST("/v1/providers", func(c echo.Context) error {
		prov := new(db.Provider)
		if err := c.Bind(prov); err != nil {
			logging.GetLogger(conf).Error(err.Error())
			resp := &MessageResp{
				Msg: "internal server error",
			}
			return c.JSON(http.StatusInternalServerError, resp)
		}
		ret, err := appContext.CreateProvider(c.Request().Context(), prov)
		if err != nil {
			logging.GetLogger(conf).Error(err.Error())
			resp := &MessageResp{
				Msg: "internal server error",
			}
			return c.JSON(http.StatusInternalServerError, resp)
		}
		resp := &MessageResp{
			Msg:  "success",
			Data: ret,
		}
		return c.JSON(http.StatusCreated, resp)
	})

	e.GET("/v1/providers", func(c echo.Context) error {
		ret, err := appContext.GetAllProviders(c.Request().Context())
		if err != nil {
			logging.GetLogger(conf).Error(err.Error())
			resp := &MessageResp{
				Msg: "internal server error",
			}
			return c.JSON(http.StatusInternalServerError, resp)
		}
		resp := &MessageResp{
			Msg:  "success",
			Data: ret,
		}
		return c.JSON(http.StatusOK, resp)
	})

	e.POST("/v1/orders", func(c echo.Context) error {
		ord := new(db.Order)
		if err := c.Bind(ord); err != nil {
			logging.GetLogger(conf).Error(err.Error())
			resp := &MessageResp{
				Msg: "internal server error",
			}
			return c.JSON(http.StatusInternalServerError, resp)
		}
		ret, err := appContext.CreateOrder(c.Request().Context(), ord)
		if err != nil {
			logging.GetLogger(conf).Error(err.Error())
			resp := &MessageResp{
				Msg: "internal server error",
			}
			return c.JSON(http.StatusInternalServerError, resp)
		}
		resp := &MessageResp{
			Msg:  "success",
			Data: ret,
		}
		return c.JSON(http.StatusCreated, resp)
	})

	e.GET("/v1/orders/:orderid", func(c echo.Context) error {
		id := c.Param("orderid")
		orderID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			logging.GetLogger(conf).Error(err.Error())
			resp := &MessageResp{
				Msg: "order id must be a number",
			}
			return c.JSON(http.StatusBadRequest, resp)
		}
		ret, err := appContext.GetOrderByID(c.Request().Context(), orderID)
		if err != nil {
			logging.GetLogger(conf).Error(err.Error())
			resp := &MessageResp{
				Msg: "internal server error",
			}
			return c.JSON(http.StatusInternalServerError, resp)
		}
		resp := &MessageResp{
			Msg:  "success",
			Data: ret,
		}
		return c.JSON(http.StatusOK, resp)
	})

	e.PUT("/v1/orders/pickup/:orderid", func(c echo.Context) error {
		id := c.Param("orderid")
		orderID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			logging.GetLogger(conf).Error(err.Error())
			resp := &MessageResp{
				Msg: "order id must be a number",
			}
			return c.JSON(http.StatusBadRequest, resp)
		}
		err = appContext.PickupOrder(c.Request().Context(), orderID)
		if err != nil {
			if err == db.ErrOrderAlreadyDelivered || err == db.ErrOrderAlreadyPickedUp || err == db.ErrOrderExpired {
				resp := &MessageResp{
					Msg: err.Error(),
				}
				return c.JSON(http.StatusBadRequest, resp)
			}
			logging.GetLogger(conf).Error(err.Error())
			resp := &MessageResp{
				Msg: "internal server error",
			}
			return c.JSON(http.StatusInternalServerError, resp)
		}
		resp := &MessageResp{
			Msg: "success",
		}
		return c.JSON(http.StatusOK, resp)

	})

	e.GET("/v1/providers/deliver/average", func(c echo.Context) error {
		ret, err := appContext.GetAverageDeliveryByProvider(c.Request().Context())
		if err != nil {
			logging.GetLogger(conf).Error(err.Error())
			resp := &MessageResp{
				Msg: "internal server error",
			}
			return c.JSON(http.StatusInternalServerError, resp)
		}
		resp := &MessageResp{
			Msg:  "success",
			Data: ret,
		}
		return c.JSON(http.StatusOK, resp)
	})

	server := &http.Server{
		Addr: ":" + conf.ServerListenPort,
	}
	go func() {
		err := e.StartServer(server)
		if err != nil {
			logging.GetLogger(conf).Fatal(err.Error())
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		logging.GetLogger(conf).Fatal(err.Error())
	}
	defer func() {
		logging.GetLogger(conf).Sync()
	}()
	return err

}
