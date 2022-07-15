package stream

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	"math/rand"
	"time"
)

const (
	StreamName          = "vehicle_data"
	LocationSubject     = "location"
	BatteryLevelSubject = "battery_level"
)

type (
	EventFields struct {
		Longitude    float64 `json:"longitude,omitempty"`
		Latitude     float64 `json:"latitude,omitempty"`
		BatteryLevel int     `json:"batteryLevel,omitempty"`
	}

	Event struct {
		MessageType string      `json:"msgType"`
		Timestamp   int64       `json:"timestamp"`
		Fields      EventFields `json:"fields"`
	}
)

func CreateStreamIfNotExists(streamName string, jc nats.JetStreamContext) error {
	_, err := jc.StreamInfo(streamName)
	if err != nil && !errors.Is(err, nats.ErrStreamNotFound) {
		return fmt.Errorf("failed to check if stream exist: %w", err)
	}

	if errors.Is(err, nats.ErrStreamNotFound) {
		_, err = jc.AddStream(&nats.StreamConfig{
			Name:   streamName,
			MaxAge: 0,
			Subjects: []string{
				streamName + ".*",
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create stream: %w", err)
		}
	}

	return err
}

func PublishMessages(jc nats.JetStreamContext) error {
	var (
		ev    Event
		bytes []byte
		err   error
	)

	rand.Seed(time.Now().UnixMilli())
	for i := 0; i < 10; i++ {
		ev = Event{
			MessageType: LocationSubject,
			Timestamp:   time.Now().UnixNano(),
			Fields: EventFields{
				Longitude: float64(rand.Int63n(360) - 180),
				Latitude:  float64(rand.Int63n(180) - 90),
			},
		}
		if bytes, err = json.Marshal(&ev); err != nil {
			fmt.Println("Error: " + err.Error())
			continue
		}
		if _, err = jc.PublishAsync(StreamName+"."+LocationSubject, bytes); err != nil {
			fmt.Println("Error: " + err.Error())
			continue
		}

		ev = Event{
			MessageType: BatteryLevelSubject,
			Timestamp:   time.Now().UnixNano(),
			Fields: EventFields{
				BatteryLevel: int(rand.Int31n(101)),
			},
		}
		if bytes, err = json.Marshal(&ev); err != nil {
			fmt.Println("Error: " + err.Error())
			continue
		}
		if _, err = jc.PublishAsync(StreamName+"."+BatteryLevelSubject, bytes); err != nil {
			fmt.Println("Error: " + err.Error())
			continue
		}

		fmt.Printf("Published batch %d\n", i+1)
	}

	return nil
}

func GetLocationsSubscription(jc nats.JetStreamContext) (*nats.Subscription, error) {
	sub, err := jc.Subscribe(StreamName+"."+LocationSubject,
		handleLocations,
		nats.MaxDeliver(3),
		nats.MaxAckPending(5),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create location subscription: %w", err)
	}

	return sub, nil
}

func handleLocations(msg *nats.Msg) {
	var ev Event

	err := json.Unmarshal(msg.Data, &ev)
	if err != nil {
		fmt.Printf(">> Consumer: Location >> Error: %s\n", err.Error())
		msg.Nak()
		return
	}

	fmt.Printf(">> Consumer: Location >> Values:\n\tLongitude: %f\n\tLatitude: %f\n",
		ev.Fields.Longitude,
		ev.Fields.Latitude)

	msg.Ack()
}
