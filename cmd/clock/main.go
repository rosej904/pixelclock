package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rosej904/pixelclock/config"
	"github.com/rosej904/pixelclock/internal/publisher"
)

func main() {
	cfg := config.Load()

	log.Printf("🕐 PixelClock Publisher starting")
	log.Printf("   Broker   : %s", cfg.BrokerURL)
	log.Printf("   Prefix   : %s", cfg.DevicePrefix)
	log.Printf("   Tick     : %ds", cfg.TickSeconds)

	pub, err := publisher.New(
		cfg.BrokerURL,
		cfg.ClientID,
		cfg.Username,
		cfg.Password,
		cfg.DevicePrefix,
	)
	if err != nil {
		log.Fatalf("❌ Failed to connect to broker: %v", err)
	}
	defer pub.Disconnect()

	// Clock style — tweak these to your liking.
	style := publisher.ClockStyle{
		TwelveHour:  false,
		ShowSeconds: true,
		Color:       []int{0, 200, 255}, // cyan digits
		Background:  []int{0, 0, 0},     // black background
		Rainbow:     false,
	}

	// Publish once immediately so the clock appears right away.
	publishClock(pub, style)

	ticker := time.NewTicker(time.Duration(cfg.TickSeconds) * time.Second)
	defer ticker.Stop()

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Println("⏱  Clock is running. Press Ctrl+C to stop.")

	for {
		select {
		case <-ticker.C:
			publishClock(pub, style)
		case sig := <-quit:
			log.Printf("📴 Received %s — shutting down", sig)
			return
		}
	}
}

func publishClock(pub *publisher.Publisher, style publisher.ClockStyle) {
	now := time.Now()
	payload := publisher.BuildClockPayload(now, style)

	if err := pub.PublishCustomApp("clock", payload); err != nil {
		log.Printf("⚠️  Publish error: %v", err)
	} else {
		log.Printf("📡 %s", publisher.ClockText(now, style))
	}
}
