package mqtt

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

type Online struct {
	Number int      `json:"num"`
	Device []string `json:"device"`
}

var onlineChan = make(chan Online)
var errChan = make(chan error)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Topic: %s | %s\n", msg.Topic(), msg.Payload())
	var res Online
	err := json.Unmarshal(msg.Payload(), &res)
	if err != nil {
		errChan <- err
		return
	}
	onlineChan <- res
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
}

func Connect() mqtt.Client {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	var broker = os.Getenv("MQTT_BROKER")
	var port = os.Getenv("MQTT_PORT")
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%s", broker, port))
	opts.SetClientID("go_mqtt_client")
	opts.SetUsername(os.Getenv("MQTT_USERNAME"))
	opts.SetPassword(os.Getenv("MQTT_PASSWORD"))
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	return client
}

func Publish(client mqtt.Client, msg interface{}, topic string) {
	jsonData, err := json.Marshal(msg)
	fmt.Println(jsonData)
	if err != nil {
		fmt.Printf("JSON marshaling failed: %s\n", err)
	}
	token := client.Publish(topic, 0, false, jsonData)
	token.Wait()
}

func Sub(client mqtt.Client, topic string) {
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Println("Subscribed to LWT", topic)
}

func CheckOnline(client mqtt.Client, topicSub, topicPub string) (Online, error) {
	type Message struct {
		Request int `json:"request"`
	}

	msg := Message{
		Request: 12,
	}

	go Sub(client, topicSub)
	Publish(client, msg, topicPub)

	select {
	case res := <-onlineChan:
		return res, nil
	case err := <-errChan:
		return Online{}, err
	case <-time.After(10 * time.Second):
		return Online{}, fmt.Errorf("timeout waiting for response")
	}
}
