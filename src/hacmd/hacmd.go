package hacmd

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"io/ioutil"
	"log"
	"lutron"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Helper Functions
//define a function for the default message handler
var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

// Creates a Unique id for the Command Center
func procUUID() (uuid string) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	uuid = fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return
}

// Main Functions
func readConfig(configstr string) (user string, pass string, broker string, procID string) {

	plan, _ := ioutil.ReadFile(configstr)
	var data map[string]interface{}
	err := json.Unmarshal(plan, &data)

	if err != nil {
		log.Fatalln(err)
	}
	user = data["user"].(string)
	pass = data["pass"].(string)
	broker = data["broker"].(string)
	if data["procID"] != nil {
		procID = data["procID"].(string)
	} else {
		procID = ""
	}
	return
}

type confighacmd struct {
	JobID      int             `json:"jobid"`
	Trigger    string          `json:"trigger"`
	ProcID     string          `json:"procid"`
	Action     string          `json:"action"`
	ActionType string          `json:"actiontype`
	Commands   []commandStruct `json:"commands"`
}
type commandStruct struct {
	URL       string `json:"url"`
	Hubid     string `json:"hubid"`
	VendorTag string `json:"vendortag"`
}
type hacmd struct {
	ProcID      string
	Configured  bool
	CmdMessages chan string
	MqttClient  mqtt.Client
}

func New(configStr string) hacmd {
	ch1 := make(chan string)
	// Read in HA Command config and get values
	// Connect MQTT Broker
	// Create a Process ID if one doesn't exist in the config file
	// Check-in with Command Process
	// Subscribe to Command Queue
	user, pass, broker, procID := readConfig(configStr)
	// Connect to MQTT broker
	mqttClient := connect_mqtt(broker, user, pass)
	// If procID does not existing in config.json then use ProcUUID and Write out new configuration file
	if len(procID) <= 0 {
		procID = procUUID()
		WriteConfig("config.json", user, pass, broker, procID)
	}

	// Check in to the control process
	checkInCtl(mqttClient, procID)

	// Subscribe to a topic
	go subscribeTo("hacmd/cmd", mqttClient, ch1)
	cmdProc := hacmd{procID, false, ch1, mqttClient}
	return cmdProc
}
func (cmdProc hacmd) ReadCommands(cmdStr string) (procid string, action string, ActionType string, commands []commandStruct) {
	res := confighacmd{}
	err := json.Unmarshal([]byte(cmdStr), &res)
	if err != nil {
		log.Fatalln(err)
	}
	procid = res.ProcID
	commands = res.Commands
	action = res.Action
	ActionType = res.ActionType
	return
}
func WriteConfig(configstr string, user string, pass string, broker string, procID string) {

	data := []byte("{\n\t\"user\": \"" + user + "\",\n\t\"pass\": \"" + pass + "\",\n\t\"broker\": \"" + broker + "\",\n\t\"procID\": \"" + procID + "\"\n}")
	_ = ioutil.WriteFile(configstr, data, 0644)

}
func (cmdProc hacmd) APICommand(procID string, hubID string, command string) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.Get(command)

	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	if json.Valid(body) {
		bodyStr := string(body)
		bodyStr = "{\"procid\":\"" + procID + "\",\"hubid\":\"" + hubID + "\",\"action\":\"result\",\"command\":\"" + command + "\",\"results\": " + bodyStr + "}"
		fmt.Println(time.Now().Format(time.RFC850) + " Pinged " + command + " procID: " + procID)
		publishTo("hacmd/ctrl", cmdProc.MqttClient, bodyStr)
	} else {
		fmt.Println("Command " + command + " did not respond with valid JSON.")
	}
	return
}

func (cmdProc hacmd) LutronCommand(command string) {
	params := strings.Split(command, "/")
	lutronHub := lutron.NewLutron(params[2], "/hacmd/inventory.txt")
	myCmd := &lutron.LutronMsg{}
	myCmd.Id, _ = strconv.Atoi(params[4])
	myCmd.Name = params[5]
	myCmd.Value, _ = strconv.ParseFloat(params[6], 64)
	myCmd.Fade, _ = strconv.ParseFloat(params[7], 64)
	if params[8] == "set" {
		myCmd.Type = lutron.Set
	}
	if params[3] == "output" {
		myCmd.Cmd = lutron.Output
	}
	myCmd.Action, _ = strconv.Atoi(params[9])
	lutronHub.Connect()
	lutronHub.SendCommand(myCmd)
	lutronHub.Disconnect()
}

func connect_mqtt(broker string, user string, pass string) (c mqtt.Client) {

	//create a ClientOptions struct setting the broker address, clientid, turn
	//off trace output and set the default message handler
	opts := mqtt.NewClientOptions().AddBroker("tcp://" + broker)
	opts.SetClientID(procUUID())
	opts.SetDefaultPublishHandler(f)
	opts.SetUsername(user)
	opts.SetPassword(pass)
	//create and start a client using the above ClientOptions
	c = mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return

}
func subscribeTo(topic string, client mqtt.Client, ch1 chan string) {
	fmt.Println("Subscribing to: " + topic)
	if token := client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		payloadStr := string(msg.Payload())
		//if strings.Contains(payloadStr, "configuration") {
		if len(payloadStr) > 0 {
			ch1 <- payloadStr
		}
		return

	}); token.Wait() && token.Error() != nil {
		return
	}
	return
}
func Unsubscribe(client mqtt.Client) {
	// Unscribe
	if token := client.Unsubscribe("hacmd/#"); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
}
func Disconnect(client mqtt.Client) {
	// Disconnect
	client.Disconnect(250)
	time.Sleep(1 * time.Second)
}
func checkInCtl(client mqtt.Client, procID string) {
	// Initial Check In Message
	msg := "{\"procID\": \"" + procID + "\",\"action\": \"initiate\"}"
	// Inital Check In a message
	publishTo("hacmd/ctrl", client, msg)
}
func publishTo(topic string, client mqtt.Client, msg string) {
	// Publish Response
	token := client.Publish(topic, 0, false, msg)
	token.Wait()
}
