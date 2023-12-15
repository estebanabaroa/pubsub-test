package main

import (
    "bytes"
    "context"
    "flag"
    "fmt"
    "os"
    "time"
    "strings"
    "crypto/rand"

    "github.com/libp2p/go-libp2p"
    "github.com/libp2p/go-libp2p/core/crypto"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    pubsub "github.com/libp2p/go-libp2p-pubsub"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    privateKeyBytes := make([]byte, 32)
    rand.Read(privateKeyBytes)

    port := flag.Int("port", 0, "port number of libp2p host")
    topicString := flag.String("topic", "demo", "pubsub topic to join")
    peerString := flag.String("peer", "", "peer multiaddress to connect to")
    privateKeyString := flag.String("private-key", string(privateKeyBytes), "private key of the libp2p identity")
    flag.Parse()

    if len([]byte(*privateKeyString)) < 32 {
        fmt.Println("--private-key argument too short")
        os.Exit(0)
    }

    // create libp2p
    h, err := makeHost(*port, *privateKeyString)
    if err != nil {
        fmt.Println(err)
        return
    }

    // print how to connect to our peer
    listenAddresses := h.Network().ListenAddresses()
    for i := 0; i < len(listenAddresses); i++ {
        fmt.Println(listenAddresses[i])
    }
    fmt.Println(h.ID())
    for i := 0; i < len(listenAddresses); i++ {
        listenAddress := fmt.Sprintf("%v", listenAddresses[i])
        if strings.Contains(listenAddress, "webtransport") && strings.Contains(listenAddress, "ip4") {
            fmt.Printf("\nrun '--peer %v/p2p/%s' in another terminal to connect to this peer\n\n", listenAddresses[i], h.ID())
        }
    }

    // peer to connect to
    directPeers := []peer.AddrInfo{}
    if *peerString != "" {
        directPeer, err := peer.AddrInfoFromString(*peerString)
        if err != nil {
            panic(err)
        }
        directPeers = []peer.AddrInfo{*directPeer}
    }

    // create pubsub
    ps, err := pubsub.NewGossipSub(
        ctx, 
        h,
        pubsub.WithDirectPeers(directPeers),
    )
    if err != nil {
        panic(err)
    }

    // join pubsub
    topic, err := ps.Join(*topicString)
    if err != nil {
        panic(err)
    }

    // publish messages every 5 second
    go func() {
        for {
            // sleep 5 second
            time.Sleep(5 * time.Second)

            if err := topic.Publish(ctx, []byte("hello from libp2p go")); err != nil {
                fmt.Println("### Publish error:", err)
            }
        }
    }()

    // sub to topic and print all messages
    sub, err := topic.Subscribe()
    if err != nil {
        panic(err)
    }
    printMessagesFrom(ctx, sub)

    // wait forever
    select {}
}

func makeHost(port int, privateKeyString string) (host.Host, error) {
    privateKeyBytes := []byte(privateKeyString)
    privateKey, _, err := crypto.GenerateEd25519Key(bytes.NewReader(privateKeyBytes))
    if err != nil {
        fmt.Println(err)
        return nil, err
    }

    return libp2p.New(
        libp2p.ListenAddrStrings(
            fmt.Sprintf("/ip4/127.0.0.1/udp/%d/quic-v1/webtransport", port),
        ),
        libp2p.Identity(privateKey),
    )
}

func printMessagesFrom(ctx context.Context, sub *pubsub.Subscription) {
    for {
        m, err := sub.Next(ctx)
        if err != nil {
            panic(err)
        }
        fmt.Println(m.ReceivedFrom, ": ", string(m.Message.Data))
    }
}
