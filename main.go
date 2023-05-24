package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

const vpHeight int = 5
const vpWidth int = 30

func main() {
	var appType string

	fmt.Println("Choose type server (s) or client (c)")
	fmt.Scanln(&appType)

	switch appType {
	case "s":
		RunServer()
	case "c":
		var username string
		fmt.Print("Your name: ")
		fmt.Scanln(&username)

		p := tea.NewProgram(initialModel(username))
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	}
}

type (
	errMsg error
)

type srvMessageMsg string

type model struct {
	viewport      viewport.Model
	username      string
	messages      []string
	textarea      textarea.Model
	senderStyle   lipgloss.Style
	receiverStyle lipgloss.Style
	conn          net.Conn
	reader        bufio.Reader
	writer        bufio.Writer
	err           error
}

func getMessage(m model) tea.Cmd {
	return func() tea.Msg {
		msg, err := m.reader.ReadString('\n')
		if err != nil {
			return tea.Quit
		}
		return srvMessageMsg(wordwrap.String(strings.TrimRight(msg, "\n"), vpWidth-5))
	}
}

func initialModel(username string) model {
	c, err := ConnectToServer()
	HandleError(err, false)

	ta := textarea.New()
	ta.Placeholder = "Send a message"
	ta.Focus()

	ta.Prompt = "| "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(vpWidth, vpHeight)
	vp.SetContent("Welcome to chat room!\nType a message and press Enter to send.")

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:      ta,
		messages:      []string{},
		username:      username,
		viewport:      vp,
		senderStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		receiverStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("6")),
		conn:          c,
		reader:        *bufio.NewReader(c),
		writer:        *bufio.NewWriter(c),
		err:           nil,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, getMessage(m))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		taCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, taCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			message := m.textarea.Value()
			m.writer.WriteString(fmt.Sprintf("%s: %s", m.username, message+"\n"))

			err := m.writer.Flush()
			if err != nil {
				panic(err)
			}

			m.messages = append(m.messages, m.senderStyle.Render("You: ")+wordwrap.String(message, vpWidth-5))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}
	case errMsg:
		m.err = msg
		return m, nil
	case srvMessageMsg:
		m.messages = append(m.messages, m.receiverStyle.Render(string(msg)))
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
	}

	return m, tea.Batch(taCmd, vpCmd, getMessage(m))
}

func (m model) View() string {
	return fmt.Sprintf("%s\n\n%s", m.viewport.View(), m.textarea.View()) + "\n\n"
}

func HandleError(err error, p bool) {
	if err == nil {
		return
	}
	if p {
		panic(err)
	}
	fmt.Println(err)
	os.Exit(1)
}
