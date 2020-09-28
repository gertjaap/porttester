// Main loop (C) 2020 Gert-Jaap Glasbergen
// TCP Server Copyright https://opensource.com/article/18/5/building-concurrent-tcp-server-go
// UDP Server Copyright https://ops.tips/blog/udp-client-and-server-in-go/
package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())

	tcpPorts := strings.Split(os.Getenv("DUMMY_TCPPORTS"), "|")
	udpPorts := strings.Split(os.Getenv("DUMMY_UDPPORTS"), "|")

	for _, p := range tcpPorts {
		port, err := strconv.Atoi(p)
		if err != nil {
			panic(err)
		}
		go tcpServer(port)
	}

	for _, p := range udpPorts {
		port, err := strconv.Atoi(p)
		if err != nil {
			panic(err)
		}
		go udpServer(port)
	}

	select {}
}

// (C) CopyRight https://opensource.com/article/18/5/building-concurrent-tcp-server-go
func tcpServer(port int) {
	for {
		l, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
		if err != nil {
			fmt.Printf("[TCP] Could not start listening on port %d: %s\n", port, err.Error())
			return
		}
		defer l.Close()
		fmt.Printf("[TCP] Listening on port [%d]\n", port)

		for {
			c, err := l.Accept()
			if err != nil {
				fmt.Println(err)
				break
			}
			go func(c net.Conn) {
				fmt.Printf("[TCP] Connection from [%s]\n", c.RemoteAddr().String())
				for {
					netData, err := bufio.NewReader(c).ReadString('\n')
					if err != nil {
						fmt.Println(err)
						break
					}

					temp := strings.TrimSpace(string(netData))
					fmt.Printf("[TCP] Received from [%s] : [%s]\n", c.RemoteAddr().String(), temp)

					if temp == "STOP" {
						break
					}

					_, err = c.Write([]byte(string(netData)))
					if err != nil {
						fmt.Printf("[TCP] Error writing to [%s] : [%s]", c.RemoteAddr().String(), err.Error())
						break
					}
					fmt.Printf("[TCP] Echoed to [%s]\n", c.RemoteAddr().String())

				}
				c.Close()
				fmt.Printf("[TCP] Closed connection to %s\n", c.RemoteAddr().String())
			}(c)
		}
	}
}

// server wraps all the UDP echo server functionality.
// (C) Copyright https://ops.tips/blog/udp-client-and-server-in-go/
func udpServer(port int) (err error) {
	for {
		ctx := context.Background()
		// ListenPacket provides us a wrapper around ListenUDP so that
		// we don't need to call `net.ResolveUDPAddr` and then subsequentially
		// perform a `ListenUDP` with the UDP address.
		//
		// The returned value (PacketConn) is pretty much the same as the one
		// from ListenUDP (UDPConn) - the only difference is that `Packet*`
		// methods and interfaces are more broad, also covering `ip`.
		pc, errr := net.ListenPacket("udp", fmt.Sprintf(":%d", port))
		if errr != nil {
			err = errr
			return
		}
		fmt.Printf("[UDP] Listening on port [%d]\n", port)

		// `Close`ing the packet "connection" means cleaning the data structures
		// allocated for holding information about the listening socket.
		defer pc.Close()

		doneChan := make(chan error, 1)
		buffer := make([]byte, 1024)

		go func() {
			for {
				n, addr, err := pc.ReadFrom(buffer)
				if err != nil {
					doneChan <- err
					return
				}

				temp := strings.TrimSpace(string(buffer[:n]))

				fmt.Printf("[UDP] Received from [%s] : [%s]\n", addr.String(), temp)

				deadline := time.Now().Add(5 * time.Second)
				err = pc.SetWriteDeadline(deadline)
				if err != nil {
					doneChan <- err
					return
				}

				// Write the packet's contents back to the client.
				n, err = pc.WriteTo([]byte(fmt.Sprintf("%s\n", temp)), addr)
				if err != nil {
					fmt.Printf("[UDP] Failed to echo to [%s] : [%s]\n", addr.String(), err)
					continue
				}

				fmt.Printf("[UDP] Echoed to [%s]\n", addr.String())
			}
		}()

		select {
		case <-ctx.Done():
			fmt.Println("cancelled")
			err = ctx.Err()
		case err = <-doneChan:
		}
	}
	return
}
