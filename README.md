# CLI Chat
Made with
- go 1.20
- [bubble tea](https://pkg.go.dev/github.com/charmbracelet/bubbletea)

## Usage
1. Start server
```bash
./server-darwin-arm64
```

2. Start client and choose your username
```bash
./client-darwin-arm64
```

3. Let's chat!

## Client commands
Client can use some commands:
- `/leave` - leaves chat and exits application
- `/setusername` - sets username for the first time. Next calls will result in `username already set` error

