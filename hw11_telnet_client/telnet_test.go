package main

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")
			err = client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})
}

func TestTelnetClient_ConnectTimeout(t *testing.T) {
	client := NewTelnetClient("127.0.0.2:12345", 500*time.Millisecond, nil, nil)
	start := time.Now()
	err := client.Connect()
	require.Error(t, err)
	require.WithinDuration(t, time.Now(), start.Add(500*time.Millisecond), time.Second)
}

func TestTelnetClient_ReceiveTimeout_NoServerData(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	defer ln.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := ln.Accept()
		require.NoError(t, err)
		defer conn.Close()
		time.Sleep(3 * time.Second)
	}()

	clientIn := &bytes.Buffer{}
	clientOut := &bytes.Buffer{}
	client := NewTelnetClient(ln.Addr().String(), 500*time.Millisecond, io.NopCloser(clientIn), clientOut)
	require.NoError(t, client.Connect())
	defer client.Close()

	done := make(chan struct{})
	go func() {
		_ = client.Receive()
		close(done)
	}()

	select {
	case <-done:
		t.Error("That should not be returned")
	case <-time.After(2 * time.Second):
	}
	wg.Wait()
}
