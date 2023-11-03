// import util from 'util'
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

// util.inspect.defaultOptions.depth = null

;(async () => {

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
    const logEvent = (event) => console.log(nodeName, event.type, event.detail)
    events.forEach(event => node.addEventListener(event, logEvent))
}

// node1 is created in go because listening on webTransport is not implemented in libp2p js
const node1 = {
    peerId: peerIdFromString('12D3KooWMbbPVGsZYh3ChQjue712NHHGNybRRXwnuSpezYjGbCDS'),
    getMultiaddrs: () => [multiaddr('/ip4/0.0.0.0/udp/9999/quic-v1/webtransport/certhash/uEiBHtBgYlLSKhAzfyP8WQs9SbCVLWmYx3YZ_TeAuOB2Tyg/certhash/uEiC1wrlGK_ZlqKv1Xh2FTmrjQCAU5SKvoKfzTabitWYSPQ')]
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
console.log(node1.getMultiaddrs(), node2.getMultiaddrs())

const topic = 'demo'

// sub
node2.services.pubsub.addEventListener('message', (evt) => {
  console.log(`node2 received: ${uint8ArrayToString(evt.detail.data)} on topic ${evt.detail.topic}`)
})
await node2.services.pubsub.subscribe(topic)

// pub
setInterval(() => {
  node2.services.pubsub.publish(topic, uint8ArrayFromString('demo message from node 2')).catch(err => {
    console.error(err)
  })
}, 1000)

})()
