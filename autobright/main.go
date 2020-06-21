package main
import (
mqtt "github.com/eclipse/paho.mqtt.golang"
	"net/url"
	"fmt"
	"time"
	"log"
	"strconv"
	"io/ioutil"
)

//MQTT server param
var mqtt_url *url.URL
var light_lvl_topic="ESP32_RoomSensor1_light_level" //Topic where light level can be readed
var light_lvl float64
var light_lvl_min=10.0  //Minimum light level fron your sensor (dark time value)
var light_lvl_max=500.0 //Maximum light level from your light sensor (bright sunny day value)
var brightness_ctl="/sys/class/backlight/radeon_bl1/brightness" //Path to brightness control file in your system
var brightness_min=60.0 //Low margin for calculated brightness
var brightness_max=150.0 //High margin for calculated brightness

var clientId="watcher_sub" //Client ID for MQTT. 
var mqttUrl="tcp://192.168.0.2:14419" //URL of MQTT server

func connect(clientId string, uri *url.URL) mqtt.Client {
	opts := createClientOptions(clientId, uri)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}
	return client
}

func createClientOptions(clientId string, uri *url.URL) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", uri.Host))
	opts.SetUsername(uri.User.Username())
	password, _ := uri.User.Password()
	opts.SetPassword(password)
	opts.SetClientID(clientId)
	opts.SetAutoReconnect(true)
	return opts
}


func watch_light_lvl(){
	client := connect(clientId, mqtt_url)
	client.Subscribe(light_lvl_topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		message:=string(msg.Payload())
		log.Printf("Got msg: %s",message)
		light_lvl,_=strconv.ParseFloat(message,32)
	})
}

func main(){
	mqtt_url, _=url.Parse(mqttUrl)
	go watch_light_lvl()
	for {
		log.Printf("Current light level: %f \n",light_lvl)
		time.Sleep(1*time.Second)

		//Calculate brightness
		lght:=light_lvl
		if(lght<light_lvl_min){
			lght=light_lvl_min
		}
		if(lght>light_lvl_max){
			lght=light_lvl_max
		}
		br_sp:=1-((light_lvl_max-lght)/(light_lvl_max-light_lvl_min))
		br_sp=brightness_min+((brightness_max-brightness_min)*br_sp)
		log.Printf("Calculated brightness: %d\n",int(br_sp))

		//Set brightness
		err := ioutil.WriteFile(brightness_ctl, []byte(strconv.Itoa(int(br_sp))), 0644)
		if (err!=nil){
			fmt.Printf("Error writing to file! %v",err)
		}
	}
}