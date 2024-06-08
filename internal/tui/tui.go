package tui

import (
	"fmt"

	"github.com/boldandbrad/pts/internal/pkg/structs"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

const (
	columnKeyName             = "name"
	columnKeyPlateAppearances = "pa"
	columnKeyPosPoints        = "pospts"
	columnKeyPosPointsPerPA   = "posptsperpa"
	columnKeyNegPoints        = "negpts"
	columnKeyNegPointsPerPA   = "negptsperpa"
	columnKeyPoints           = "pts"
	columnKeyPointsPerPA      = "ptsperpa"
)

var (
	styleSubtle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888"))

	styleBase = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#eee")).
			BorderForeground(lipgloss.Color("#89b4fa")).
			Align(lipgloss.Right)
)

type Model struct {
	statTable table.Model
	team      string
	year      string
}

func makeRow(stick structs.Stick) table.Row {
	return table.NewRow(table.RowData{
		columnKeyName: stick.Name,
		columnKeyPlateAppearances: func() string {
			if stick.PlateAppearances > 0 {
				return fmt.Sprintf("%d", stick.PlateAppearances)
			}
			return "0"
		}(),
		columnKeyPosPoints: func() string {
			if stick.PositivePoints() > 0 {
				return fmt.Sprintf("%d", stick.PositivePoints())
			}
			return "0"
		}(),
		columnKeyPosPointsPerPA: func() string {
			if stick.PlateAppearances > 0 {
				return fmt.Sprintf("%.3f", float64(stick.PositivePoints())/float64(stick.PlateAppearances))
			}
			return "0"
		}(),
		columnKeyNegPoints: func() string {
			if stick.NegativePoints() > 0 {
				return fmt.Sprintf("-%d", stick.NegativePoints())
			}
			return "0"
		}(),
		columnKeyNegPointsPerPA: func() string {
			if stick.PlateAppearances > 0 {
				return fmt.Sprintf("-%.3f", float64(stick.NegativePoints())/float64(stick.PlateAppearances))
			}
			return "0"
		}(),
		columnKeyPoints: func() string {
			return fmt.Sprintf("%d", stick.PositivePoints()-stick.NegativePoints())
		}(),
		columnKeyPointsPerPA: func() string {
			if stick.PlateAppearances > 0 {
				return fmt.Sprintf("%.3f", float64(stick.PositivePoints()-stick.NegativePoints())/float64(stick.PlateAppearances))
			}
			return "0"
		}(),
	})

}

func NewModel(sticks []structs.Stick, year string, team string) Model {
	var tableRows []table.Row
	for _, stick := range sticks {
		tableRows = append(tableRows, makeRow(stick))
	}

	model := Model{
		statTable: table.New([]table.Column{
			table.NewColumn(columnKeyName, "Name", 20).WithStyle(lipgloss.NewStyle().Align(lipgloss.Left)),
			table.NewColumn(columnKeyPlateAppearances, "PA", 6),
			table.NewColumn(columnKeyPosPoints, "+P", 6),
			table.NewColumn(columnKeyPosPointsPerPA, "+P/PA", 6),
			table.NewColumn(columnKeyNegPoints, "-P", 6),
			table.NewColumn(columnKeyNegPointsPerPA, "-P/PA", 6),
			table.NewColumn(columnKeyPoints, "P", 6),
			table.NewColumn(columnKeyPointsPerPA, "P/PA", 6).WithStyle(lipgloss.NewStyle().Bold(true)),
		}).
			WithRows(tableRows).
			BorderRounded().
			WithBaseStyle(styleBase).
			WithPageSize(12).
			SortByDesc(columnKeyPointsPerPA).
			Focused(true),
		team: team,
		year: year,
	}

	model.updateFooter()

	return model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.statTable, cmd = m.statTable.Update(msg)
	cmds = append(cmds, cmd)

	m.updateFooter()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) updateFooter() {
	// highlightedRow := m.statTable.HighlightedRow()

	footerText := fmt.Sprintf(
		"Viewing %s %s - Pg. %d/%d", m.team, m.year,
		m.statTable.CurrentPage(),
		m.statTable.MaxPages(),
	)

	m.statTable = m.statTable.WithStaticFooter(footerText)
}

func (m Model) View() string {
	// selected := m.statTable.HighlightedRow().Data[0].(string)
	view := lipgloss.JoinVertical(
		lipgloss.Left,
		styleSubtle.Render("Press ctrl+c or esc to quit."),
		m.statTable.View(),
	) + "\n"

	return lipgloss.NewStyle().MarginLeft(1).Render(view)
}
