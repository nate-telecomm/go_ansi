package ansi

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	_ "time"

	"golang.org/x/term"
)

// --------------------
// captureKey
// --------------------

// captureKey reads a key press from stdin in raw mode and returns a key type and value.
// It returns one of "Character", "Arrow", "Special" (or "error" if something goes wrong).
func captureKey() (string, string) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "error", err.Error()
	}
	defer term.Restore(fd, oldState)

	b := make([]byte, 3)
	n, err := os.Stdin.Read(b)
	if err != nil {
		return "error", err.Error()
	}
	keyStr := string(b[:n])

	if runtime.GOOS == "windows" {
		// Windows-specific handling (roughly similar to Python’s msvcrt.getch)
		if n > 0 && (b[0] == 0 || b[0] == 224) {
			if n < 2 {
				os.Stdin.Read(b[1:2])
			}
			switch b[1] {
			case 'H':
				return "Arrow", "up"
			case 'P':
				return "Arrow", "down"
			case 'K':
				return "Arrow", "left"
			case 'M':
				return "Arrow", "right"
			}
		}
		if n == 1 {
			if b[0] == 8 {
				return "Special", "backspace"
			} else if b[0] == 13 {
				return "Special", "enter"
			}
		}
		return "Character", keyStr
	} else {
		// Unix-like handling
		if keyStr == "\x1b[A" {
			return "Arrow", "up"
		} else if keyStr == "\x1b[B" {
			return "Arrow", "down"
		} else if keyStr == "\x1b[C" {
			return "Arrow", "right"
		} else if keyStr == "\x1b[D" {
			return "Arrow", "left"
		} else if keyStr == "\x7f" {
			return "Special", "backspace"
		} else if keyStr == "\r" || keyStr == "\n" {
			return "Special", "enter"
		} else {
			return "Character", keyStr
		}
	}
}

// --------------------
// Colors
// --------------------

const (
	Black       = "\033[0;30m"
	Red         = "\033[0;31m"
	Green       = "\033[0;32m"
	Brown       = "\033[0;33m"
	Blue        = "\033[0;34m"
	Purple      = "\033[0;35m"
	Cyan        = "\033[0;36m"
	LightGray   = "\033[0;37m"
	DarkGray    = "\033[1;30m"
	LightRed    = "\033[1;31m"
	LightGreen  = "\033[1;32m"
	Yellow      = "\033[1;33m"
	LightBlue   = "\033[1;34m"
	LightPurple = "\033[1;35m"
	LightCyan   = "\033[1;36m"
	LightWhite  = "\033[1;37m"
	Bold        = "\033[1m"
	Faint       = "\033[2m"
	Italic      = "\033[3m"
	Underline   = "\033[4m"
	Blink       = "\033[5m"
	Negative    = "\033[7m"
	Crossed     = "\033[9m"
	End         = "\033[0m"
)

// --------------------
// Fill
// --------------------

// Fill fills the terminal with light blue negative-colored spaces.
func Fill() {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println("Error getting terminal size:", err)
		return
	}
	for v := 0; v < height; v++ {
		MovePos(v+1, 1)
		fmt.Print(LightBlue + Negative + strings.Repeat(" ", width) + End)
	}
}

// --------------------
// Cursor Functions
// --------------------

// ShowCursor makes the cursor visible.
func ShowCursor() {
	fmt.Print("\033[?25h")
}

// HideCursor hides the cursor.
func HideCursor() {
	fmt.Print("\033[?25l")
}

// SaveCursor saves the current cursor position.
func SaveCursor() {
	fmt.Print("\033[s")
}

// LoadCursor restores the saved cursor position.
func LoadCursor() {
	fmt.Print("\033[u")
}

// MovePos moves the cursor to a specific line and column.
func MovePos(line, col int) {
	fmt.Printf("\033[%d;%dH", line, col)
}

// WritePos writes a string at a given line and column.
// If save is true, the current cursor position is saved and then restored.
func WritePos(line, col int, str string, save bool) {
	if save {
		SaveCursor()
	}
	fmt.Printf("\033[%d;%dH%s", line, col, str)
	if save {
		LoadCursor()
	}
}

// Move moves the cursor in the given direction ("up", "down", "right", or "left")
// by n positions.
func Move(direc string, n int) {
	var code string
	switch direc {
	case "up":
		code = "A"
	case "down":
		code = "B"
	case "right":
		code = "C"
	case "left":
		code = "D"
	default:
		return
	}
	fmt.Printf("\033[%d%s", n, code)
}

// --------------------
// nPrint
// --------------------

// nPrint clears the current line (if newline is false) and then prints the string.
// If empty is false it prefixes the string with a character (e.g. "#").
func nPrint(str, character string, newline, empty bool) {
	if !newline {
		fmt.Print("\r\033[2K")
	} else {
		fmt.Println()
	}
	if !empty {
		fmt.Printf("[%s] %s", character, str)
	} else {
		fmt.Print(str)
	}
}

// --------------------
// dInput with Autocompletion
// --------------------

// findClosestMatch returns the first completion that starts with input.
func findClosestMatch(input string, completions []string) string {
	if input == "" {
		return ""
	}
	for _, word := range completions {
		if strings.HasPrefix(word, input) {
			return word
		}
	}
	return ""
}

// autocomplete appends the remaining characters of the closest match.
func autocomplete(input string, completions []string) string {
	closest := findClosestMatch(input, completions)
	if closest != "" && closest != input {
		return input + closest[len(input):]
	}
	return input
}

