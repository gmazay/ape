package main

import (
	"ape/conf"
	"ape/controller"
	"ape/db"
	"database/sql"
	//"github.com/kpango/glg"
	"ape/glg"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	//"fmt"
	//"log"
	"os"
	"os/signal"
	"syscall"
	//"crypto/tls"
)

var dbh *sql.DB
var cfg *conf.Config

func init() {
	var err error
	cfg, err = cfg.Read()
	if err != nil {
		panic("Read config file error: " + err.Error())
	}

	dbh = db.Connect(cfg, "main")
}

func main() {

	//f, err := os.OpenFile("/tmp/ape.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	//if err != nil { log.Fatal(err) }
	//defer f.Close()
	///log.SetOutput(f)
	//lg := log.New(f, "", log.LstdFlags)

	infolog := glg.FileWriter(cfg.Infolog, 0644)
	errlog := glg.FileWriter(cfg.Errlog, 0644)
	defer infolog.Close()
	defer errlog.Close()
	//glg.Get().
	log := glg.New().
		SetMode(glg.BOTH).
		AddLevelWriter(glg.INFO, infolog).
		AddLevelWriter(glg.LOG, infolog).
		AddLevelWriter(glg.PRINT, infolog).
		AddLevelWriter(glg.WARN, errlog).
		AddLevelWriter(glg.ERR, errlog)
	//log.Infof("Start server on %s", cfg.Listen)

	c := controller.NewController(cfg, dbh, log)
	router := fasthttprouter.New()

	//router.GET("/gws", c.GetGws)
	//router.GET("/home/*filepath", c.Dispatcher)

	make_routes(c, cfg, router)

	s := fasthttp.Server{
		// https://godoc.org/github.com/valyala/fasthttp#Server
		Name:          "Ape",
		Handler:       router.Handler,
		MaxConnsPerIP: cfg.MaxConnsPerIP,
		Logger:        log,
	}

	//-- Soft reload config by SIGURG, stop by other --//
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGURG, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	go func() {
		var err error
		for {
			sig := <-sigs
			log.Infof("Got signal: %s", sig)

			switch sig {
			case syscall.SIGURG:
			    log.Info("Reload config")
				cfg, err = cfg.Read()
				if err != nil {
					log.Errorf("Read config file error: " + err.Error())
					continue
				}
				c = controller.NewController(cfg, dbh, log)
				router = fasthttprouter.New()

				make_routes(c, cfg, router)

				s.Handler = router.Handler
			default:
			    dbh.Close()
			    log.Info("Stop server")
			    os.Exit(0)
			}

		}
	}()
	//------------------------------------------------//

	//err := s.ListenAndServe(cfg.Listen)
	//certFile := "/opt/ape/miatel.ru.crt"
	//keyFile := "/opt/ape/miatel.ru.key"
	//fmt.Println( cfg.CertFile, cfg.KeyFile, cfg.MaxConnsPerIP )
	var err error
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		log.Infof("Start server on HTTPS %s", cfg.Listen)
		err = s.ListenAndServeTLS(cfg.Listen, cfg.CertFile, cfg.KeyFile) // Listen HTTPS
	} else {
		log.Infof("Start server on HTTP %s", cfg.Listen)
		err = s.ListenAndServe(cfg.Listen)                               // Listen HTTP
	}
	if err != nil {
		log.Errorf("ListenAndServe: %s", err)
	}

}

func make_routes(c *controller.Controller, cfg *conf.Config, router *fasthttprouter.Router) {
	for path, route := range cfg.Route {
		for method, _ := range route.Method {
			switch method {
			case "GET":		router.GET (path, c.Dispatcher(path))
			case "POST":	router.POST(path, c.Dispatcher(path))
			case "PUT":		router.PUT (path, c.Dispatcher(path))
			case "DELETE":	router.PUT (path, c.Dispatcher(path))
			}
		}
	}
}
