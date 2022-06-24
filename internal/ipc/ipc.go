// DSUL - Disturb State USB Light : IPC module.
package ipc

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"github.com/hymnis/dsul-go/internal/settings"
	ipc "github.com/hymnis/golang-ipc"
)

var (
	verbose bool = false
	debug   bool = false
)

type Message struct {
	Type   string
	Key    string
	Value  string
	Secret string
}

// Runner parts //

// Start runner for IPC server.
func ServerRunner(cfg *settings.Config, output_handling struct {
	Verbose bool
	Debug   bool
}, cmd_channel chan string, rsp_channel chan string) {
	verbose = output_handling.Verbose
	debug = output_handling.Debug
	sc_config := &ipc.ServerConfig{
		Network: false,
	}

	if cfg.Network.Listen {
		sc_config = &ipc.ServerConfig{
			Network:     true,
			NetworkPort: cfg.Network.Port,
		}
	}
	sc, err := ipc.StartServer("dsul", sc_config)
	if err != nil {
		log.Println(err)
		return
	}

	out_channel := make(chan Message) // sent over IPC, in 'serverSend'
	go responseHandler(rsp_channel, out_channel)
	go serverReceive(cfg, sc, cmd_channel, out_channel)
	go serverSend(sc, out_channel)

	select {}
}

func serverSend(sc *ipc.Server, out_channel chan Message) {
	for {
		select {
		case message := <-out_channel:
			msg_str := encodeToBytes(message)
			msg_type := 2 // 2 == data
			_ = sc.Write(msg_type, msg_str)
			if debug {
				log.Printf("[ipc] Server Sent: %v\n", msg_str)
			}
			time.Sleep(time.Second / 30)
		}
	}
}

func serverReceive(cfg *settings.Config, sc *ipc.Server, cmd_channel chan string, out_channel chan Message) {
	for {
		m, err := sc.Read()

		if err == nil {
			if m.MsgType == 1 {
				if debug {
					log.Printf("[ipc] Server Received, auth: %v\n", m.Data)
				}
			}
			if m.MsgType == 2 {
				if debug {
					log.Printf("[ipc] Server Received, data: %v\n", m.Data)
				}
				cmd := decodeToMessage(m.Data)
				// Authentication if needed
				if cfg.Password != "" && cfg.Password != cmd.Secret {
					log.Printf("[ipc] Server Authentication failed\n")
				} else {
					if cmd.Type == "set" {
						// Send "set" message to cmd_channel (received by serial module)
						cmd_channel <- fmt.Sprintf("%s:%s", cmd.Key, cmd.Value)
					} else if cmd.Type == "get" {
						// Get and return information (to IPC client)
						if cmd.Key == "information" {
							if cmd.Value == "all" {
								// Request hardware state, returned via rsp_channel
								cmd_channel <- fmt.Sprintf("%s:%s", cmd.Key, cmd.Value)
							}
						}
					}
				}
			}
		} else {
			log.Println("[ipc] Error: " + err.Error())
			break
		}
	}
}

// Handles responses from serial device.
func responseHandler(rsp_channel chan string, out_channel chan Message) {
	for {
		select {
		case response := <-rsp_channel:
			out_channel <- Message{"set", "response", response, ""} // action, key, value, secret
		}
	}
}

func ClientRunner(cfg *settings.Config, output_handling struct {
	Verbose bool
	Debug   bool
}, ipc_message chan Message, ipc_response chan Message, done chan bool) {
	verbose = output_handling.Verbose
	debug = output_handling.Debug
	cc_config := &ipc.ClientConfig{
		Network: false,
		Timeout: 5,
	}

	if cfg.Network.Server != "" {
		cc_config = &ipc.ClientConfig{
			Network:       true,
			NetworkPort:   cfg.Network.Port,
			NetworkServer: cfg.Network.Server,
			Timeout:       5,
		}
	}

	cc, err := ipc.StartClient("dsul", cc_config)
	if err != nil {
		log.Println(err)
		return
	}

	ready := make(chan bool) // used to determine when ready to send in 'clientSend'
	go clientReceive(cc, ready, ipc_response)
	go clientSend(cc, ready, ipc_message, done)

	select {}
}

func clientSend(cc *ipc.Client, ready chan bool, ipc_message chan Message, done chan bool) {
	select {
	case <-ready:
		for {
			select {
			case message, more := <-ipc_message:
				if more { // channel is open and more data will come
					// Data
					msg_str := encodeToBytes(message)
					msg_type := 2 // 2 == data
					_ = cc.Write(msg_type, msg_str)
					if debug {
						log.Printf("[ipc] Client Sent, data: %v\n", msg_str)
					}
					time.Sleep(time.Second / 30)
				} else { // channel has been closed by client
					done <- true // send done message once all messages are handled and channel is closed
					return       // exit functions since we are all done
				}
			}
		}
	}
}

func clientReceive(cc *ipc.Client, ready chan bool, ipc_response chan Message) {
	for {
		m, err := cc.Read()

		if err != nil {
			// An error is only returned if the recieved channel has been closed,
			// so you know the connection has either been intentionally closed or has timmed out waiting to connect/re-connect.
			log.Fatal("[ipc] Error: ", err)
		}

		if m.MsgType == -1 { // message type -1 is status change
			if m.Status == "Connected" {
				ready <- true
			}
		}

		if m.MsgType == -2 { // message type -2 is an error, these won't automatically cause the recieve channel to close.
			log.Println("[ipc] Error: " + err.Error())
		}

		if m.MsgType == 1 { // message type 1 is authentication
			// ...
			if debug {
				log.Printf("[ipc] Client Received, auth: %v\n", m.Data)
			}
		}

		if m.MsgType == 2 { // message type 2 is data (messages)
			if debug {
				log.Printf("[ipc] Client Received, data: %v\n", m.Data)
			}
			response := decodeToMessage(m.Data)
			ipc_response <- response
		}
	}
}

func encodeToBytes(p interface{}) []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(p)
	if err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}

func decodeToMessage(input []byte) Message {
	cmd := Message{}
	dec := gob.NewDecoder(bytes.NewReader(input))
	err := dec.Decode(&cmd)
	if err != nil {
		log.Fatal(err)
	}
	return cmd
}

func decodeToString(input []byte) string {
	str := ""
	dec := gob.NewDecoder(bytes.NewReader(input))
	err := dec.Decode(&str)
	if err != nil {
		log.Fatal(err)
	}
	return str
}
