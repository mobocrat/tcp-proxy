package main

import (
	"log"
	"net"
	"sync"
)

func main() {
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		panic(err)
	}
	for {
		client, err := listener.Accept()
		if err != nil {
			log.Println(err)
		} else {
			go handle(client)
		}
	}
}

func handle(client net.Conn) {
	defer client.Close()
	upstream, err := net.Dial("tcp", "1.1.1.1:53")
	if err != nil {
		log.Println(err)
	}
	defer upstream.Close()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		buffer := make([]byte, 65536)
		for {
			n, err := client.Read(buffer)
			if err != nil {
				if err, ok := err.(net.Error); ok && err.Timeout() {
					continue
				}
				log.Println(err)
				break
			}
			log.Println("client: ", string(buffer[:n]))
			if _, err := upstream.Write(buffer[:n]); err != nil {
				log.Println(err)
				break
			}
		}
		wg.Done()
	}()
	go func() {
		buffer := make([]byte, 65536)
		for {
			n, err := upstream.Read(buffer)
			if err != nil {
				if err, ok := err.(net.Error); ok && err.Timeout() {
					continue
				}
				log.Println(err)
				break
			}
			log.Println("server: ", string(buffer[:n]))
			if _, err := client.Write(buffer[:n]); err != nil {
				log.Println(err)
				break
			}
		}
		wg.Done()
	}()
	wg.Wait()
}
