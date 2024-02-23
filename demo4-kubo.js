import kubo from './kubo.js'
import * as IpfsHttpClient from 'ipfs-http-client'

await kubo.start()

const ipfsClient = IpfsHttpClient.create({url: 'http://localhost:5001/api/v0'})
const topic = 'demo'

const onMessageReceived = (message) => {
  console.log(`received message from ${message.from}:`, new TextDecoder().decode(message.data))
}

await ipfsClient.pubsub.subscribe(topic, onMessageReceived)

setInterval(async () => {
  const message = new Date().toISOString()
  try {
    await ipfsClient.pubsub.publish(topic, Buffer.from(message))
    console.log('published message:', message)
  }
  catch (e) {
    console.log(e)
  }
}, 1000)
