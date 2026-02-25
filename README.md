# go-cloud-events-pubsub-nats

A simple Go example demonstrating the **Publish/Subscribe** messaging pattern with [NATS](https://nats.io).

## What is NATS?

[NATS](https://nats.io) is a lightweight, high-performance messaging system for cloud-native applications, IoT, and microservices. It uses a **subject-based** addressing model where:

- **Publishers** send messages to a named subject (topic).
- **Subscribers** listen on a subject and receive matching messages.
- Publishers and subscribers are fully **decoupled** — neither needs to know about the other.

## Prerequisites

- **Go 1.25+**
- **A running NATS server** —  pick your preferred method of [installing a NATS Server](https://docs.nats.io/running-a-nats-service/introduction/installation)

### with Docker

```bash
docker run -d --name nats-server -p 4222:4222 -p 8222:8222 nats:latest
```

###  from your terminal :

install nats-server via [command line](https://docs.nats.io/running-a-nats-service/introduction/installation#getting-the-binary-from-the-command-line) 
or your [package manager](https://docs.nats.io/running-a-nats-service/introduction/installation#installing-via-a-package-manager) 

then run it directly (in dev):
```bash
nats-server -js -DV -sd ./nats_data -m 8222
```
or with our helper script :
```bash
scripts/runNatsStreamServerDev.sh
```


| Port | Purpose                              |
|------|--------------------------------------|
| 4222 | Client connections (used by this app)|
| 8222 | HTTP monitoring dashboard            |

> **Tip:** You can verify the server is running by visiting [http://localhost:8222](http://localhost:8222) in your browser.

## How to Try

### 1. Clone & build

```bash
git clone https://github.com/lao-tseu-is-alive/go-cloud-events-pubsub-nats.git
cd go-cloud-events-pubsub-nats
go build -o nats-basic ./cmd/nats-basic/
```

### 2. Start a subscriber (Terminal 1)

The subscriber blocks and waits for messages — start it **first**:

```bash
./nats-basic -mode sub -subject "greetings"
```

### 3. Publish a message (Terminal 2)

Open a second terminal and publish:

```bash
./nats-basic -mode pub -subject "greetings" -msg "Hello NATS World!"
```
You should see the terminal print:

```
[pub] 2026/02/25 10:45:20 Connecting to NATS server at nats://127.0.0.1:4222 …
[pub] 2026/02/25 10:45:20 ✅ Connected to NATS server successfully.
[pub] 2026/02/25 10:45:20 Publishing to subject "greetings" …
[pub] 2026/02/25 10:45:20 ✅ Message published — subject: "greetings", payload: "Hello NATS World!"
```

You should see the subscriber terminal print:

```
[sub] 2026/02/25 10:45:12 Connecting to NATS server at nats://127.0.0.1:4222 …
[sub] 2026/02/25 10:45:12 ✅ Connected to NATS server successfully.
[sub] 2026/02/25 10:45:12 Subscribing to subject "greetings" — waiting for messages (Ctrl+C to quit) …
[sub] 2026/02/25 10:45:20 📩 Received on [greetings]: Hello NATS World!
```

### 4. Try wildcards

NATS supports two wildcard tokens in subject names:

| Token | Matches                        | Example                          |
|-------|--------------------------------|----------------------------------|
| `*`   | Exactly one token              | `sensor.*.temperature`           |
| `>`   | One or more tokens (tail match)| `events.>`                       |

```bash
# Subscribe to all events
./nats-basic -mode sub -subject "events.>"

# In another terminal, publish to different sub-subjects
./nats-basic -mode pub -subject "events.user.login"    -msg '{"user":"alice"}'
./nats-basic -mode pub -subject "events.order.created"  -msg '{"order":42}'
```

Both messages will be received by the single subscriber.

```
[sub] 2026/02/25 10:48:32 Connecting to NATS server at nats://127.0.0.1:4222 …
[sub] 2026/02/25 10:48:32 ✅ Connected to NATS server successfully.
[sub] 2026/02/25 10:48:32 Subscribing to subject "events.>" — waiting for messages (Ctrl+C to quit) …
[sub] 2026/02/25 10:49:15 📩 Received on [events.user.login]: {"user":"alice"}
[sub] 2026/02/25 10:49:29 📩 Received on [events.order.created]: {"order":42}
```

## CLI Reference

```
Usage of nats-basic:
  -mode string
        Operating mode: "pub" (publish) or "sub" (subscribe) — required
  -msg string
        Message payload to publish — required only in "pub" mode
  -subject string
        NATS subject (topic) to publish/subscribe to — required
  -url string
        NATS server URL (default "nats://127.0.0.1:4222")
```

## Project Structure

```
.
├── cmd/
│   └── nats-basic/
│       └── natsPubSub.go   # Main client — pub/sub with CLI flags
├── go.mod
├── go.sum
└── README.md
```

## Key Concepts Illustrated in the Code

| Concept              | Where / What                                                                 |
|----------------------|------------------------------------------------------------------------------|
| **Connect**          | `nats.Connect()` — establishes a TCP connection with auto-reconnect          |
| **Publish**          | `nc.Publish()` — fire-and-forget message send                                |
| **Flush**            | `nc.Flush()` — ensures buffered messages are sent before program exits       |
| **Subscribe**        | `nc.Subscribe()` — async callback invoked per message on a separate goroutine|
| **Drain**            | `nc.Drain()` — graceful shutdown: processes in-flight messages then closes   |
| **Graceful shutdown**| OS signal handling (`SIGINT`/`SIGTERM`) to stop the subscriber cleanly       |

## Going Further

- [NATS Documentation](https://docs.nats.io/)
- [NATS Go Client](https://github.com/nats-io/nats.go)
- **JetStream** — for persistent, at-least-once / exactly-once delivery: [JetStream docs](https://docs.nats.io/nats-concepts/jetstream)
- **NATS CLI** — a handy tool to inspect subjects, publish, subscribe, and manage streams: [nats-io/natscli](https://github.com/nats-io/natscli)

## License

[MIT](LICENSE)
