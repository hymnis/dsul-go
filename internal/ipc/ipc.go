// DSUL - Disturb State USB Light : IPC module.
package ipc

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"github.com/hymnis/dsul-go/internal/settings"
	ipc "github.com/james-barrow/golang-ipc"
)

var verbose bool = false
var unlocked bool = true

type Command struct {
	Action string
	Key    string
	Value  string
}

// Runner parts //

// Start runner for IPC server.
func ServerRunner(cfg *settings.Config, verbosity bool, cmd_channel chan string, rsp_channel chan string) {
	verbose = verbosity
	sc, err := ipc.StartServer("dsul", nil)
	if err != nil {
		log.Println(err)
		return
	}

	out_channel := make(chan Command) // sent over IPC, in 'serverSend'
	go responseHandler(rsp_channel, out_channel)
	go serverReceive(sc, cmd_channel, out_channel, cfg.Password)
	go serverSend(sc, out_channel)

	select {}
}

func serverSend(sc *ipc.Server, out_channel chan Command) {
	for {
		select {
		case command := <-out_channel:
			msg_str := encodeToBytes(command)
			msg_type := 2 // 2 == data
			_ = sc.Write(msg_type, msg_str)
			if verbose {
				log.Printf("[ipc] Server Sent: %v\n", msg_str)
			}
			time.Sleep(time.Second / 30)
		}
	}
}

func serverReceive(sc *ipc.Server, cmd_channel chan string, out_channel chan Command, password string) {
	for {
		m, err := sc.Read()

		if err == nil {
			if m.MsgType == 1 && password != "" {
				if verbose {
					log.Printf("[ipc] Server Received, auth: %v\n", m.Data)
				}
				in_password := decodeToString(m.Data)
				if in_password == password {
					if verbose {
						log.Println("[ipc] Authentication successful.")
					}
					unlocked = true
				} else {
					if verbose {
						log.Println("[ipc] Authentication failed.")
					}
					unlocked = false
				}
			}
			if m.MsgType == 2 && unlocked {
				if verbose {
					log.Printf("[ipc] Server Received, data: %v\n", m.Data)
				}
				cmd := decodeToCommand(m.Data)
				if cmd.Action == "set" {
					// Send "set" command to cmd_channel (received by serial module)
					cmd_channel <- fmt.Sprintf("%s:%s", cmd.Key, cmd.Value)
				} else if cmd.Action == "get" {
					// Get and return information (to IPC client)
					if cmd.Key == "information" {
						if cmd.Value == "all" {
							// Request hardware state, returned via rsp_channel
							cmd_channel <- fmt.Sprintf("%s:%s", cmd.Key, cmd.Value)
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
func responseHandler(rsp_channel chan string, out_channel chan Command) {
	for {
		select {
		case response := <-rsp_channel:
			out_channel <- Command{"set", "response", response} // action, key, value
		}
	}
}

func ClientRunner(cfg *settings.Config, verbosity bool, ipc_command chan Command, ipc_response chan Command, done chan bool) {
	verbose = verbosity
	conf := &ipc.ClientConfig{
		Timeout: 5,
	}
	cc, err := ipc.StartClient("dsul", conf)
	if err != nil {
		log.Println(err)
		return
	}

	ready := make(chan bool) // used to determine when ready to send in 'clientSend'
	go clientReceive(cc, ready, ipc_response)
	go clientSend(cc, ready, ipc_command, done, cfg.Password)

	select {}
}

func clientSend(cc *ipc.Client, ready chan bool, ipc_command chan Command, done chan bool, password string) {
	select {
	case <-ready:
		for {
			select {
			case command, more := <-ipc_command:
				if more { // channel is open and more data will come
					// Auth
					if password != "" {
						msg_str := encodeToBytes(password)
						msg_type := 1 // 1 == auth
						_ = cc.Write(msg_type, msg_str)
						if verbose {
							log.Printf("[ipc] Client Sent, auth: %v\n", msg_str)
						}
					}

					// Data
					msg_str := encodeToBytes(command)
					msg_type := 2 // 2 == data
					_ = cc.Write(msg_type, msg_str)
					if verbose {
						log.Printf("[ipc] Client Sent, data: %v\n", msg_str)
					}
					time.Sleep(time.Second / 30)
				} else { // channel has been closed by client
					done <- true // send done message once all commands are handled and channel is closed
					return       // exit functions since we are all done
				}
			}
		}
	}
}

func clientReceive(cc *ipc.Client, ready chan bool, ipc_response chan Command) {
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
			if verbose {
				log.Printf("[ipc] Client Received, auth: %v\n", m.Data)
			}
		}

		if m.MsgType == 2 { // message type 2 is data (commands)
			if verbose {
				log.Printf("[ipc] Client Received, data: %v\n", m.Data)
			}
			response := decodeToCommand(m.Data)
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

func decodeToCommand(input []byte) Command {
	cmd := Command{}
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
