package editor

import "github.com/wx13/sith/file"

func (editor *Editor) Search(multiFile bool) {
	searchTerm := editor.screen.GetPromptAnswer("search:", &editor.searchHist)
	if searchTerm == "" {
		editor.screen.Notify("Cancelled")
		return
	}
	editor.MultiFileSearch(searchTerm, multiFile)
}

func (editor *Editor) MultiFileSearch(searchTerm string, multiFile bool) (int, int, error) {

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

	editor.screen.Notify("Not Found")
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

	editor.MultiFileSearchAndReplace(searchTerm, replaceTerm, multiFile, replaceAll)
}

func (editor *Editor) MultiFileSearchAndReplace(searchTerm, replaceTerm string, multiFile, replaceAll bool) {

	var idx0, row0, col0 int
	idx0 = -1
	numMatches := 0
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
