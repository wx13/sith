package editor

import (
	"errors"

	"github.com/wx13/sith/ui"
)

// searchPrompt prompts the user for a search term.
func (editor *Editor) searchPrompt() (string, error) {
	prompt := ui.MakePrompt(editor.screen, editor.keyboard)
	searchTerm := prompt.GetAnswer("search:", &editor.searchHist, editor.AutoComplete)
	if searchTerm == "" {
		editor.file.NotifyUser("Cancelled")
		return "", errors.New("Cancelled")
	}
	return searchTerm, nil
}

// SearchLineFo searches the current line from cursor to the end.
func (editor *Editor) SearchLineFo() {
	searchTerm, err := editor.searchPrompt()
	if err == nil {
		editor.file.SearchLineFo(searchTerm)
	}
}

// SearchLineBa searches the current line from cursor to the start.
func (editor *Editor) SearchLineBa() {
	searchTerm, err := editor.searchPrompt()
	if err == nil {
		editor.file.SearchLineBa(searchTerm)
	}
}

// AllLineFo searches the current line from cursor to the start, and makes multiple
// cursors (one for each match).
func (editor *Editor) AllLineFo() {
	searchTerm, err := editor.searchPrompt()
	if err == nil {
		editor.file.AllLineFo(searchTerm)
	}
}

// AllLineBa searches the current line from cursor to the start, and makes multiple
// cursors (one for each match).
func (editor *Editor) AllLineBa() {
	searchTerm, err := editor.searchPrompt()
	if err == nil {
		editor.file.AllLineBa(searchTerm)
	}
}

// Search searches the entire buffer (or set of buffers if multiFile is true).
func (editor *Editor) Search(multiFile bool) {
	searchTerm, err := editor.searchPrompt()
	if err == nil {
		editor.MultiFileSearch(searchTerm, multiFile)
	}
}

// MarkedSearch searches between the cursors.
func (editor *Editor) MarkedSearch(searchTerm string) (int, int, error) {
	loop := false
	row, col, err := editor.file.MarkedSearch(searchTerm, loop)
	if err == nil {
		editor.file.CursorGoTo(row, col)
	} else {
		editor.file.NotifyUser("Not Found")
	}
	return row, col, err
}

func (editor *Editor) otherIndexes(curr, max int) []int {
	idxs := []int{}
	for idx := curr + 1; idx < max; idx++ {
		idxs = append(idxs, idx)
	}
	for idx := 0; idx < curr; idx++ {
		idxs = append(idxs, idx)
	}
	return idxs
}

// MultiFileSearch searches all the file buffers.
func (editor *Editor) MultiFileSearch(searchTerm string, multiFile bool) (int, int, error) {

	if editor.file.MultiCursor.Length() > 1 {
		return editor.MarkedSearch(searchTerm)
	}

	// Search remainder of current file.
	row, col, err := editor.file.SearchFromCursor(searchTerm)
	if err == nil {
		editor.file.CursorGoTo(row, col)
		return row, col, err
	}

	// Search other files.
	if multiFile && len(editor.files) > 0 {
		for _, idx := range editor.otherIndexes(editor.fileIdx, len(editor.files)) {
			theFile := editor.files[idx]
			row, col, err := theFile.SearchFromStart(searchTerm)
			if err == nil {
				editor.SwitchFile(idx)
				editor.file.CursorGoTo(row, col)
				return row, col, err
			}
		}
	}

	// Search start of current file.
	row, col, err = editor.file.SearchFromStart(searchTerm)
	if err == nil {
		editor.file.CursorGoTo(row, col)
		return row, col, err
	}

	editor.file.NotifyUser("Not Found")
	return row, col, err
}

// SearchAndReplace searches and replaces.
func (editor *Editor) SearchAndReplace(multiFile bool) {
	prompt := ui.MakePrompt(editor.screen, editor.keyboard)
	searchTerm := prompt.GetAnswer("search:", &editor.searchHist)
	if searchTerm == "" {
		editor.screen.Notify("Cancelled")
		return
	}

	replaceTerm := prompt.GetAnswer("replace:", &editor.replaceHist)

	replaceAll, err := prompt.AskYesNo("Replace All?")
	if err != nil {
		editor.screen.Notify("Cancelled")
		return
	}

	if editor.file.MultiCursor.Length() > 1 {
		editor.MarkedSearchAndReplace(searchTerm, replaceTerm, replaceAll)
	} else {
		editor.MultiFileSearchAndReplace(searchTerm, replaceTerm, multiFile, replaceAll)
	}
}

// MarkedSearchAndReplace does search-and-replace between cursors.
func (editor *Editor) MarkedSearchAndReplace(searchTerm, replaceTerm string, replaceAll bool) {
	for {

		row, col, err := editor.MultiFileSearch(searchTerm, false)

		if err == nil {
			err := editor.file.AskReplace(searchTerm, replaceTerm, row, col, replaceAll)
			if err != nil {
				editor.screen.Notify("Cancelled")
				return
			}
		} else {
			editor.screen.Notify("Not Found")
			break
		}

	}
}

// MultiFileSearchAndReplace is just like SearchAndReplace but for all the file buffers.
func (editor *Editor) MultiFileSearchAndReplace(searchTerm, replaceTerm string, multiFile, replaceAll bool) {

	var idx0, row0, col0 int
	idx0 = -1
	numMatches := 0
	mc := editor.file.MultiCursor.Dup()
	for {
		row, col, err := editor.MultiFileSearch(searchTerm, multiFile)
		if err == nil {
			if idx0 < 0 {
				idx0 = editor.fileIdx
				row0, col0 = row, col
			} else if idx0 == editor.fileIdx && row0 == row && col0 == col {
				break
			}
			numMatches++
		} else {
			break
		}
	}
	editor.file.MultiCursor = mc

	for {

		row, col, err := editor.MultiFileSearch(searchTerm, multiFile)
		numMatches--
		if numMatches < 0 {
			break
		}

		if err == nil {
			err := editor.file.AskReplace(searchTerm, replaceTerm, row, col, replaceAll)
			if err != nil {
				editor.screen.Notify("Cancelled")
				return
			}
		} else {
			editor.screen.Notify("Not Found")
			break
		}

	}
}
