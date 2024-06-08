package mqtt

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
	"log"
	"os"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Topic: %s | %s\n", msg.Topic(), msg.Payload())
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

//func sub(client mqtt.Client) {
//	topic := "ibme/device/data/D001"
//	token := client.Subscribe(topic, 1, nil)
//	token.Wait()
//	fmt.Println("Subscribed to LWT", topic)
//}
