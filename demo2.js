import { createLibp2p } from 'libp2p'
import { peerIdFromString } from '@libp2p/peer-id'
import { multiaddr } from '@multiformats/multiaddr'
import { gossipsub } from '@chainsafe/libp2p-gossipsub'
import { webTransport } from '@libp2p/webtransport'
import { mplex } from '@libp2p/mplex'
import { yamux } from '@chainsafe/libp2p-yamux'
import { noise } from '@chainsafe/libp2p-noise'
import { fromString as uint8ArrayFromString } from 'uint8arrays/from-string'
import { toString as uint8ArrayToString } from 'uint8arrays/to-string'

const log = (...args) => {
    console.log(...args)
    try {
        const p = document.createElement('p')
        let textContent = ''
        for (const [i, arg] of args.entries()) {
            if (textContent !== '') {
                textContent += ' '
            }
            if (typeof arg === 'object') {
                textContent += JSON.stringify(arg, null, 2)
                if (args.length - 1 !== i) {
                    textContent += '\n'
                }
            }
            else {
                textContent += arg
            }
        }
        p.textContent = textContent
        p.style.whiteSpace = 'pre'
        document.body.appendChild(p)
    }
    catch (e) {
        console.log(e)
    }
}

const logEvents = (nodeName, node) => {
    const events = [
        'connection:close',
        'connection:open',
        'connection:prune',
        'peer:connect',
        'peer:disconnect',
        'peer:discovery',
        'peer:identify',
        'peer:update',
        'self:peer:update',
        'start',
        'stop',
        'transport:close',
        'transport:listening',
    ]
    const logEvent = (event) => log(nodeName, event.type, event.detail)
    events.forEach(event => node.addEventListener(event, logEvent))
}

let peerId, peerMultiAddress
try {
    // example /ip4/0.0.0.0/udp/36338/quic-v1/webtransport/certhash/uEiAKYNN-nS53kXTQTzM0__ksuMI9mEczZA9YbNyyth-NLw/certhash/uEiD9Ja-UEHkuLjtaVKX7D1wZVri6GKkrmxFOue8-9YMAzg/p2p/Qmbi4zsa9H8WDZk8akwSiBqEe8U1Gedh8Mh4R4aCfR3Twx
    const split = location.search.replace(/^\?/, '').split('/p2p/')
    peerId = split[1]
    peerMultiAddress = split[0]
}
catch (e) {
    log('failed to get peer from query string')
    log(e)
}

;(async () => {

// node1 is created in go because listening on webTransport is not implemented in libp2p js
const node1 = {
    peerId: peerIdFromString(peerId),
    getMultiaddrs: () => [multiaddr(peerMultiAddress)]
}

const createNode2 = async () => {
    const node = await createLibp2p({
        // can't listen using webtransport in libp2p js
        // addresses: {listen: []},
        transports: [webTransport()],
        streamMuxers: [yamux(), mplex()],
        connectionEncryption: [noise()],
        services: {
            // we add the Pubsub module we want
            pubsub: gossipsub({
                allowPublishToZeroPeers: true,
                directPeers: [{id: node1.peerId, addrs: node1.getMultiaddrs()}]
            })
        }
    })
    logEvents('node2', node)
    return node
}
const node2 = await createNode2()

// log addresses
log('node1', node1.getMultiaddrs(), 'node2', node2.getMultiaddrs())

const topic = 'demo'

// sub
node2.services.pubsub.addEventListener('message', (evt) => {
    log(`node2 received: ${uint8ArrayToString(evt.detail.data)} on topic ${evt.detail.topic}`)
})
await node2.services.pubsub.subscribe(topic)

// pub
setInterval(() => {
  node2.services.pubsub.publish(topic, uint8ArrayFromString('demo message from node 2')).catch(err => {
    console.error(err)
  })
}, 1000)

})()
