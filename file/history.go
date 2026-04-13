package file

import (
	"fmt"
	"time"

	"github.com/wx13/sith/terminal"
	"github.com/wx13/sith/ui"
)

// ShowHistory displays a menu of saved buffer states and allows the user
// to preview and jump to a selected state.
func (file *File) ShowHistory() {
	if file.buffHist == nil {
		file.NotifyUser("No history available")
		return
	}

	states := file.buffHist.GetSavedStates()
	if len(states) == 0 {
		file.NotifyUser("No saved states")
		return
	}

	// Find the current state index for initial cursor position
	currentIdx := 0
	for i, s := range states {
		if s.IsCurrent {
			currentIdx = i
			break
		}
	}

	keyboard := terminal.NewKeyboard()
	keyboard.SetScreen(file.screen.GetTcell())
	menu := ui.NewMenu(file.screen, keyboard)

	for {
		// Build menu choices
		choices := file.formatStateChoices(states)

		// Show menu and get selection
		idx, action := menu.Choose(choices, currentIdx, "", "escape")
		menu.Clear()

		if action == "cancel" || action == "escape" {
			return
		}

		selectedState := states[idx]

		// If selecting current state, nothing to do
		if selectedState.IsCurrent {
			file.NotifyUser("Already at this state")
			currentIdx = idx
			continue
		}

		// Show diff preview
		accepted := file.showDiffPreview(selectedState, keyboard)
		if accepted {
			// Jump to the selected state
			buffer, mc := file.buffHist.JumpToState(selectedState)
			file.buffer.ReplaceBuffer(buffer)
			file.MultiCursor.ReplaceMC(mc)
			file.RequestFlush()
			file.NotifyUser(fmt.Sprintf("Jumped to %s", file.formatTimestamp(selectedState.Timestamp)))
			return
		}
		// Otherwise, go back to menu
		currentIdx = idx
	}
}

// formatStateChoices formats the state list for menu display.
func (file *File) formatStateChoices(states []StateInfo) []string {
	choices := make([]string, len(states))
	for i, s := range states {
		choices[i] = file.formatStateInfo(s)
	}
	return choices
}

// formatStateInfo formats a single state for display.
func (file *File) formatStateInfo(s StateInfo) string {
	var timestamp string
	if s.IsCurrent && !s.IsSaved {
		timestamp = "now          "
	} else {
		timestamp = file.formatTimestamp(s.Timestamp)
	}

	var delta string
	if s.IsCurrent {
		if s.IsSaved {
			delta = "(current)"
		} else {
			delta = "(unsaved)"
		}
	} else if s.LineDelta == 0 {
		delta = "same length"
	} else if s.LineDelta > 0 {
		delta = fmt.Sprintf("+%d lines", s.LineDelta)
	} else {
		delta = fmt.Sprintf("%d lines", s.LineDelta)
	}

	return fmt.Sprintf("%s  %s", timestamp, delta)
}

// formatTimestamp formats a timestamp for display, including date if not today.
func (file *File) formatTimestamp(t time.Time) string {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)
	weekAgo := today.AddDate(0, 0, -7)

	tDate := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	timeStr := t.Format("3:04:05 PM")

	if tDate.Equal(today) {
		return timeStr
	} else if tDate.Equal(yesterday) {
		return fmt.Sprintf("yesterday %s", t.Format("3:04 PM"))
	} else if tDate.After(weekAgo) {
		return fmt.Sprintf("%s %s", t.Format("Mon"), t.Format("3:04 PM"))
	} else if t.Year() == now.Year() {
		return fmt.Sprintf("%s %s", t.Format("Jan 2"), t.Format("3:04 PM"))
	} else {
		return t.Format("Jan 2, 2006 3:04 PM")
	}
}

