import path from 'path'
import {spawn, spawnSync} from 'child_process'
import ps from 'node:process'
// import {path as kuboPath} from 'kubo'
import {fileURLToPath} from 'url'
const rootPath = path.dirname(fileURLToPath(import.meta.url))
const kuboPath = () => path.join(rootPath, 'kubo')
const ipfsDataPath = path.join(rootPath, '.ipfs')
const env = {IPFS_PATH: ipfsDataPath}

const bootstrap = '/ip4/127.0.0.1/tcp/37185/p2p/12D3KooWJRP1bWPJYeSusEvVcvRs2ynprT61xaW47bJ9W5H15Wr1'

// use this custom function instead of spawnSync for better logging
// also spawnSync might have been causing crash on start on windows
const spawnAsync = (...args) =>
  new Promise((resolve, reject) => {
    const spawedProcess = spawn(...args)
    spawedProcess.on('exit', (exitCode, signal) => {
      if (exitCode === 0) resolve()
      else reject(Error(`spawnAsync process '${spawedProcess.pid}' exited with code '${exitCode}' signal '${signal}'`))
    })
    spawedProcess.stderr.on('data', (data) => console.error(data.toString()))
    spawedProcess.stdin.on('data', (data) => console.log(data.toString()))
    spawedProcess.stdout.on('data', (data) => console.log(data.toString()))
    spawedProcess.on('error', (data) => console.error(data.toString()))
  })

const kubo = {}

kubo.sync = (...args) => {
  console.log('kubo', ...args)
  return spawnSync(kuboPath(), args, {env, hideWindows: true}).stdout.toString()
}

kubo.async = (...args) => {
  console.log('kubo', ...args)
  return spawnAsync(kuboPath(), args, {env, hideWindows: true})
}

kubo.start = async () => {
  // init kubo client on first launch
  try {
    await spawnAsync(kuboPath(), ['init'], {env, hideWindows: true})
  } catch (e) {}

  await spawnAsync(kuboPath(), ['config', '--json', 'Bootstrap', `["${bootstrap}"]`], {env, hideWindows: true})

  await new Promise((resolve, reject) => {
    const kuboProcess = spawn(kuboPath(), ['daemon', '--migrate', '--enable-namesys-pubsub', '--enable-pubsub-experiment'], {env, hideWindows: true})
    console.log(`kubo daemon process started with pid ${kuboProcess.pid}`)
    let lastError
    kuboProcess.stderr.on('data', (data) => {
      lastError = data.toString()
      console.error(data.toString())
    })
    kuboProcess.stdin.on('data', (data) => console.log(data.toString()))
    kuboProcess.stdout.on('data', (data) => console.log(data.toString()))
    kuboProcess.on('error', (data) => console.error(data.toString()))
    kuboProcess.on('exit', () => {
      console.error(`kubo process with pid ${kuboProcess.pid} exited`)
      reject(Error(lastError))
    })
    process.on('exit', () => {
      try {
        ps.kill(kuboProcess.pid)
      } catch (e) {
        console.log(e)
      }
      try {
        // sometimes kubo doesnt exit unless we kill pid +1
        ps.kill(kuboProcess.pid + 1)
      } catch (e) {
        console.log(e)
      }
    })

    kuboProcess.stdout.on('data', (data) => {
      if (data.toString().match('Daemon is ready')) {
        resolve()
      }
    })
  })
}

export default kubo