package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/SQUASHD/hbit"
	"github.com/SQUASHD/hbit/events"
	"github.com/SQUASHD/hbit/http"
	"github.com/SQUASHD/hbit/rpg"
	"github.com/SQUASHD/hbit/rpg/character"
	"github.com/SQUASHD/hbit/rpg/quest"
	"github.com/SQUASHD/hbit/rpg/rpgdb"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	connectionStr := os.Getenv("RPG_DB_URL")
	db, err := hbit.NewDatabase(hbit.NewDbParams{
		ConnectionStr: connectionStr,
		Driver:        hbit.DbDriverLibsql,
	})
	if err != nil {
		log.Fatalf("cannot connect to rpg database: %s", err)
	}

	err = hbit.DBMigrateUp(db, hbit.MigrationData{
		FS:      rpg.Migrations,
		Dialect: "sqlite",
		Dir:     "schemas",
	})
	if err != nil {
		log.Fatalf("failed to run migration of rpg database: %v", err)
	}
	rabbitmqUrl := os.Getenv("RABBITMQ_URL")
	rpgPublisher, rpgConn, err := events.NewPublisher(rabbitmqUrl)
	if err != nil {
		log.Fatalf("cannot create rpg publisher: %s", err)
	}
	defer rpgConn.Close()
	charPublisher, charConn, err := events.NewPublisher(rabbitmqUrl)
	if err != nil {
		log.Fatalf("cannot create rpg publisher: %s", err)
	}
	defer charConn.Close()

	queries := rpgdb.New(db)

	questSvc := quest.NewService(db, queries)
	charSvc := character.NewService(db, queries, charPublisher)

	rpgSvc := rpg.NewService(rpg.NewServiceParams{
		CharacterSvc: charSvc,
		QuestSvc:     questSvc,
		Publisher:    rpgPublisher,
		Queries:      queries,
		Db:           db,
	})

	consumer, conn, err := events.NewRPGEventConsumer(rabbitmqUrl)
	if err != nil {
		log.Fatalf("cannot create rpg consumer: %s", err)
	}
	defer conn.Close()

	rpgRouter := http.NewRPGRouter(charSvc, questSvc, rpgSvc)
	wrappedRouter := http.ChainMiddleware(
		rpgRouter,
	)
	server, err := http.NewServer(
		wrappedRouter,
		http.WithServerOptionsPortFromEnv("RPG_SVC_PORT"),
	)
	if err != nil {
		log.Fatalf("cannot create server: %s", err)
	}

	closed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		fmt.Println("\nShutting down server...")

		consumer.Close()

		ctx, cancel := context.WithTimeout(context.Background(), server.IdleTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown failure: %v", err)
		}

		if err := rpgSvc.CleanUp(); err != nil {
			log.Fatalf("RPG service cleanup failure: %v", err)
		}

		close(closed)
	}()
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := consumer.Run(
			events.RPGMessageHandler(rpgSvc),
		); err != nil {
			log.Fatalf("cannot start consuming: %s", err)
		}
	}()
	fmt.Printf("Server is running on port %s\n", server.Addr)
	if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("cannot start server: %s", err)
	}
	wg.Wait()

	<-closed
	log.Println("Server closed")
}
