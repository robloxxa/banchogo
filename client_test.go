package banchogo

import (
	"os"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func initBanchoClient() *Client {
	username := os.Getenv("IRC_USERNAME")
	password := os.Getenv("IRC_PASSWORD")
	apiKey := os.Getenv("API_KEY")

	if username == "" || password == "" {
		panic("Tests require env values IRC_USERNAME and IRC_PASSWORD to work")
	}

	return NewBanchoClient(ClientOptions{
		Username: username,
		Password: password,
		ApiKey:   apiKey,
	})
}

func initBanchoClientWithConnect() *Client {
	b := initBanchoClient()

	err := b.Connect()
	if err != nil {
		panic(err)
	}
	return b
}

func TestClient_Connect(t *testing.T) {
	c := initBanchoClient()
	done := make(chan struct{})
	c.OnceConnect(func() {
		close(done)
	})

	err := c.Connect()
	defer c.Disconnect()
	if err != nil {
		t.Fatalf("unexpected error: %e", err)
	}
	select {
	case <-done:
		break
	case <-time.After(10 * time.Second):
		t.Fatal("connect event doesn't occur")
	}
}

func TestClient_ConnectWithWrongCreds(t *testing.T) {
	c := initBanchoClient()

	c.Username = "123"
	c.Password = "321"

	err := c.Connect()
	defer c.Disconnect()
	if err == nil {
		t.Error("client doesn't return error when credentials are wrong")
	}
}

func TestClient_Disconnect(t *testing.T) {
	numGoroutines := runtime.NumGoroutine()
	c := initBanchoClient()
	done := make(chan struct{})
	c.OnceDisconnect(func(error) {
		close(done)
	})

	err := c.Connect()
	if err != nil {
		t.Error(err)
	}
	c.Disconnect()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Error("disconnect event doesn't occur")
	}
	<-time.After(2 * time.Second)
	// Check if some goroutines doesn't exit
	if numGoroutines != runtime.NumGoroutine() {
		t.Error("possible goroutine leak")
	}
}

func TestClient_Reconnect(t *testing.T) {
	b := initBanchoClientWithConnect()
	defer b.Disconnect()
	b.OnConnectState(func(state ConnectState) {
		if state == Reconnecting {
			t.Log("Reconnecting!")
		}
	})
	b.OnRawMessage(func(s []string) {
		t.Log(s)
	})
	time.Sleep(3 * time.Second)
	done := make(chan struct{})
	err := b.conn.Close()
	if err != nil {
		t.Log("could not close connection", err)
	}

	b.OnceConnect(func() {
		close(done)
	})

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Error("reconnection doesn't work")
	}

}

func TestClient_Ratelimit(t *testing.T) {
	b := initBanchoClientWithConnect()

	d := b.OnDisconnect(func(error) {
		t.Fatal("Ratelimiter failed")
	})
	for i := 0; i < 20; i++ {
		err := b.GetSelf().SendMessage(strconv.Itoa(i))
		if err != nil {
			t.Fatal(err)
		}
	}
	d()
	b.Disconnect()
}
