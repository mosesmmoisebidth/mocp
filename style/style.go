package style

import (
	"fmt"
	"strings"
	"time"
)

// ANSI color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
	Blink     = "\033[5m"
	BlinkFast = "\033[6m"
	Reverse   = "\033[7m"
	Hidden    = "\033[8m"
	Strike    = "\033[9m"

	// Regular colors
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// Bright colors
	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"
)

var Logo = fmt.Sprintf(`
%s╔═════════════════════════════════════════════════════════════╗
║   %s███╗   ███╗ ██████╗  ██████╗██████╗     %s©2025%s        ║
║   %s████╗ ████║██╔═══██╗██╔════╝██╔══██╗               %s║
║   %s██╔████╔██║██║   ██║██║     ██████╔╝               %s║
║   %s██║╚██╔╝██║██║   ██║██║     ██╔═══╝                %s║
║   %s██║ ╚═╝ ██║╚██████╔╝╚██████╗██║                    %s║
║   %s╚═╝     ╚═╝ ╚═════╝  ╚═════╝╚═╝                    %s║
╠═════════════════════════════════════════════════════════════╣
║ %sFile Transfer Made Simple - By Mucyo Moses%s                 ║
║ %shttps://github.com/mosesmmoisebidth%s                      ║
║ %shttps://moses.it.com%s                                     ║
╠═════════════════════════════════════════════════════════════╣
║ %s☕ Buy me a coffee: https://buymeacoffee.com/mucyomoses%s   ║
╚═════════════════════════════════════════════════════════════╝`,
	Cyan, BrightBlue, Yellow, Cyan,
	BrightBlue, Cyan,
	BrightBlue, Cyan,
	BrightBlue, Cyan,
	BrightBlue, Cyan,
	BrightBlue, Cyan,
	BrightGreen, Cyan,
	BrightYellow, Cyan,
	BrightMagenta, Cyan,
	BrightYellow, Cyan)

var CoffeeIcon = fmt.Sprintf(`
%s   ( ( (    
    ) ) )    %sBuy Me%s
  ........   %sA Coffee%s
  |      |]  
  \      /   
   '----'%s`, BrightYellow, BrightWhite, BrightYellow, BrightWhite, BrightYellow, Reset)

func ProgressBar(current, total int, prefix string) string {
	const width = 40
	unknownTotal := false
	if total <= 0 {
		// Unknown total (e.g., chunked upload). We'll treat it specially so the
		// progress bar still animates and the size string shows "?" for total.
		unknownTotal = true
		// Prevent division by zero in other calculations
		if current == 0 {
			total = 1
		} else {
			total = current
		}
	}

	// Calculate progress
	var filled int
	if unknownTotal {
		// Make the filled portion change as bytes arrive to give visual feedback.
		filled = current % (width + 1)
	} else {
		filled = int(float64(current) * float64(width) / float64(total))
	}
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}

	// New characters for a smooth, dense look
	filledChar := "▉"
	emptyChar := "·"
	filledPart := strings.Repeat(filledChar, filled)
	emptyPart := strings.Repeat(emptyChar, width-filled)

	// Percentage (only meaningful when total is known)
	var percentage float64
	if unknownTotal {
		percentage = 0.0
	} else {
		percentage = float64(current) * 100.0 / float64(total)
	}

	// Size display using FormatSize for consistency. If total is unknown, show "?"
	var sizeStr string
	if unknownTotal {
		sizeStr = FormatSize(int64(current)) + " / ?"
	} else {
		sizeStr = FormatSize(int64(current)) + " / " + FormatSize(int64(total))
	}

	// Colors
	border := BrightCyan
	fill := BrightGreen
	empty := BrightBlack
	pct := BrightYellow
	sz := BrightWhite

	bar := fmt.Sprintf("%s%s%s%s%s", fill, filledPart, Reset, empty, emptyPart)

	return fmt.Sprintf("%s %s[%s%s]%s %s%.1f%%%s %s%s%s",
		Bold+prefix+Reset,
		border,
		bar,
		Reset,
		border,
		pct, percentage, Reset,
		sz, sizeStr, Reset)
}

func Success(msg string) string {
	return fmt.Sprintf("%s%s✨ %s %s✨%s",
		BrightGreen, Bold, msg, Reset, BrightGreen)
}

func RetroFrame(content string) string {
	return fmt.Sprintf("%s╔══%s %s %s══╗%s\n%s║%s %s %s║%s\n%s╚════════╝%s",
		Cyan, BrightBlue, content, Cyan, Reset,
		Cyan, Reset, content, Cyan, Reset,
		Cyan, Reset)
}

// Backwards-compatible helpers expected by server/ui.go
var MainLogo = Logo

func InfoBox(title, content string) string {
	// Simple info box fallback
	return fmt.Sprintf("%s[ %s ]%s %s", BrightCyan, title, Reset, content)
}

func AnimatedProgressBar(current, total int64, prefix string) string {
	// Keep same appearance but accept int64 values
	return ProgressBar(int(current), int(total), prefix)
}

