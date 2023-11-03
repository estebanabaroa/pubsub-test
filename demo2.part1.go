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
    "github.com/libp2p/go-libp2p/core/network"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/core/peerstore"

    "github.com/multiformats/go-multiaddr"
)

func handleStream(s network.Stream) {
    log.Println("Got a new stream!")

    // Create a buffer stream for non-blocking read and write.
    rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

    go readData(rw)
    go writeData(rw)

    // stream 's' will stay open until you close it (or the other side closes it).
}

func readData(rw *bufio.ReadWriter) {
    for {
        str, _ := rw.ReadString('\n')

        if str == "" {
            return
        }
        if str != "\n" {
            // Green console colour:    \x1b[32m
            // Reset console colour:    \x1b[0m
            fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
        }

    }
}

func writeData(rw *bufio.ReadWriter) {
    stdReader := bufio.NewReader(os.Stdin)

    for {
        fmt.Print("> ")
        sendData, err := stdReader.ReadString('\n')
        if err != nil {
            log.Println(err)
            return
        }

        rw.WriteString(fmt.Sprintf("%s\n", sendData))
        rw.Flush()
    }
}

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

    h, err := makeHost(*port, *privateKeyString)
    if err != nil {
        log.Println(err)
        return
    }

    if *peerString == "" {
        startPeer(ctx, h, handleStream)
    } else {
        startPeerAndConnect(ctx, h, *peerString)
    }

    // Wait forever
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

func startPeer(ctx context.Context, h host.Host, streamHandler network.StreamHandler) {
    // Set a function as stream handler.
    // This function is called when a peer connects, and starts a stream with this protocol.
    // Only applies on the receiving side.
    h.SetStreamHandler("/chat/1.0.0", streamHandler)

    log.Printf("run '--peer %v/p2p/%s' on another terminal\n", h.Network().ListenAddresses()[1], h.ID())
    log.Println("waiting for incoming connection")
    log.Println()
}

func startPeerAndConnect(ctx context.Context, h host.Host, destination string) {
    log.Println("This node's multiaddresses:")
    for _, la := range h.Addrs() {
        log.Printf(" - %v\n", la)
    }
    log.Println()

    // Turn the destination into a multiaddr.
    maddr, err := multiaddr.NewMultiaddr(destination)
    if err != nil {
        log.Println(err)
        panic(err)
    }

    // Extract the peer ID from the multiaddr.
    info, err := peer.AddrInfoFromP2pAddr(maddr)
    if err != nil {
        log.Println(err)
        panic(err)
    }

    // Add the destination's peer multiaddress in the peerstore.
    // This will be used during connection and stream creation by libp2p.
    h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

    // Start a stream with the destination.
    // Multiaddress of the destination peer is fetched from the peerstore using 'peerId'.
    s, err := h.NewStream(context.Background(), info.ID, "/chat/1.0.0")
    if err != nil {
        panic(err)
    }
    log.Println("Established connection to destination")

    // Create a buffered stream so that read and writes are non-blocking.
    rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
    go writeData(rw)
    go readData(rw)
}
