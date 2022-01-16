package main

import (
	"fmt"
	"io"

	"github.com/Pragmatic-Kernel/EveGonline/common"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 40
const defaultWidth = 20

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("226"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type model struct {
	list     list.Model
	items    []item
	quitting bool
}

type model2 struct {
	list     list.Model
	km       string
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model2) Init() tea.Cmd {
	return nil
}

func (m model2) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "b":
			m := model{list: m.list}
			return m, nil

		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "q":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	return m, cmd
}

func (m model2) View() string {
	if m.quitting {
		return quitTextStyle.Render("Bye!")
	}
	return "\n" + m.km
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch keypress := msg.String(); keypress {
		case "q":
			m.quitting = true
			return m, tea.Quit
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "e":
			_, ok := m.list.SelectedItem().(item)
			if ok {
				item := m.list.SelectedItem().(item)
				enrichedKM, err := getKillmail(fmt.Sprint(item.ID))
				if err != nil {
					return m, tea.Quit
				}
				kmString, _ := formatKillmail(enrichedKM)
				if err != nil {
					return m, tea.Quit
				}
				return model2{m.list, kmString, false}, nil
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return quitTextStyle.Render("See you soon!")
	}
	return "\n" + m.list.View()
}

type item common.EnrichedKMShort

func (i item) FilterValue() string {
	km := common.EnrichedKMShort(i)
	res := formatKillmailShort(&km)
	return res
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	km := common.EnrichedKMShort(i)
	str := formatKillmailShort(&km)
	str = fmt.Sprintf("%d. %s", index+1, str)
	// Pad result to mitigate the index < 2 chars
	if index < 9 {
		str = " " + str
	}
	if index < 99 {
		str = " " + str
	}

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render("> " + s)
		}
	}

	fmt.Fprint(w, fn(str))
}
