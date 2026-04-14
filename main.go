package main

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

func wrk(ports <-chan int, res chan<- int, wg *sync.WaitGroup, ip string) {
	defer wg.Done()

	for p := range ports {
		address := fmt.Sprintf("%s:%d", ip, p)
		conn, err := net.DialTimeout("tcp", address, 1*time.Second)
		
		if err != nil {
			continue
		}
		conn.Close()
		res <- p
	}
}

func main() {
	ip := "scanme.nmap.org"
	numwrks := 100

	ports := make(chan int, numwrks)
	res := make(chan int)

	var wg sync.WaitGroup
	var openports []int

	for range in cap(ports) {
		wg.Add(1)
		go wrk(ports, res, &wg, ip)
	}

	go func() {
		for i := 1; i <= 1024; i++ {
			ports <- i
		}
		close(ports)
	}()
	go func() {
		wg.Wait()
		close(res)
	}()

	for port := range res {
		openports = append(openports, port)
	}

	sort.Ints(openports)
	fmt.Println("Открытые порты на", ip, ":")
	for _, port := range openports {
		fmt.Printf("- %d открыт\n", port)
	}
}
