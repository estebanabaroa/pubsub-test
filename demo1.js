import { createLibp2p } from 'libp2p'
import { gossipsub } from '@chainsafe/libp2p-gossipsub'
import { tcp } from '@libp2p/tcp'
import { webSockets } from '@libp2p/websockets'
import { mplex } from '@libp2p/mplex'
import { yamux } from '@chainsafe/libp2p-yamux'
import { noise } from '@chainsafe/libp2p-noise'
import { fromString as uint8ArrayFromString } from 'uint8arrays/from-string'
import { toString as uint8ArrayToString } from 'uint8arrays/to-string'

const createNode1 = async () => {
    const node = await createLibp2p({
        addresses: {
            listen: ['/ip4/127.0.0.1/tcp/0']
        },
        transports: [tcp()],
        streamMuxers: [yamux(), mplex()],
        connectionEncryption: [noise()],
        services: {
            // we add the Pubsub module we want
            pubsub: gossipsub({
                allowPublishToZeroPeers: true,
                awaitRpcMessageHandler: true,
            })
        }
    })
    return node
}
const node1 = await createNode1()

const createNode2 = async () => {
    const node = await createLibp2p({
        addresses: {
            listen: ['/ip4/127.0.0.1/tcp/0']
        },
        transports: [tcp()],
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
    return node
}
const node2 = await createNode2()

// log addresses
console.log(node1.getMultiaddrs(), node2.getMultiaddrs())

const topic = 'news'

// sub
node1.services.pubsub.addEventListener('message', (evt) => {
  console.log(`node1 received: ${uint8ArrayToString(evt.detail.data)} on topic ${evt.detail.topic}`)
})
node2.services.pubsub.addEventListener('message', (evt) => {
  console.log(`node2 received: ${uint8ArrayToString(evt.detail.data)} on topic ${evt.detail.topic}`)
})
await node1.services.pubsub.subscribe(topic)
await node2.services.pubsub.subscribe(topic)

// pub
setInterval(() => {
  node2.services.pubsub.publish(topic, uint8ArrayFromString('Bird bird bird, bird is the word!')).catch(err => {
    console.error(err)
  })
}, 1000)
setInterval(() => {
  node1.services.pubsub.publish(topic, uint8ArrayFromString('Bear bear bear, bear is the word!')).catch(err => {
    console.error(err)
  })
}, 2000)
