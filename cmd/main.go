package main

import (
	"log"

	"github.com/DeepanshuMishraa/vid-processing-go.git/config"
	"github.com/DeepanshuMishraa/vid-processing-go.git/db"
	"github.com/DeepanshuMishraa/vid-processing-go.git/queue"
	"github.com/DeepanshuMishraa/vid-processing-go.git/utils"
	"github.com/DeepanshuMishraa/vid-processing-go.git/worker"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to Load Env Variables: %v", err)
	}
	log.Println("Loaded Env Variables")

	dbPool, err := db.Connect(cfg.DATABASE_URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	rmqConn, err := queue.ConnectRabbitMQ(cfg.RABBIT_MQ_URL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rmqConn.Close()

	r2Svc, err := utils.NewR2Service(cfg)
	if err != nil {
		log.Fatalf("Failed to create R2 service: %v", err)
	}

	go func() {
		if err := worker.Start(rmqConn, dbPool, r2Svc); err != nil {
			log.Fatalf("Worker exited: %v", err)
		}
	}()

	router := gin.Default()
	router.SetTrustedProxies(nil)

	router.GET("/health", gin.HandlerFunc(func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	}))

	router.Run(":" + cfg.PORT)
}
