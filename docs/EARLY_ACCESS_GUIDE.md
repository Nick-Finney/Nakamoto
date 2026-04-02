# Nakamoto Early Access — Tester Guide

Welcome to the Nakamoto Early Access program! You've been invited to test a new blockchain built from the ground up with privacy, decentralized storage, and Bitcoin integration at its core.

---

## Quick Start — Docker (Recommended)

The fastest and safest way to run a Nakamoto node. Docker sandboxes the node so it **cannot access your files or system** — only network ports are exposed.

**One command (via Tor):**
```bash
curl -o nakamoto.tar http://772lyewe5kdag26koumq4ifzmhtwlp2rtwkhpscdrnvkem5l6nh67iad.onion/files/nakamoto-testnet.tar && docker load -i nakamoto.tar && docker run -p 8089:8089 -p 9333:9333 nakamoto-testnet
```

Then open **http://localhost:8089** in your browser. That's it.

**Manage your node:**
```bash
# Run in background
docker run -d --name nakamoto -p 8089:8089 -p 9333:9333 nakamoto-testnet

# View logs
docker logs -f nakamoto

# Stop
docker stop nakamoto

# Restart
docker start nakamoto

# Reset everything (fresh start)
docker rm nakamoto && docker run -d --name nakamoto -p 8089:8089 -p 9333:9333 nakamoto-testnet
```

> **Don't have Docker?** Install it from https://docs.docker.com/get-docker/ (free, works on Windows/Mac/Linux).

---

## Alternative: Direct Install (No Docker)

If you prefer to run the binary directly (no sandboxing):

### 1. Download

**Option A — Tor (recommended for privacy):**
Visit the onion download page in Tor Browser:
```
http://772lyewe5kdag26koumq4ifzmhtwlp2rtwkhpscdrnvkem5l6nh67iad.onion
```

**Option B — Direct download (if you have a link from the dev team):**
Download the binary for your platform:
- `nakamoto-linux-amd64` — Linux
- `nakamoto-windows-amd64.exe` — Windows
- `NakamotoSetup-2.8.0.exe` — Windows installer
- `nakamoto-chrome-extension.zip` — Chrome browser extension

### 2. Verify Download (optional but recommended)

Each file has a `.sha256` checksum file. Verify with:
```bash
# Linux/Mac
sha256sum -c nakamoto-linux-amd64.sha256

# Windows (PowerShell)
Get-FileHash nakamoto-windows-amd64.exe -Algorithm SHA256
```

### 3. Install & Run

**Linux:**
```bash
chmod +x nakamoto-linux-amd64
./nakamoto-linux-amd64 --config nakamoto-config.json
```

**Windows (installer):**
Run `NakamotoSetup-2.8.0.exe` and follow the prompts.

**Windows (standalone):**
```
nakamoto-windows-amd64.exe --config nakamoto-config.json
```

**Chrome Extension:**
1. Unzip `nakamoto-chrome-extension.zip`
2. Open Chrome → `chrome://extensions`
3. Enable "Developer mode" (top right)
4. Click "Load unpacked" → select the unzipped folder
5. The Nakamoto browser icon appears in your toolbar

### 4. Create a Wallet

Once the node is running, open the browser UI (Chrome extension or desktop app) and:
1. Go to the **Wallet** tab
2. Click **Create Wallet**
3. Save your wallet ID — you'll need it to join the chat room

### 5. Join the Early Access Chat Room

In the browser UI:
1. Go to the **Chat** tab
2. Look for the room **"Nakamoto Early Access"**
3. Request to join (the room requires approval — the dev team will approve you)

This is where we coordinate, discuss bugs, and share feedback. All messages are end-to-end encrypted on the blockchain.

---

## What to Test

We'd love your feedback on everything, but here are the priorities:

### High Priority
- **Node sync** — Does your node connect to the network and sync blocks?
- **Wallet** — Can you create a wallet and see your balance?
- **Browser** — Does the built-in browser load `nak://` sites?
- **Chat** — Can you send and receive messages in the Early Access room?

### Medium Priority
- **Staking** — Try staking tokens as a validator
- **Storage** — Try the guardian/storage provider features
- **Settings** — Do preferences save and persist?

### Lower Priority
- **Bitcoin integration** — BTC-to-NAK conversion (testnet)
- **Fork governance** — Proposing and voting on network changes

---

## Reporting Bugs

Please report bugs in the **Nakamoto Early Access** chat room with:
1. **What you did** (steps to reproduce)
2. **What you expected**
3. **What actually happened**
4. **Your platform** (Windows/Linux/Mac, browser version)
5. **Screenshot** if applicable (you can share files in the chat room)

---

## Useful Info

| Item | Value |
|------|-------|
| Network | Testnet (tNAK tokens, no real value) |
| Block time | 12 seconds |
| API port | 8089 (default) |
| P2P port | 9333 (default) |
| Seed node | `772lyewe5kdag26koumq4ifzmhtwlp2rtwkhpscdrnvkem5l6nh67iad.onion:9333` |

---

## FAQ

**Q: Is this real money?**
A: No. This is testnet. Tokens (tNAK) have no monetary value. This is for testing only.

**Q: Do I need Tor?**
A: No, but it's recommended for privacy. The node connects to the seed via Tor automatically if available.

**Q: My node won't sync / connect?**
A: Make sure port 9333 is not blocked by your firewall. Post in the chat room and we'll help debug.

**Q: Can I run this on a VPS/server?**
A: Yes! The Linux binary runs great on any Ubuntu/Debian server. Just make sure ports 8089 and 9333 are accessible.

---

Thank you for testing! Your feedback directly shapes the future of Nakamoto.
