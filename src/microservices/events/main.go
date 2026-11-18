package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

type EventResponse struct {
	Status    string `json:"status"`
	Partition int    `json:"partition"`
	Offset    int64  `json:"offset"`
	Event     Event  `json:"event"`
}

type Config struct {
	Port         string
	KafkaBrokers string
}

func loadConfig() Config {
	port := getenv("PORT", "8082")
	brokers := getenv("KAFKA_BROKERS", "localhost:9092")

	return Config{
		Port:         port,
		KafkaBrokers: brokers,
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

var kafkaWriter *kafka.Writer

func main() {
	config := loadConfig()

	kafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP(strings.Split(config.KafkaBrokers, ",")...),
		Balancer: &kafka.LeastBytes{},
	}
	defer kafkaWriter.Close()

	startConsumers(config.KafkaBrokers)

	//http.HandleFunc("/api/events/movie", movieEventHandler)
	//http.HandleFunc("/api/events/user", userEventHandler)
	//http.HandleFunc("/api/events/payment", paymentEventHandler)

	http.HandleFunc("/api/events/health", healthHandler)
	http.HandleFunc("/api/events/movie", movieEventHandler)
	http.HandleFunc("/api/events/user", userEventHandler)
	http.HandleFunc("/api/events/payment", paymentEventHandler)

	address := ":" + config.Port
	log.Printf("Starting events service on %s", address)
	log.Fatal(http.ListenAndServe(address, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"status": true})
}

func publishEvent(ctx context.Context, topic string, event Event) (kafka.Message, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return kafka.Message{}, err
	}

	message := kafka.Message{
		Topic: topic,
		Key:   []byte(event.ID),
		Value: payload,
		Time:  time.Now(),
	}

	err = kafkaWriter.WriteMessages(ctx, message)
	return message, err
}

func startConsumers(brokers string) {
	topics := []string{
		"movie-events",
		"user-events",
		"payment-events",
	}

	for _, topic := range topics {
		go consumeTopic(brokers, topic)
	}
}

func consumeTopic(brokers, topic string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  strings.Split(brokers, ","),
		Topic:    topic,
		GroupID:  "events-service-consumer",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	log.Printf("Kafka consumer started for topic %s", topic)

	for {
		message, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("failed to read message from topic %s: %v", topic, err)
			continue
		}

		var event Event
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("failed to decode message from topic %s: %v", topic, err)
			continue
		}

		log.Printf(
			"Event processed: topic=%s partition=%d offset=%d id=%s type=%s payload=%s",
			topic,
			message.Partition,
			message.Offset,
			event.ID,
			event.Type,
			string(event.Payload),
		)
	}
}

func movieEventHandler(w http.ResponseWriter, r *http.Request) {
	handleEvent(w, r, "movie-events", "movie")
}

func userEventHandler(w http.ResponseWriter, r *http.Request) {
	handleEvent(w, r, "user-events", "user")
}

func paymentEventHandler(w http.ResponseWriter, r *http.Request) {
	handleEvent(w, r, "payment-events", "payment")
}

func handleEvent(w http.ResponseWriter, r *http.Request, topic, eventType string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	event := Event{
		ID:        eventType + "-" + time.Now().UTC().Format("20060102150405.000000000"),
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Payload:   payload,
	}

	message, err := publishEvent(r.Context(), topic, event)
	if err != nil {
		log.Printf("failed to publish %s event: %v", eventType, err)
		http.Error(w, "failed to publish event", http.StatusInternalServerError)
		return
	}

	response := EventResponse{
		Status:    "success",
		Partition: message.Partition,
		Offset:    message.Offset,
		Event:     event,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}
