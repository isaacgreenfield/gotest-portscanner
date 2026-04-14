package main

import (
	"flag"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
)

type ScanResult struct {
	Port   int
	Banner string
}

func worker(ports <-chan int, res chan<- ScanResult, wg *sync.WaitGroup, ip string) {
	defer wg.Done()

	for p := range ports {
		address := fmt.Sprintf("%s:%d", ip, p)
		conn, err := net.DialTimeout("tcp", address, 1*time.Second)

		if err != nil {
			continue
		}
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))

		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		
		if err != nil {
			conn.SetReadDeadline(time.Now().Add(2 * time.Second)) 
			conn.Write([]byte("GET / HTTP/1.0\r\n\r\n"))
			n, _ = conn.Read(buffer)
		}

		conn.Close()

		banner := string(buffer[:n])
		banner = strings.Split(banner, "\n")[0]
		banner = strings.TrimSpace(banner)

		res <- ScanResult{
			Port:   p,
			Banner: banner,
		}
	}
}

func main() {
	ip := flag.String("ip", "scanme.nmap.org", "IP-адрес или домен для сканирования")
	sP := flag.Int("start", 1, "Начальный порт для сканирования")
	eP := flag.Int("end", 1024, "Конечный порт для сканирования")
	numWrk := flag.Int("wrk", 100, "Количество одновременных потоков")

	flag.Parse()
	fmt.Printf("Запуск сканера: IP=%s, Порты: %d-%d, Воркеры: %d\n", *ip, *sP, *eP, *numWrk)
	fmt.Println("Сканирование началось, подождите...")

	ports := make(chan int, *numWrk)
	res := make(chan ScanResult)

	var wg sync.WaitGroup
	var openPorts []ScanResult
	
	for range *numWrk {
		wg.Add(1)
		go worker(ports, res, &wg, *ip)
	}

	go func() {
		for i := *sP; i <= *eP; i++ {
			ports <- i
		}
		close(ports)
	}()
	go func() {
		wg.Wait()
		close(res)
	}()
	
	for result := range res {
		openPorts = append(openPorts, result)
	}
	sort.Slice(openPorts, func(i, j int) bool {
		return openPorts[i].Port < openPorts[j].Port
	})

	fmt.Println("\n--- РЕЗУЛЬТАТЫ ---")
	if len(openPorts) == 0 {
		fmt.Printf("На %s нет открытых портов в диапазоне %d-%d\n", *ip, *sP, *eP)
	} else {
		fmt.Printf("Открытые порты на %s:\n", *ip)
		for _, result := range openPorts {
			if result.Banner != "" {
				fmt.Printf("✓ Порт %d открыт [Сервис: %s]\n", result.Port, result.Banner)
			} else {
				fmt.Printf("✓ Порт %d открыт [Сервис: нет ответа]\n", result.Port)
			}
		}
	}
}
