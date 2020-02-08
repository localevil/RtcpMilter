package main

import (
	"flag"
	"log"
	"net"
	"os"

	"github.com/phalaaxx/milter"
)

//RcptMilter mail filter delete rcpts limit overheder
type RcptMilter struct {
	milter.Milter
	rcptCount int
	rcpts     map[string]bool
}

//Connect init RcptMilters fields
func (r *RcptMilter) Connect(host string, family string, port uint16, addr net.IP, m *milter.Modifier) (milter.Response, error) {
	r.rcptCount = 0
	r.rcpts = make(map[string]bool)
	return milter.RespContinue, nil
}

//RcptTo save rcptTo to map end and mark it as deleteble or not
func (r *RcptMilter) RcptTo(rcptTo string, m *milter.Modifier) (milter.Response, error) {
	del, ok := r.rcpts[rcptTo]
	if !ok {
		r.rcptCount++
		if r.rcptCount >= 5 {
			del = true
		}
		r.rcpts[rcptTo] = del
	}
	return milter.RespContinue, nil
}

//Body delete marked in RcptTo function rcpts
func (r *RcptMilter) Body(m *milter.Modifier) (milter.Response, error) {
	for rcptTo, del := range r.rcpts {
		if del {
			log.Printf("Rcpt deleted: %s", rcptTo)
			m.DeleteRecipient(rcptTo)
		}
	}

	return milter.RespContinue, nil
}

//RunServer of mail filter
func RunServer(socket net.Listener) {
	// declare milter init function
	init := func() (milter.Milter, milter.OptAction, milter.OptProtocol) {
		return &RcptMilter{},
			milter.OptRemoveRcpt | milter.OptChangeBody,
			milter.OptNoHelo | milter.OptNoMailFrom | milter.OptNoHeaders | milter.OptNoEOH | milter.OptNoBody
	}
	// start server
	if err := milter.RunServer(socket, init); err != nil {
		log.Fatal(err)
	}
}

func main() {
	var protocol, address string

	flag.StringVar(&protocol,
		"proto",
		"",
		"Protocol family (unix or tcp)")
	flag.StringVar(&address,
		"port",
		"",
		"Bind to address or unix domain socket")

	flag.Parse()

	// make sure the specified protocol is either unix or tcp
	if protocol != "unix" && protocol != "tcp" {
		log.Fatal("invalid protocol name")
	}

	// make sure socket does not exist
	if protocol == "unix" {
		os.Remove(address)
	}

	log.Printf("Start RtcpMilter on %s", address)

	// open socket
	socket, err := net.Listen(protocol, address)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		socket.Close()
		log.Printf("Stop RtcpMilter")
	}()

	// check socket is unix
	if protocol == "unix" {
		if err := os.Chmod(address, 0660); err != nil {
			log.Fatal(err)
		}
		defer os.Remove(address)
	}

	// start listen socket by milter
	go RunServer(socket)

	select {}
}
