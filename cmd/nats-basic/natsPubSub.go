// natsPubSub.go — A simple NATS Pub/Sub demonstration client.
//
// PURPOSE:
//
//	This program demonstrates the core Publish/Subscribe messaging pattern
//	using NATS (https://nats.io). NATS is a lightweight, high-performance
//	messaging system designed for cloud-native applications, IoT, and
//	microservices architectures.
//
// PUB/SUB PATTERN:
//
//	In the Pub/Sub model, publishers send messages to a "subject" (a named
//	channel) without knowing who — if anyone — is listening. Subscribers
//	express interest in a subject and receive messages published to it.
//	This decouples producers from consumers, enabling scalable architectures.
//
// PREREQUISITES:
//
//  1. A running NATS server. The easiest way is with Docker:
//     docker run -d --name nats-server -p 4222:4222 -p 8222:8222 nats:latest
//     Port 4222 is the client port, 8222 is the HTTP monitoring port.
//
//  2. Or install nats-server natively: https://docs.nats.io/running-a-nats-service/introduction/installation
//
// USAGE:
//
//	Subscribe mode (start this first, it will block and wait for messages):
//	  go run natsPubSub.go -mode sub -subject "greetings"
//
//	Publish mode  (in another terminal):
//	  go run natsPubSub.go -mode pub -subject "greetings" -msg "Hello NATS World!"
//
//	You should see the subscriber terminal print the received message.
//
// NATS DEFAULT URL:
//
//	By default the client connects to nats://127.0.0.1:4222 (nats.DefaultURL).
//	You can override this with the -url flag if your server runs elsewhere.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
)

const (
	// modePub and modeSub are the two operating modes of this program.
	modePub = "pub"
	modeSub = "sub"
	APP     = "NATS-BASIC"
)

func main() {
	// ─── CLI Flag Definitions ──────────────────────────────────────────
	// flag.String returns a *string; we dereference them below after Parse().
	mode := flag.String("mode", "", `Operating mode: "pub" (publish) or "sub" (subscribe) — required`)
	subject := flag.String("subject", "", "NATS subject (topic) to publish/subscribe to — required")
	msg := flag.String("msg", "", `Message payload to publish — required only in "pub" mode`)
	natsURL := flag.String("url", nats.DefaultURL, "NATS server URL (default: nats://127.0.0.1:4222)")

	flag.Parse()

	// ─── Input Validation ──────────────────────────────────────────────
	if *mode == "" || *subject == "" {
		fmt.Fprintln(os.Stderr, "Error: -mode and -subject flags are required.")
		flag.Usage()
		os.Exit(1)
	}

	if *mode != modePub && *mode != modeSub {
		fmt.Fprintf(os.Stderr, "Error: -mode must be %q or %q, got %q.\n", modePub, modeSub, *mode)
		flag.Usage()
		os.Exit(1)
	}

	if *mode == modePub && *msg == "" {
		fmt.Fprintln(os.Stderr, `Error: -msg flag is required when using -mode "pub".`)
		flag.Usage()
		os.Exit(1)
	}

	// ─── Logger Setup ──────────────────────────────────────────────────
	// Prefix the log output with the mode so it's easy to distinguish
	// publisher vs subscriber output in your terminals.
	l := log.New(os.Stdout, fmt.Sprintf("%s [%s] ", APP, *mode), log.LstdFlags)

	// ─── Connect to NATS ───────────────────────────────────────────────
	// nats.Connect establishes a TCP connection to the NATS server.
	// It will automatically attempt to reconnect if the connection drops.
	// The returned *nats.Conn is safe for concurrent use.
	l.Printf("Connecting to NATS server at %s …", *natsURL)
	// Connections can be assigned a name which will appear in some of the server monitoring data
	// it is highly recommended as a friendly connection name will help in monitoring, error reporting, debugging, and testing.
	nc, err := nats.Connect(*natsURL, nats.Name(APP))
	if err != nil {
		l.Fatalf("💥 Failed to connect to NATS at %s: %v", *natsURL, err)
	}
	// Always close the connection when done to release resources.
	defer nc.Close()
	l.Println("✅ Connected to NATS server successfully.")

	// ─── Mode Dispatch ─────────────────────────────────────────────────
	switch *mode {
	case modePub:
		publish(nc, l, *subject, *msg)
	case modeSub:
		subscribe(nc, l, *subject)
	}
}

// publish sends a single message to the given NATS subject.
//
// KEY CONCEPT — Fire and Forget:
//
//	nc.Publish is asynchronous from the client's perspective: it buffers
//	the message and returns immediately. The message is flushed to the
//	server in the background. We call nc.Flush() explicitly here to
//	ensure the message has been sent before the program exits.
//
//	If you need delivery guarantees (at-least-once, exactly-once),
//	consider using NATS JetStream instead of core NATS Pub/Sub.
func publish(nc *nats.Conn, l *log.Logger, subject, msg string) {
	l.Printf("Publishing to subject %q …", subject)

	// Publish takes a subject and a byte slice payload.
	// NATS messages are opaque byte arrays — you can send JSON, Protobuf,
	// plain text, or any binary format.
	if err := nc.Publish(subject, []byte(msg)); err != nil {
		l.Fatalf("💥 Failed to publish: %v", err)
	}

	// Flush ensures all buffered messages are sent to the server.
	// Without this, the program might exit before the message is actually
	// transmitted over the network.
	if err := nc.Flush(); err != nil {
		l.Fatalf("💥 Failed to flush: %v", err)
	}

	l.Printf("✅ Message published — subject: %q, payload: %q", subject, msg)
}

// subscribe listens for messages on the given NATS subject.
//
// KEY CONCEPT — Async Subscription:
//
//	nc.Subscribe registers a callback that NATS invokes on a separate
//	goroutine each time a message arrives. This is the most common
//	subscription pattern.
//
//	The subscriber runs indefinitely until interrupted (Ctrl+C).
//
// WILDCARDS:
//
//	NATS supports two wildcard tokens in subjects:
//	  *  — matches a single token:   "sensor.*.temperature"
//	  >  — matches one or more tokens: "sensor.>"
//	Example: subscribing to "events.>" will receive messages published to
//	"events.user.login", "events.order.created", etc.
func subscribe(nc *nats.Conn, l *log.Logger, subject string) {
	l.Printf("Subscribing to subject %q — waiting for messages (Ctrl+C to quit) …", subject)

	// The callback function is invoked asynchronously for every message
	// that matches the subject. m.Data contains the raw payload bytes.
	sub, err := nc.Subscribe(subject, func(m *nats.Msg) {
		l.Printf("📩 Received on [%s]: %s", m.Subject, string(m.Data))
	})
	if err != nil {
		l.Fatalf("💥 Failed to subscribe: %v", err)
	}
	// Unsubscribe is called when the function exits to cleanly remove
	// the subscription from the server.
	defer func() {
		if err := sub.Unsubscribe(); err != nil {
			l.Printf("⚠️  Error during unsubscribe: %v", err)
		}
	}()

	// ─── Graceful Shutdown ─────────────────────────────────────────────
	// We block the main goroutine by waiting for an OS signal (SIGINT or
	// SIGTERM). Without this, the program would exit immediately after
	// subscribing, because Subscribe is non-blocking.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh // Block until signal is received

	l.Printf("🛑 Received signal %v — shutting down gracefully …", sig)

	// Drain ensures that all in-flight messages are processed before
	// the connection is closed.  This is the recommended shutdown
	// pattern for NATS subscribers.
	if err := nc.Drain(); err != nil {
		l.Printf("⚠️  Error during drain: %v", err)
	}
	l.Println("👋 Bye!")
}
