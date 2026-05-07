package ui

import (
	"fmt"
	"time"

	"acr-uploader/internal/image"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	baseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("63"))

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("240"))
)

type ImageSelector struct {
	table   table.Model
	images  []*image.ImageInfo
	ready   bool
}

func NewImageSelector(images []*image.ImageInfo) *ImageSelector {
	columns := []table.Column{
		{Title: "仓库名", Width: 30},
		{Title: "标签", Width: 15},
		{Title: "镜像ID", Width: 12},
		{Title: "创建时间", Width: 20},
		{Title: "大小", Width: 12},
	}

	rows := make([]table.Row, 0)
	for _, img := range images {
		createdTime := time.Unix(img.Created, 0).Format("2006-01-02 15:04:05")
		rows = append(rows, table.Row{
			truncateString(img.Name, 30),
			truncateString(img.Tag, 15),
			img.ID,
			createdTime,
			img.GetSizeHuman(),
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("63")).
		Bold(false)
	t.SetStyles(s)

	return &ImageSelector{
		table:  t,
		images: images,
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (i *ImageSelector) Init() tea.Cmd {
	return nil
}

func (i *ImageSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return i, tea.Quit
		case "ctrl+c":
			return i, tea.Quit
		}
	}

	var cmd tea.Cmd
	i.table, cmd = i.table.Update(msg)
	return i, cmd
}

func (i *ImageSelector) View() string {
	if !i.ready {
		return "Loading images..."
	}
	return baseStyle.Render(i.table.View())
}

func (i *ImageSelector) SelectedImage() *image.ImageInfo {
	selectedRow := i.table.SelectedRow()
	if len(selectedRow) == 0 {
		return nil
	}

	selectedIndex := i.table.Cursor()
	if selectedIndex >= 0 && selectedIndex < len(i.images) {
		return i.images[selectedIndex]
	}

	return nil
}

func (i *ImageSelector) SetReady(ready bool) {
	i.ready = ready
}

func SelectImageInteractive(images []*image.ImageInfo) (*image.ImageInfo, error) {
	if len(images) == 0 {
		return nil, fmt.Errorf("没有可用的镜像")
	}

	fmt.Println("使用上下方向键选择镜像，按 Enter 确认：")
	fmt.Println()

	selector := NewImageSelector(images)
	selector.SetReady(true)

	p := tea.NewProgram(selector)
	model, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("启动交互程序失败: %w", err)
	}

	selector = model.(*ImageSelector)
	return selector.SelectedImage(), nil
}