// showDiffPreview shows a scrollable diff preview and returns true if user accepts.
func (file *File) showDiffPreview(state StateInfo, keyboard *terminal.Keyboard) bool {
	removals, additions := file.buffHist.GetStateDiff(state)

	// Build diff lines
	diffLines := []string{}
	diffColors := []terminal.Attribute{}

	if len(removals) == 0 && len(additions) == 0 {
		diffLines = append(diffLines, "(no line changes)")
		diffColors = append(diffColors, terminal.ColorDefault)
	}

	for _, line := range removals {
		display := "- " + line
		if len(display) > 60 {
			display = display[:57] + "..."
		}
		diffLines = append(diffLines, display)
		diffColors = append(diffColors, terminal.ColorRed)
	}

	for _, line := range additions {
		display := "+ " + line
		if len(display) > 60 {
			display = display[:57] + "..."
		}
		diffLines = append(diffLines, display)
		diffColors = append(diffColors, terminal.ColorGreen)
	}

	// Display parameters
	cols, rows := file.screen.Size()
	boxWidth := cols - 8
	boxHeight := rows - 8
	if boxHeight > len(diffLines)+4 {
		boxHeight = len(diffLines) + 4
	}
	if boxHeight < 6 {
		boxHeight = 6
	}
	col0 := 4
	row0 := 3

	scroll := 0
	maxScroll := len(diffLines) - (boxHeight - 4)
	if maxScroll < 0 {
		maxScroll = 0
	}

	for {
		// Draw box
		file.drawDiffBox(row0, col0, boxHeight, boxWidth, state, diffLines, diffColors, scroll)
		file.screen.Flush()

		// Get input
		cmd, _ := keyboard.GetKey()
		switch cmd {
		case "enter":
			file.clearBox(row0, col0, boxHeight, boxWidth)
			return true
		case "escape", "ctrlC":
			file.clearBox(row0, col0, boxHeight, boxWidth)
			return false
		case "arrowDown":
			if scroll < maxScroll {
				scroll++
			}
		case "arrowUp":
			if scroll > 0 {
				scroll--
			}
		case "pageDown":
			scroll += boxHeight - 4
			if scroll > maxScroll {
				scroll = maxScroll
			}
		case "pageUp":
			scroll -= boxHeight - 4
			if scroll < 0 {
				scroll = 0
			}
		}
	}
}

// drawDiffBox draws the diff preview box.
func (file *File) drawDiffBox(row0, col0, height, width int, state StateInfo,
	lines []string, colors []terminal.Attribute, scroll int) {

	borderColor := terminal.ColorBlue

	// Title
	title := fmt.Sprintf(" Preview: revert to %s ", file.formatTimestamp(state.Timestamp))
	if len(title) > width-2 {
		title = title[:width-2]
	}
	topBorder := title
	for len(topBorder) < width {
		topBorder += "-"
	}
	file.screen.WriteStringColor(row0, col0, topBorder, terminal.ColorWhite|terminal.AttrBold, borderColor)

	// Content area
	contentHeight := height - 4
	visibleLines := lines[scroll:]
	if len(visibleLines) > contentHeight {
		visibleLines = visibleLines[:contentHeight]
	}

	for i := 0; i < contentHeight; i++ {
		row := row0 + 1 + i
		file.screen.WriteStringColor(row, col0, "| ", borderColor, terminal.ColorDefault)
		if i < len(visibleLines) {
			lineColor := colors[scroll+i]
			display := visibleLines[i]
			if len(display) > width-4 {
				display = display[:width-7] + "..."
			}
			file.screen.WriteStringColor(row, col0+2, display, lineColor, terminal.ColorDefault)
			// Clear rest of line
			for j := len(display); j < width-4; j++ {
				file.screen.WriteString(row, col0+2+j, " ")
			}
		} else {
			for j := 0; j < width-4; j++ {
				file.screen.WriteString(row, col0+2+j, " ")
			}
		}
		file.screen.WriteStringColor(row, col0+width-2, " |", borderColor, terminal.ColorDefault)
	}

	// Scroll indicator
	scrollInfo := ""
	if len(lines) > contentHeight {
		scrollInfo = fmt.Sprintf(" [%d-%d of %d] ", scroll+1, scroll+len(visibleLines), len(lines))
	}

	// Bottom border with instructions
	instructions := " [Enter] Accept  [Esc] Back "
	bottomBorder := instructions + scrollInfo
	for len(bottomBorder) < width {
		bottomBorder += "-"
	}
	file.screen.WriteStringColor(row0+height-3, col0, bottomBorder, terminal.ColorWhite, borderColor)
}

// clearBox clears the diff preview box area.
func (file *File) clearBox(row0, col0, height, width int) {
	for row := row0; row < row0+height-2; row++ {
		for col := col0; col < col0+width; col++ {
			file.screen.WriteString(row, col, " ")
		}
	}
}
