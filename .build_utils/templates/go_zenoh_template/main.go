package main

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"

	zenoh "github.com/eclipse-zenoh/zenoh-go/zenoh"
	msgpack "github.com/hashicorp/go-msgpack/codec"
)

type TaggedString = struct {
	ID      int
	Message string
}

func dataHandler(sample zenoh.Sample) {

	dec := sample.Payload().Bytes()

	decoded := TaggedString{}

	err := msgpack.NewDecoder(bytes.NewReader(dec), nil).Decode(&decoded)
	if err != nil {
		fmt.Printf("failed to decode payload: %v\n", err)
		return
	}

	fmt.Printf(">> [Subscriber] Received ('%s': ID: '%s', Message: '%s')",

		sample.KeyExpr().String(),
		decoded.ID,
		decoded.Message,
	)

	// check if attachment exists
	if sample.Attachement().IsSome() {
		fmt.Printf(" (%s)", sample.Attachement().Unwrap().String())
	}
	fmt.Print("\n")
}
func cfgFromEnv() (zenoh.Config, error) {
	configPath := os.Getenv("ZENOH_CONFIG")

	var cfg zenoh.Config
	var err error

	if configPath != "" {
		cfg, err = zenoh.NewConfigFromFile(configPath)
	} else {
		cfg = zenoh.NewConfigDefault()
	}

	return cfg, err
}

func main() {

	cfg, err := cfgFromEnv()
	if err != nil {
		cfg = zenoh.NewConfigDefault()
	}

	session, err := zenoh.Open(cfg, nil)

	if err != nil {
		fmt.Printf("failed to open zenoh session: %v\n", err)
		os.Exit(1)
	}

	defer session.Drop()

	keyexpr, err := zenoh.NewKeyExpr("demo/out/*")
	sub, err := session.DeclareSubscriber(keyexpr, zenoh.Closure[zenoh.Sample]{Call: dataHandler}, nil)

	if err != nil {
		fmt.Printf("Unable to declare subscriber for key expression '%s': %v\n", keyexpr, err)
		os.Exit(-1)
	}

	defer sub.Drop()

	keyexpr, err = zenoh.NewKeyExpr("demo/out/go")
	pub, err := session.DeclarePublisher(keyexpr, nil)
	if err != nil {
		fmt.Printf("Unable to declare publisher for key expression '%s': %v\n", keyexpr, err)
		os.Exit(-1)
	}
	defer pub.Drop()

	payload := []interface{}{99, "hello from go!"}

	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf, &msgpack.MsgpackHandle{})
	if err := enc.Encode(payload); err != nil {
		panic(err)
	}

	pub.Put(zenoh.NewZBytes(buf.Bytes()), nil)
	fmt.Printf("Sent: TaggedString{id: %d, s: %q}\n", payload[0], payload[1])

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press CTRL-C to quit...")
	<-stop

}
