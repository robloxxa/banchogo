package main

import (
	"os"
	"testing"
)

func newBanchoClientFromEnv(t *testing.T) *BanchoClient {
	t.Helper()
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")

	return NewBanchoClient(&Config{
		Username: username,
		Password: password,
		Host:     host,
		Port:     port,
	})
}

func TestNewBanchoClient(t *testing.T) {
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	b := NewBanchoClient(&Config{
		Username: username,
		Password: password,
		Host:     host,
		Port:     port,
	})
	if (host == "" && b.Options.Host != "irc.ppy.sh") || (port == "" && b.Options.Port != "6667") {
		t.Error("BanchoClient constructor doesn't initialize standard host and port if not presented")
	}
}

func TestBanchoClient_Connect(t *testing.T) {
	b := newBanchoClientFromEnv(t)
	{
		u, p := b.Options.Username, b.Options.Password
		b.Options.Username = ""
		err := b.Connect()
		if err == nil {
			t.Error("Client trying to connect with empty username")
		}

		b.Options.Username = u
		b.Options.Password = ""

		err = b.Connect()
		if err == nil {
			t.Error("Client trying to connect with empty password")
		}

		b.Options.Username = ""

		err = b.Connect()
		if err == nil {
			t.Error("Client trying to connect with empty username and password")
		}
		b.Options.Username, b.Options.Password = u, p
	}
	t.Logf("%s", b.Options)
	b.OnMessage = func(message string) {
		t.Logf("%s", message)
	}
	b.Connect()
}
