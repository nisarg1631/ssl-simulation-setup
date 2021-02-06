package proxy

import (
	"log"
	"net"
	"sync"
	"time"
)

type MulticastServer struct {
	connection     *net.UDPConn
	running        bool
	consumer       func([]byte)
	mutex          sync.Mutex
	SkipInterfaces []string
}

func NewMulticastServer(consumer func([]byte)) (r *MulticastServer) {
	r = new(MulticastServer)
	r.consumer = consumer
	return
}

func (r *MulticastServer) Start(multicastAddress string) {
	r.running = true
	go r.receive(multicastAddress)
}

func (r *MulticastServer) Stop() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.running = false
	if err := r.connection.Close(); err != nil {
		log.Println("Could not close connection: ", err)
	}
}

func (r *MulticastServer) receive(multicastAddress string) {
	var currentIfiIdx = 0
	for r.isRunning() {
		ifis, _ := net.Interfaces()
		currentIfiIdx = (currentIfiIdx + 1) % len(ifis)
		ifi := ifis[currentIfiIdx]
		r.receiveOnInterface(multicastAddress, ifi)
		time.Sleep(1 * time.Second)
	}
}

func (r *MulticastServer) isRunning() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.running
}

func (r *MulticastServer) skipInterface(ifiName string) bool {
	for _, skipIfi := range r.SkipInterfaces {
		if skipIfi == ifiName {
			return true
		}
	}
	return false
}

func (r *MulticastServer) receiveOnInterface(multicastAddress string, ifi net.Interface) {
	addr, err := net.ResolveUDPAddr("udp", multicastAddress)
	if err != nil {
		log.Printf("Could resolve multicast address %v: %v", multicastAddress, err)
		return
	}

	r.connection, err = net.ListenMulticastUDP("udp", &ifi, addr)
	if err != nil {
		log.Printf("Could not listen at %v: %v", multicastAddress, err)
		return
	}

	if err := r.connection.SetReadBuffer(maxDatagramSize); err != nil {
		log.Println("Could not set read buffer: ", err)
	}

	log.Printf("Listening on %s (%s)", multicastAddress, ifi.Name)

	first := true
	data := make([]byte, maxDatagramSize)
	for {
		if err := r.connection.SetDeadline(time.Now().Add(300 * time.Millisecond)); err != nil {
			log.Println("Could not set deadline on connection: ", err)
		}
		n, _, err := r.connection.ReadFromUDP(data)
		if err != nil {
			log.Println("ReadFromUDP failed:", err)
			break
		}

		if first {
			log.Printf("Got first data packets from %s (%s)", multicastAddress, ifi.Name)
			first = false
		}

		r.consumer(data[:n])
	}

	log.Printf("Stop listening on %s (%s)", multicastAddress, ifi.Name)

	if err := r.connection.Close(); err != nil {
		log.Println("Could not close listener: ", err)
	}
}