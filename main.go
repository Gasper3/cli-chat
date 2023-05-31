package main

import (
	"fmt"
	"log"

	"chat-app/app"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var appType string

	fmt.Println("Choose type server (s) or client (c)")
	fmt.Scanln(&appType)

	switch appType {
	case "s":
		app.RunServer()
	case "c":
		var username string
		fmt.Print("Your name: ")
		fmt.Scanln(&username)

		p := tea.NewProgram(app.InitialModel(username))
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	}
}
