# CLI Chat
Made with
- go 1.20
- [bubble tea](https://pkg.go.dev/github.com/charmbracelet/bubbletea)

## Usage
1. Clone repository
```
git clone https://github.com/Gasper3/cli-chat.git
```
2. Inside of cli-chat directory run build command or simply use `go run`
```
go run .
```
or
```
go build -o bin/cli-chat .
```
3. Choose if you want to run server or client (`s` or `c`)
4. As client choose your username
5. Let's Go and chat

## Client commands
Client can use some commands:
- `/leave` - leaves chat and exits application
- `/setusername` - sets username for the first time. Next calls will result in `username already set` error
