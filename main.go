package main

import (
	"flag"
	"fmt"
	kitlog "github.com/go-kit/kit/log"
	"net/http"
	"os"
	"os/signal"
	dao "marathon/cargo-assistant/dao"
	svc "marathon/cargo-assistant/service"
	"syscall"
	ca "marathon/cargo-assistant"
	tp "marathon/cargo-assistant/transport"

)

func main() {
	var (
		mysqlUrl      = flag.String("mysqlUrl", "10.2.4.113", "")
		mysqlPort     = flag.String("mysqlPort", "3306", "")
		mysqlUsername = flag.String("mysqlUsername", "root", "")
		mysqlPassword = flag.String("mysqlPassword", "sfai", "")
		mysqlDBName   = flag.String("mysqlDBName", "marathon", "")
		httpAddr      = flag.String("addr", ":8081", "The address of listen and serve")
	)
	flag.Parse()
	var logger kitlog.Logger
	logger = kitlog.NewJSONLogger(os.Stderr)
	var groupDao dao.IGroupDao
	var proMktBaseDao dao.IProMarketBaseDao
	var joinDao dao.IJoinDao
	var groupService svc.IGroupService
	var joinService svc.IJoinService
	errs := make(chan error)
	var err error
	//init DB

	ca.NewMysqlManager(*mysqlUrl, *mysqlPort, *mysqlDBName, *mysqlUsername, *mysqlPassword,logger)
	groupDao, err = dao.NewGroupDao(logger)
	if err != nil {
		errs <- err
	}
	proMktBaseDao, err = dao.NewProMarketBaseDao(logger)
	if err != nil {
		errs <- err
	}
	joinDao,err = dao.NewJoinDao(logger)
	if err != nil {
		errs <- err
	}
	groupService = svc.NewGroupService(groupDao, proMktBaseDao)
	joinService = svc.NewJoinService(joinDao)
	httpHandler := tp.MakeHttpHandler(groupService, logger)
	joinHttpHandler := tp.MakeJoinHttpHandler(joinService, logger)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()
	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errs <- http.ListenAndServe(*httpAddr, httpHandler)
	}()
	logger.Log("exit:", <-errs)
}
