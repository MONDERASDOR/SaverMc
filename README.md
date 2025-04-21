# Go Minecraft Server (Java Edition)

This is a minimal Minecraft server implementation in Go, targeting the Java Edition protocol.

## Features
- Listens for TCP connections on port 25565
- Responds to Minecraft ping/status requests (shows up in the multiplayer list)
- Publicly accessible (binds to 0.0.0.0)

## Usage

1. Install Go 1.21 or later
2. Run:
   ```sh
   go run main.go
   ```
3. Add the server in your Minecraft client (Java Edition)
4. To allow connections from other devices or the internet:
   - Make sure port 25565 is open on your firewall
   - If behind a router, port forward 25565 to your server’s local IP address

## Roadmap
- [x] Ping/status response
- [x] Login handshake
- [ ] Player join/leave
- [ ] World/chunk handling

## License
MIT
