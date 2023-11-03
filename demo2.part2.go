package main

import (
    "bytes"
    "bufio"
    "context"
    "flag"
    "fmt"
    "log"
    "os"

    "github.com/libp2p/go-libp2p"
    "github.com/libp2p/go-libp2p/core/crypto"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    pubsub "github.com/libp2p/go-libp2p-pubsub"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    port := flag.Int("port", 0, "port number of libp2p host")
    peerString := flag.String("peer", "", "peer multiaddress to connect to")
    privateKeyString := flag.String("private-key", "", "private key of the libp2p identity")
    flag.Parse()

    if len([]byte(*privateKeyString)) < 32 {
        fmt.Println("--private-key argument too short")
        os.Exit(0)
    }

    // create libp2p
    h, err := makeHost(*port, *privateKeyString)
    if err != nil {
        log.Println(err)
        return
    }

    log.Printf("run '--peer %v/p2p/%s' in another terminal to connect to this peer\n", h.Network().ListenAddresses()[1], h.ID())
    log.Println("waiting for incoming connection")
    log.Println()

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
    topic, err := ps.Join("demo")
    if err != nil {
        panic(err)
    }
    go streamConsoleTo(ctx, topic)

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
        log.Println(err)
        return nil, err
    }

    return libp2p.New(
        libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1/webtransport", port)),
        libp2p.Identity(privateKey),
    )
}

func streamConsoleTo(ctx context.Context, topic *pubsub.Topic) {
    reader := bufio.NewReader(os.Stdin)
    for {
        s, err := reader.ReadString('\n')
        if err != nil {
            panic(err)
        }
        if err := topic.Publish(ctx, []byte(s)); err != nil {
            fmt.Println("### Publish error:", err)
        }
    }
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
