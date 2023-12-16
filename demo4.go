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
    dht "github.com/libp2p/go-libp2p-kad-dht"
    "github.com/libp2p/go-libp2p/p2p/discovery/routing"
    discovery "github.com/libp2p/go-libp2p/p2p/discovery/util"
    "github.com/libp2p/go-libp2p/core/network"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    privateKeyBytes := make([]byte, 32)
    rand.Read(privateKeyBytes)

    port := flag.Int("port", 0, "port number of libp2p host")
    topicString := flag.String("topic", "demo", "pubsub topic to join")
    privateKeyString := flag.String("private-key", string(privateKeyBytes), "private key of the libp2p identity")
    flag.Parse()

    if len([]byte(*privateKeyString)) < 32 {
        fmt.Println("--private-key argument too short")
        os.Exit(0)
    }

    // create libp2p peer used as bootstrap
    h, err := makeBootstrapPeer(ctx, *port, *privateKeyString)
    if err != nil {
        fmt.Println(err)
        return
    }

    // print how to connect to our bootstrap peer
    listenAddresses := h.Network().ListenAddresses()
    bootstrapMultiaddressString := ""
    for i := 0; i < len(listenAddresses); i++ {
        fmt.Println(listenAddresses[i])
    }
    fmt.Println(h.ID())
    for i := 0; i < len(listenAddresses); i++ {
        listenAddress := fmt.Sprintf("%v", listenAddresses[i])
        if strings.Contains(listenAddress, "webtransport") && strings.Contains(listenAddress, "ip4") {
            fmt.Printf("\nrun '--peer %v/p2p/%s' in another terminal to connect to this peer\n\n", listenAddresses[i], h.ID())
            bootstrapMultiaddressString = fmt.Sprintf("%v/p2p/%s", listenAddresses[i], h.ID())
        }
    }

    // create pubsub peers and connect them to boostrap peer
    go makePubsubPeer(ctx, bootstrapMultiaddressString, topicString)
    go makePubsubPeer(ctx, bootstrapMultiaddressString, topicString)
    go makePubsubPeer(ctx, bootstrapMultiaddressString, topicString)

    // wait forever
    select {}
}

func makeBootstrapPeer(ctx context.Context, port int, privateKeyString string) (host.Host, error) {
    privateKeyBytes := []byte(privateKeyString)
    privateKey, _, err := crypto.GenerateEd25519Key(bytes.NewReader(privateKeyBytes))
    if err != nil {
        fmt.Println(err)
        return nil, err
    }

    // the bootstrap peer doesn't sub to pubsub to test peer discovery
    h, err := libp2p.New(
        libp2p.ListenAddrStrings(
            fmt.Sprintf("/ip4/127.0.0.1/udp/%d/quic-v1/webtransport", port),
        ),
        libp2p.Identity(privateKey),

        // options that maybe should be enabled
        libp2p.EnableHolePunching(), // disabled by default
        libp2p.EnableNATService(), // not sure if disabled by default
        libp2p.EnableRelayService(), // // not sure if disabled by default
    )
    if err != nil {
        fmt.Println(err)
        return nil, err
    }

    // create dht peer discovery
    kdht, err := dht.New(
        ctx, 
        h, 
        dht.ProtocolPrefix("/plebbit/lan"),
        // dht.Mode(dht.ModeServer),
    )
    if err != nil {
        return nil, err
    }
    rendezvous := "plebbit"
    go peerDiscovery(ctx, h, kdht, rendezvous)

    return h, nil
}

func makePubsubPeer(ctx context.Context, bootstrapMultiaddressString string, topicString *string) {
    h, err := libp2p.New(
        // options that maybe should be enabled
        libp2p.EnableHolePunching(), // disabled by default
        libp2p.EnableNATService(), // not sure if disabled by default
        libp2p.EnableRelayService(), // // not sure if disabled by default
    )
    if err != nil {
        fmt.Println(err)
        return
    }

    // connect to bootstrap peer
    directPeer, err := peer.AddrInfoFromString(bootstrapMultiaddressString)
    if err != nil {
        panic(err)
    }
    h.Connect(ctx, *directPeer)

    // create dht peer discovery
    kdht, err := dht.New(
        ctx, 
        h, 
        dht.ProtocolPrefix("/plebbit/lan"),
        // dht.Mode(dht.ModeServer),
    )
    if err != nil {
        fmt.Println(err)
    }
    rendezvous := "plebbit"
    go peerDiscovery(ctx, h, kdht, rendezvous)

    // create pubsub
    ps, err := pubsub.NewGossipSub(
        ctx, 
        h,
        pubsub.WithDirectPeers([]peer.AddrInfo{*directPeer}),
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

            message := fmt.Sprintf("hello from libp2p go %s", h.ID())
            if err := topic.Publish(ctx, []byte(message)); err != nil {
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

func peerDiscovery(ctx context.Context, h host.Host, dht *dht.IpfsDHT, rendezvous string) {
    routingDiscovery := routing.NewRoutingDiscovery(dht)

    discovery.Advertise(ctx, routingDiscovery, rendezvous)

    ticker := time.NewTicker(time.Second * 10)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:

            peers, err := discovery.FindPeers(ctx, routingDiscovery, rendezvous)
            if err != nil {
                panic(err)
            }

            for _, p := range peers {
                if p.ID == h.ID() {
                    continue
                }
                if h.Network().Connectedness(p.ID) != network.Connected {
                    _, err = h.Network().DialPeer(ctx, p.ID)
                    if err != nil {
                        fmt.Printf("Failed to connect to peer (%s): %s", p.ID, err.Error())
                        continue
                    }
                    fmt.Printf("Connected to peer %s", p.ID.Pretty())
                }
            }
        }
    }
}
