package editor

import "github.com/wx13/sith/file"

func (editor *Editor) Search(multiFile bool) {
	searchTerm := editor.screen.GetPromptAnswer("search:", &editor.searchHist)
	if searchTerm == "" {
		editor.file.NotifyUser("Cancelled")
		return
	}
	editor.MultiFileSearch(searchTerm, multiFile)
}

func (editor *Editor) MarkedSearch(searchTerm string) (int, int, error) {
	editor.file.MultiCursor = editor.file.MultiCursor.OuterMost()
	maxRow := editor.file.MultiCursor[1].Row()
	row, col, err := editor.file.Buffer[:maxRow+1].Search(searchTerm, editor.file.MultiCursor[0], false)
	if err == nil {
		editor.file.CursorGoTo(row, col)
	} else {
		editor.file.NotifyUser("Not Found")
	}
	return row, col, err
}

func (editor *Editor) MultiFileSearch(searchTerm string, multiFile bool) (int, int, error) {

	if len(editor.file.MultiCursor) > 1 {
		return editor.MarkedSearch(searchTerm)
	}

	// Search remainder of current file.
	row, col, err := editor.file.Buffer.Search(searchTerm, editor.file.MultiCursor[0], false)
	if err == nil {
		editor.file.CursorGoTo(row, col)
		return row, col, err
	}

	// Search other files.
	if multiFile {
		for idx := editor.fileIdx + 1; idx != editor.fileIdx; idx++ {
			if idx >= len(editor.files) {
				idx = 0
			}
			theFile := editor.files[idx]
			row, col, err := theFile.Buffer.Search(searchTerm, file.MakeCursor(0, -1), false)
			if err == nil {
				editor.SwitchFile(idx)
				editor.file.CursorGoTo(row, col)
				return row, col, err
			}
		}
	}

	// Search start of current file.
	row, col, err = editor.file.Buffer.Search(searchTerm, file.MakeCursor(0, -1), false)
	if err == nil {
		editor.file.CursorGoTo(row, col)
		return row, col, err
	}

	editor.file.NotifyUser("Not Found")
	return row, col, err
}

func (editor *Editor) SearchAndReplace(multiFile bool) {
	searchTerm := editor.screen.GetPromptAnswer("search:", &editor.searchHist)
	if searchTerm == "" {
		editor.screen.Notify("Cancelled")
		return
	}

	replaceTerm := editor.screen.GetPromptAnswer("replace:", &editor.replaceHist)

	replaceAll, err := editor.screen.AskYesNo("Replace All?")
	if err != nil {
		editor.screen.Notify("Cancelled")
		return
	}

	if len(editor.file.MultiCursor) > 1 {
		editor.MarkedSearchAndReplace(searchTerm, replaceTerm, replaceAll)
	} else {
		editor.MultiFileSearchAndReplace(searchTerm, replaceTerm, multiFile, replaceAll)
	}
}

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
