package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/Pragmatic-Kernel/EveGonline/common"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 40
const defaultWidth = 20

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("226"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)
var (
	titleStyle2 = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle2.Copy().BorderStyle(b)
	}()
)

type model struct {
	list     list.Model
	items    []item
	quitting bool
	width    int
	height   int
}

type model2 struct {
	list     list.Model
	km       string
	quitting bool
	viewport viewport.Model
	width    int
	height   int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model2) Init() tea.Cmd {
	return nil
}

func (m model2) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "b":
			m := model{list: m.list, width: m.width, height: m.height}
			return m, nil

		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "q":
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMarginHeight

	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model2) View() string {
	if m.quitting {
		return quitTextStyle.Render("Bye!")
	}
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

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
				kmString, err := formatKillmail(enrichedKM)
				if err != nil {
					return m, tea.Quit
				}
				if *debug {
					log.Println("km:")
					log.Println(kmString)
				}
				m2 := model2{m.list, kmString, false, viewport.New(m.width, m.height-7), m.width, m.height}
				m2.viewport.SetContent(kmString)
				m2.viewport.HighPerformanceRendering = false
				return m2, nil
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		return m, nil

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

func (m model2) headerView() string {
	title := titleStyle2.Render("Kill Details")
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model2) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
