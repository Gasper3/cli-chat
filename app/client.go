package app

import (
	"bufio"
	"chat-app/utils"
	"flag"
	"fmt"
	"net"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

const vpHeight int = 5
const vpWidth int = 30

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
	connReader    bufio.Reader
	connWriter    bufio.Writer
	err           error
}

func (m model) formatSendMessage(msg string) string {
	return m.senderStyle.Render("You: ") + wordwrap.String(msg, vpWidth-5)
}

func (m *model) appendToViewport(msg string) {
	m.messages = append(m.messages, msg)
	m.viewport.SetContent(strings.Join(m.messages, "\n"))
	m.textarea.Reset()
	m.viewport.GotoBottom()
}

func (m model) Init() tea.Cmd {
	m.connWriter.WriteString(fmt.Sprintf("/setusername::%s\n", m.username))
	m.connWriter.Flush()
	return tea.Sequence(tea.ClearScreen, textarea.Blink, getMessageCmd(m))
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

			m.connWriter.WriteString(message + "\n")
			err := m.connWriter.Flush()
			utils.FatalOnError(err)

			if message == "/leave" {
				return m, tea.Quit
			}

			m.appendToViewport(m.formatSendMessage(message))
		}
	case errMsg:
		m.err = msg
		return m, nil
	case srvMessageMsg:
		m.appendToViewport(string(msg))
	}

	return m, tea.Batch(taCmd, vpCmd, getMessageCmd(m))
}

func (m model) View() string {
	return fmt.Sprintf("%s\n\n%s", m.viewport.View(), m.textarea.View()) + "\n\n"
}

func InitialModel(username string) model {
	c, err := connectToServer()
	utils.FatalOnError(err)

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
		connReader:    *bufio.NewReader(c),
		connWriter:    *bufio.NewWriter(c),
		err:           nil,
	}
}

func connectToServer() (net.Conn, error) {
	addr := flag.String("h", "127.0.0.1", "Host address")
	port := flag.String("p", "8000", "Port")
	flag.Parse()

	return net.Dial("tcp", *addr+":"+*port)
}

func getMessageCmd(m model) tea.Cmd {
	return func() tea.Msg {
		msg, err := m.connReader.ReadString('\n')
		if err != nil {
			return tea.Quit
		}
		return srvMessageMsg(wordwrap.String(strings.TrimRight(msg, "\n"), vpWidth-5))
	}
}