// dInput provides an interactive input prompt with autocomplete based on a list of completions.
func dInput(completions []string, prompt string) string {
	var text []rune
	NPrint(prompt+" ", "#", false, true)
	for {
		keyType, key := captureKey()
		if keyType == "Special" {
			if key == "enter" {
				break
			} else if key == "backspace" && len(text) > 0 {
				text = text[:len(text)-1]
			}
		} else if keyType == "Character" {
			text = append(text, []rune(key)...)
		}
		un := string(text)
		fin := ""
		// Process each word separately, coloring correctly if it matches a completion.
		words := strings.Split(un, " ")
		for i, word := range words {
			if word != "" {
				match := false
				for _, comp := range completions {
					if word == comp {
						match = true
						break
					}
				}
				if match {
					fin += Green + word + End
				} else {
					fin += Red + word + End
				}
				if i < len(words)-1 {
					fin += " "
				}
			}
		}
		// Autocomplete for the last word
		lastWord := ""
		if len(words) > 0 {
			lastWord = words[len(words)-1]
		}
		autoWord := autocomplete(lastWord, completions)
		if len(autoWord) > len(lastWord) {
			fin += Faint + autoWord[len(lastWord):] + End
		}
		NPrint(prompt+" "+fin, "#", false, true)
	}
	fmt.Println()
	return string(text)
}

// --------------------
// MultiProgressBar
// --------------------

// ProgressBar represents an individual progress bar.
type ProgressBar struct {
	Progress int
	Total    int
	Line     int
}

// MultiProgressBar manages several progress bars concurrently.
type MultiProgressBar struct {
	Bars map[string]*ProgressBar
	Lock sync.Mutex
}

// NewMultiProgressBar creates and returns a new MultiProgressBar.
func NewMultiProgressBar() *MultiProgressBar {
	return &MultiProgressBar{
		Bars: make(map[string]*ProgressBar),
	}
}

// AddBar adds a new progress bar with the given name and total.
func (mpb *MultiProgressBar) AddBar(name string, total int) {
	mpb.Lock.Lock()
	defer mpb.Lock.Unlock()
	mpb.Bars[name] = &ProgressBar{Progress: 0, Total: total, Line: len(mpb.Bars)}
}

// UpdateBar updates the progress of a named bar.
func (mpb *MultiProgressBar) UpdateBar(name string, progress int) {
	mpb.Lock.Lock()
	defer mpb.Lock.Unlock()
	if bar, ok := mpb.Bars[name]; ok {
		if progress > bar.Total {
			bar.Progress = bar.Total
		} else {
			bar.Progress = progress
		}
		mpb.draw()
	}
}

// draw renders all the progress bars.
func (mpb *MultiProgressBar) draw() {
	// Move cursor up for the number of bars and clear each line.
	for i := 0; i < len(mpb.Bars); i++ {
		fmt.Print("\033[F") // Move cursor up one line.
		fmt.Print("\033[K") // Clear the line.
	}
	// Sort the bars by their line number.
	type barEntry struct {
		Name string
		Bar  *ProgressBar
	}
	var entries []barEntry
	for name, bar := range mpb.Bars {
		entries = append(entries, barEntry{Name: name, Bar: bar})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Bar.Line < entries[j].Bar.Line
	})
	// Draw each progress bar.
	for _, entry := range entries {
		bar := entry.Bar
		percent := 1.0
		if bar.Total > 0 {
			percent = float64(bar.Progress) / float64(bar.Total)
		}
		barLen := 50
		filledLen := int(float64(barLen) * percent)
		filled := strings.Repeat(Green+"█"+End, filledLen)
		empty := strings.Repeat("-", barLen-filledLen)
		fmt.Printf("%s: [%s%s] %d/%d\n", entry.Name, filled, empty, bar.Progress, bar.Total)
	}
	// Reset any attributes.
	fmt.Print("\033[0m")
}

// FinishBar sets a progress bar to complete.
func (mpb *MultiProgressBar) FinishBar(name string) {
	mpb.Lock.Lock()
	defer mpb.Lock.Unlock()
	if bar, ok := mpb.Bars[name]; ok {
		bar.Progress = bar.Total
		mpb.draw()
	}
}

// RemoveBar removes a progress bar.
func (mpb *MultiProgressBar) RemoveBar(name string) {
	mpb.Lock.Lock()
	defer mpb.Lock.Unlock()
	delete(mpb.Bars, name)
	mpb.recalculateLines()
	mpb.draw()
}

// recalculateLines resets the line numbers for each progress bar.
func (mpb *MultiProgressBar) recalculateLines() {
	lineNum := 0
	for _, bar := range mpb.Bars {
		bar.Line = lineNum
		lineNum++
	}
}

// --------------------
// Screen Management
// --------------------

// NewScreen switches to an alternate screen and clears it.
func NewScreen() {
	fmt.Print("\033[?1049h")
	ClearScreen()
}

// ClearScreen clears the terminal.
func ClearScreen() {
	fmt.Print("\033[2J")
}

// ExitScreen switches back from the alternate screen.
func ExitScreen() {
	fmt.Print("\033[?1049l")
}

func Gap(word string, num int, space string) string {
  if num > len(word) {
    return strings.Repeat(space, num-len(word))
  }
    return strings.Repeat(space, len(word)-num)
}
