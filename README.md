# demo1: gossipsub in nodejs using tcp transport

```
node demo1.js
```

# demo2: gossipsub in go and chrome using webtransport (nodejs doesn't support webtransport yet)

```
npm run webpack:watch
```
```
go run demo2.go
```
- open `demo2.html` in chrome (no webtransport in firefox/safari yet) and add `?<multiaddress>` to the URL, e.g. `demo2.html?/ip4/127.0.0.1/udp/56916/quic-v1/webtransport/certhash/uEiBA6BEvjGlHJb7ZhLKazj6VFl03pX7kbDN4SU2gFM5yuw/certhash/uEiCeZhADwPPyWJe4Ert9nlpXyDv9RF4b_0DdCIqHAQdvQQ/p2p/12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq387JEYKDBj4kx6nXTN`
> NOTE: can set a port, private key and peer to connect to by using `go run demo2.go --private-key aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa --port 9999 --peer /ip4/127.0.0.1/udp/9999/quic-v1/webtransport/certhash/uEiC1wrlGK_ZlqKv1Xh2FTmrjQCAU5SKvoKfzTabitWYSPQ/certhash/uEiDw8Q_dufIwOTz8aaCm9ACPXeKbf_l8le-GqtdWrbY-rA/p2p/12D3KooWMbbPVGsZYh3ChQjue712NHHGNybRRXwnuSpezYjGbCDS`
> NOTE: can debug libp2p connection issues in the browser by doing `localStorage.debug = 'libp2p:*'`

# demo3: peer discovery in nodejs

```
node demo3
```

# demo4: peer discovery in go and chrome using webtransport (nodejs doesn't support webtransport yet)

```
go run demo4.go --private-key aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa --port 9999
```
- open `demo2.html` in chrome (no webtransport in firefox/safari yet) and add `?<multiaddress>` to the URL, e.g. `demo2.html?/ip4/127.0.0.1/udp/56916/quic-v1/webtransport/certhash/uEiBA6BEvjGlHJb7ZhLKazj6VFl03pX7kbDN4SU2gFM5yuw/certhash/uEiCeZhADwPPyWJe4Ert9nlpXyDv9RF4b_0DdCIqHAQdvQQ/p2p/12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq387JEYKDBj4kx6nXTN`
> NOTE: can debug libp2p connection issues in the browser by doing `localStorage.debug = 'libp2p:*'`
