// DSUL - Disturb State USB Light : IPC module.
package ipc

import (
	"fmt"
	"log"
	"time"

	ipc "github.com/james-barrow/golang-ipc"
)

// Runner parts //

// Start runner for IPC server.
func ServerRunner() {
	//&ipc.ServerConfig{Encryption: false}

	sc, err := ipc.StartServer("testtest", nil)
	if err != nil {
		log.Println(err)
		return
	}

	go func() {
		for {
			m, err := sc.Read()

			if err == nil {
				if m.MsgType > 0 {
					log.Println("Server recieved: "+string(m.Data)+" - Message type: ", m.MsgType)
				}
			} else {
				log.Println("Server error")
				log.Println(err)
				break
			}
		}
	}()

	go serverSend(sc)
}

func serverSend(sc *ipc.Server) {
	for {
		err := sc.Write(3, []byte("Hello Client 4"))
		err = sc.Write(23, []byte("Hello Client 5"))
		err = sc.Write(65, []byte("Hello Client 6"))

		if err != nil {
			fmt.Println(err)
		}

		time.Sleep(time.Second / 30)
	}
}

func client() {
	//config := &ipc.ClientConfig{Encryption: false}

	cc, err := ipc.StartClient("testtest", nil)
	if err != nil {
		log.Println(err)
		return
	}

	go func() {
		for {
			m, err := cc.Read()

			if err != nil {
				// An error is only returned if the recieved channel has been closed,
				//so you know the connection has either been intentionally closed or has timmed out waiting to connect/re-connect.
				break
			}

			if m.MsgType == -1 { // message type -1 is status change
				//log.Println("Status: " + m.Status)
			}

			if m.MsgType == -2 { // message type -2 is an error, these won't automatically cause the recieve channel to close.
				log.Println("Error: " + err.Error())
			}

			if m.MsgType > 0 { // all message types above 0 have been recieved over the connection
				log.Println(" Message type: ", m.MsgType)
				log.Println("Client recieved: " + string(m.Data))
			}
		}
	}()

	go clientSend(cc)
}

func clientSend(cc *ipc.Client) {
	for {
		_ = cc.Write(14, []byte("hello server 4"))
		_ = cc.Write(44, []byte("hello server 5"))
		_ = cc.Write(88, []byte("hello server 6"))

		time.Sleep(time.Second / 20)
	}

}

func clientRecv(c *ipc.Client) {
	for {
		m, err := c.Read()

		if err != nil {
			// An error is only returned if the recieved channel has been closed,
			//so you know the connection has either been intentionally closed or has timmed out waiting to connect/re-connect.
			break
		}

		if m.MsgType == -1 { // message type -1 is status change
			//log.Println("Status: " + m.Status)
		}

		if m.MsgType == -2 { // message type -2 is an error, these won't automatically cause the recieve channel to close.
			log.Println("Error: " + err.Error())
		}

		if m.MsgType > 0 { // all message types above 0 have been recieved over the connection
			log.Println(" Message type: ", m.MsgType)
			log.Println("Client recieved: " + string(m.Data))
		}
	}
}
