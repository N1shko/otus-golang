package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type newClient struct {
	address    string
	timeout    time.Duration
	in         io.ReadCloser
	out        io.Writer
	connection net.Conn
}

func (client *newClient) Connect() error {
	connection, err := net.DialTimeout("tcp", client.address, client.timeout)
	if err != nil {
		return err
	}
	client.connection = connection
	log.Printf("Established connection to %s", client.address)
	return nil
}

func (client *newClient) Close() error {
	if err := client.connection.Close(); err != nil {
		return err
	}
	return nil
}

func (client *newClient) Send() error {
	scanner := bufio.NewReader(client.in)
	buf := make([]byte, 1024)

	for {
		n, err := scanner.Read(buf)
		switch {
		case errors.Is(err, io.EOF):
			log.Println("Detected EOF from input")
			return nil
		case err != nil:
			return fmt.Errorf("Send read error: %w", err)
		}

		if _, err := client.connection.Write(buf[:n]); err != nil {
			return fmt.Errorf("Send write error: %w", err)
		}
	}
}

func (client *newClient) Receive() error {
	buf := make([]byte, 1024)
	for {
		client.connection.SetReadDeadline(time.Now().Add(client.timeout))
		n, err := client.connection.Read(buf)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				continue
			}
			if errors.Is(err, io.EOF) {
				log.Print("Connection was closed by peer")
				return nil
			}
			log.Printf("Receive error: %v", err)
			return err
		}
		if n > 0 {
			_, err = client.out.Write(buf[:n])
			if err != nil {
				return (err)
			}
		}
	}
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &newClient{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}
