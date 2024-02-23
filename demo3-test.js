import { createLibp2p } from 'libp2p'
import { gossipsub } from '@chainsafe/libp2p-gossipsub'
import { tcp } from '@libp2p/tcp'
import { mplex } from '@libp2p/mplex'
import { yamux } from '@chainsafe/libp2p-yamux'
import { noise } from '@chainsafe/libp2p-noise'
import { mdns } from '@libp2p/mdns'
import { bootstrap } from '@libp2p/bootstrap'
import { fromString as uint8ArrayFromString } from 'uint8arrays/from-string'
import { toString as uint8ArrayToString } from 'uint8arrays/to-string'
import { createEd25519PeerId } from '@libp2p/peer-id-factory'
import { identifyService } from 'libp2p/identify'
import { kadDHT } from '@libp2p/kad-dht'
import util from 'util'
util.inspect.defaultOptions.depth = null

const topic = 'demo'
const totalNodes = 5

let nodeCount = 0
const nodes = []

const createNode = async () => {
    nodeCount++

    // peer ids end with the node count, helps reading logs
    let peerId
    while (!peerId?.toString().endsWith(nodeCount)) {
        peerId = await createEd25519PeerId()
    }

    const peerDiscovery = [
        // do peer discovery on local network
        mdns()
    ]
    // the first node is the bootstrap node
    if (nodeCount !== 1) {
        peerDiscovery.push(bootstrap({
            list: nodes[0].getMultiaddrs()
        }))
    }

    const node = await createLibp2p({
        addresses: {
            listen: ['/ip4/127.0.0.1/tcp/0']
        },
        peerDiscovery,
        transports: [tcp()],
        streamMuxers: [yamux(), mplex()],
        connectionEncryption: [noise()],
        // connectionGater: {
        //     denyDialMultiaddr: async () => false,
        // },
        services: {
            identify: identifyService(), // required for peer discovery of pubsub
            dht: kadDHT(), // p2p peer discovery
            pubsub: gossipsub({
                // doPX: nodeCount == 1 ? true : false,
                allowPublishToZeroPeers: true,
                // emitSelf: true,
                // directPeers: nodeCount !== 1 ? [{id: nodes[0].peerId, addrs: nodes[0].getMultiaddrs()}] : []
            })
        },
        peerId
    })

    logEvents(`node${nodeCount}${nodeCount === 1 ? '-bootstrap' : ''}`, node)
    return node
}

const logEvents = (nodeName, node) => {
    const events = [
        'connection:close',
        'connection:open',
        'connection:prune',
        // 'peer:connect',
        'peer:disconnect',
        'peer:discovery',
        'peer:identify',
        'peer:update',
        // 'self:peer:update',
        'start',
        'stop',
        'transport:close',
        'transport:listening',
    ]
    const format = (obj) => {
        const formatted = {}
        for (const i in obj) {
            if (typeof obj[i] === 'function') {
                continue
            }
            formatted[i] = obj[i]
        }
        return formatted
    }
    const logEvent = (event) => console.log(nodeName, event.type, format(event.detail))
    events.forEach(event => node.addEventListener(event, logEvent))
}

while(nodeCount < totalNodes) {
    const node = await createNode()
    const nodeName = `node${nodeCount}${nodeCount === 1 ? '-bootstrap' : ''}`
    console.log(nodeName, node.getMultiaddrs())

    // don't let the boostrap node join pubsub, to test peer discovery
    if (nodeCount !== 1) {
        // sub
        node.services.pubsub.addEventListener('message', (evt) => {
          console.log(`${nodeName} received: ${uint8ArrayToString(evt.detail.data)} on topic ${evt.detail.topic}`)
        })
        await node.services.pubsub.subscribe(topic)

        // pub
        setInterval(() => {
          node.services.pubsub.publish(topic, uint8ArrayFromString(`demo message from ${nodeName}`)).catch(err => {
            console.error(err)
          })
        }, 5000 * Math.random())  
    }

    nodes.push(node)
}
