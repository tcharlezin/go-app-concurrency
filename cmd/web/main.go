package main

import (
	"database/sql"
	"encoding/gob"
	"final-project/data"
	"fmt"
	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "80"

func main() {

	// Connect to the database
	db := initDB()
	db.Ping()

	// Create sessions
	session := initSession()

	// Create loggers
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Create channels

	// Create a wait group
	wg := sync.WaitGroup{}

	// Setup the application config
	app := Config{
		Session:       session,
		DB:            db,
		Wait:          &wg,
		InfoLog:       infoLog,
		ErrorLog:      errorLog,
		Models:        data.New(db),
		ErrorChan:     make(chan error),
		ErrorChanDone: make(chan bool),
	}

	// Setup mail
	app.Mailer = app.CreateMail()
	go app.listenForMail()

	// Listen for signals
	go app.ListenForShutdown()

	// Listen for errors
	go app.listenForErrors()

	// Listen for web connections
	app.serve()
}

func (app *Config) listenForErrors() {
	for {
		select {
		case err := <-app.ErrorChanDone:
			app.ErrorLog.Println(err)
		case <-app.ErrorChanDone:
			return
		}
	}
}

func (app *Config) serve() {
	// start http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	app.InfoLog.Println("Starting web server...")
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func initDB() *sql.DB {
	conn := connectToDb()

	if conn == nil {
		log.Panicln("Can't connect to the database!")
	}

	return conn
}

func connectToDb() *sql.DB {
	counts := 0

	dsn := os.Getenv("DSN")

	for {
		conn, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres is not yet ready...")
		} else {
			log.Println("Connected to database...")
			return conn
		}

		if counts > 10 {
			return nil
		}

		log.Println("Backign off for 1 second")
		time.Sleep(1 * time.Second)
		counts++
		continue
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initSession() *scs.SessionManager {
	// Define as a object valid to store in session
	gob.Register(data.User{})

	session := scs.New()
	session.Store = redisstore.New(initRedis())
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true

	return session
}

func initRedis() *redis.Pool {
	redispool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", os.Getenv("REDIS"))
		},
	}

	return redispool
}

func (app *Config) ListenForShutdown() {
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.Shutdown()
	os.Exit(0)
}

func (app *Config) Shutdown() {
	// Perform any cleaup tasks
	app.InfoLog.Println("Would run cleanup tasks...")

	// Block until waitgroup is empty
	app.Wait.Wait()

	app.Mailer.DoneChan <- true
	app.ErrorChanDone <- true

	app.InfoLog.Println("Closing channels and shutdown application...")
	close(app.Mailer.MailerChan)
	close(app.Mailer.ErrorChan)
	close(app.Mailer.DoneChan)
	close(app.ErrorChan)
	close(app.ErrorChanDone)
}

func (app *Config) CreateMail() Mail {
	// create channels
	errorChan := make(chan error)
	mailerChan := make(chan Message, 100)
	mailerDoneChan := make(chan bool)

	m := Mail{
		Domain:      "localhost",
		Host:        "localhost",
		Port:        1025,
		Encryption:  "none",
		FromAddress: "Info@mycompany.com",
		FromName:    "Info",
		ErrorChan:   errorChan,
		MailerChan:  mailerChan,
		DoneChan:    mailerDoneChan,
		Wait:        app.Wait,
	}

	return m
}
