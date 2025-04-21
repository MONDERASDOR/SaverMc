# Go Minecraft Server (Java Edition)

This is a minimal Minecraft server implementation in Go, targeting the Java Edition protocol.

## Features
- Listens for TCP connections on port 25565
- Responds to Minecraft ping/status requests (shows up in the multiplayer list)

## Usage

1. Install Go 1.21 or later
2. Run:
   ```sh
   go run main.go
   ```
3. Add the server in your Minecraft client (Java Edition)

## Roadmap
- [x] Ping/status response
- [ ] Login handshake
- [ ] Player join/leave
- [ ] World/chunk handling

## License
MIT