// PlainProgressBar returns the same information as ProgressBar but without
// ANSI escape sequences (safe for terminals that don't support colors).
func PlainProgressBar(current, total int64, prefix string) string {
	const width = 40
	if total <= 0 {
		// Unknown total: animate the bar by moving the filled length
		filled := int(current) % (width + 1)
		filledPart := strings.Repeat("#", filled)
		emptyPart := strings.Repeat("-", width-filled)
		sizeStr := FormatSize(current) + " / ?"
		return fmt.Sprintf("%s [%s%s] %.0f%% %s", prefix, filledPart, emptyPart, 0.0, sizeStr)
	}
	filled := int(float64(current) * float64(width) / float64(total))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	filledPart := strings.Repeat("#", filled)
	emptyPart := strings.Repeat("-", width-filled)
	percentage := float64(current) * 100.0 / float64(total)
	sizeStr := FormatSize(current) + " / " + FormatSize(total)
	return fmt.Sprintf("%s [%s%s] %.1f%% %s", prefix, filledPart, emptyPart, percentage, sizeStr)
}

// AnimatedPlainProgressBar accepts int64 and returns a plain bar
func AnimatedPlainProgressBar(current, total int64, prefix string) string {
	return PlainProgressBar(current, total, prefix)
}

// FormatDuration returns a mm:ss or hh:mm:ss string for the given duration
func FormatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	secs := int(d.Seconds())
	h := secs / 3600
	m := (secs % 3600) / 60
	s := secs % 60
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

// FormatRate formats bytes per second into a human readable string (e.g., 48.46KB/s)
func FormatRate(bps float64) string {
	if bps <= 0 {
		return "0B/s"
	}
	const (
		KB = 1024.0
		MB = KB * 1024.0
	)
	if bps >= MB {
		return fmt.Sprintf("%.2fMB/s", bps/MB)
	}
	if bps >= KB {
		return fmt.Sprintf("%.2fKB/s", bps/KB)
	}
	return fmt.Sprintf("%.0fB/s", bps)
}

// AnimatedProgressBar returns a styled progress bar with elapsed, ETA and rate
func AnimatedProgressBarWithStats(current, total int64, prefix string, rate float64, elapsed time.Duration) string {
	base := ProgressBar(int(current), int(total), prefix)
	// Build time/rate suffix
	elapsedStr := FormatDuration(elapsed)
	var suffix string
	if total <= 0 {
		// unknown total: show elapsed and rate
		suffix = fmt.Sprintf("[%s, %s]", elapsedStr, FormatRate(rate))
	} else {
		// known total: compute ETA
		var eta time.Duration
		if rate > 0 {
			remaining := float64(total - current)
			etaSecs := remaining / rate
			eta = time.Duration(etaSecs) * time.Second
		} else {
			eta = 0
		}
		etaStr := FormatDuration(eta)
		suffix = fmt.Sprintf("[%s<%s, %s]", elapsedStr, etaStr, FormatRate(rate))
	}
	return fmt.Sprintf("%s %s", base, suffix)
}

// AnimatedPlainProgressBar returns a plain (no color) progress bar with stats
func AnimatedPlainProgressBarWithStats(current, total int64, prefix string, rate float64, elapsed time.Duration) string {
	base := PlainProgressBar(current, total, prefix)
	elapsedStr := FormatDuration(elapsed)
	if total <= 0 {
		return fmt.Sprintf("%s [%s, %s]", base, elapsedStr, FormatRate(rate))
	}
	var eta time.Duration
	if rate > 0 {
		etaSecs := float64(total-current) / rate
		eta = time.Duration(etaSecs) * time.Second
	}
	return fmt.Sprintf("%s [%s<%s, %s]", base, elapsedStr, FormatDuration(eta), FormatRate(rate))
}

// FormatSize returns a human readable size string using bytes, KB or MB
func FormatSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d bytes", size)
	}
	if size < 1048576 {
		kb := float64(size) / 1024.0
		if kb > 0 && kb < 0.1 {
			kb = 0.1
		}
		return fmt.Sprintf("%.1f KB", kb)
	}
	mb := float64(size) / 1048576.0
	if mb > 0 && mb < 0.01 {
		mb = 0.01
	}
	return fmt.Sprintf("%.2f MB", mb)
}

// ColorForInterface returns a color escape sequence for a given interface name
func ColorForInterface(name string) string {
	// Lightweight heuristic: pick color based on name keywords
	lname := strings.ToLower(name)
	switch {
	case strings.Contains(lname, "loopback"), strings.Contains(lname, "lo"):
		return BrightYellow
	case strings.Contains(lname, "wi"), strings.Contains(lname, "wlan"), strings.Contains(lname, "wifi"):
		return BrightGreen
	case strings.Contains(lname, "eth"), strings.Contains(lname, "en"):
		return BrightCyan
	default:
		return BrightMagenta
	}
}

func LoadingSpinner(msg string) string {
	// Simple static spinner line for now
	return fmt.Sprintf("%s⠋ %s %s", BrightMagenta, msg, Reset)
}

func SuccessMessage(msg string) string {
	return Success(msg)
}

func RetroBox(content string) string {
	return RetroFrame(content)
}

func Coffee() string {
	return CoffeeIcon
}

func ErrorMessage(msg string) string {
	return fmt.Sprintf("%s%s✖ %s%s", BrightRed, Bold, msg, Reset)
}
