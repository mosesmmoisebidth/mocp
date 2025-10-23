package qr

import (
	"fmt"
	"image"
	"log"
	"strings"

	"github.com/claudiodangelis/qrcp/style"
	"github.com/skip2/go-qrcode"
)

// RenderString as a QR code, with optional sideContent (e.g., progress bar)
// RenderStringWithSide prints a QR with optional side content.
func RenderStringWithSide(s string, inverseColor bool, sideContent []string) {
	q, err := qrcode.New(s, qrcode.Medium)
	if err != nil {
		log.Fatal(err)
	}

	qrString := q.ToSmallString(inverseColor)
	lines := strings.Split(strings.TrimSpace(qrString), "\n")
	width := len(lines[0])
	maxLines := len(lines)
	if len(sideContent) > maxLines {
		maxLines = len(sideContent)
	}

	// Print top border
	fmt.Print(style.BrightCyan + "╔")
	fmt.Print(strings.Repeat("═", width+2))
	fmt.Println("╗" + style.Reset)

	// Print QR code with side content
	for i := 0; i < maxLines; i++ {
		qrPart := ""
		if i < len(lines) && len(lines[i]) > 0 {
			qrPart = fmt.Sprintf("%s║ %s%s%s ║%s",
				style.BrightCyan,
				style.BrightWhite,
				lines[i],
				style.BrightCyan,
				style.Reset)
		} else {
			qrPart = fmt.Sprintf("%s║ %s%s ║%s",
				style.BrightCyan,
				strings.Repeat(" ", width),
				style.BrightCyan,
				style.Reset)
		}
		side := ""
		if i < len(sideContent) {
			side = "  " + sideContent[i]
		}
		fmt.Println(qrPart + side)
	}

	// Print bottom border
	fmt.Print(style.BrightCyan + "╚")
	fmt.Print(strings.Repeat("═", width+2))
	fmt.Println("╝" + style.Reset)

	// Print centered instruction
	fmt.Printf("\n%s📱 Scan this QR code with your device%s\n",
		style.BrightGreen, style.Reset)

}

// Backwards compatible
func RenderString(s string, inverseColor bool) {
	RenderStringWithSide(s, inverseColor, nil)
}

// RenderStringWithSideOverwrite moves the cursor up to the previous printed QR
// (if any) and reprints a new QR+side content in its place. This allows in-place
// updates without flooding the terminal.
func RenderStringWithSideOverwrite(s string, inverseColor bool, sideContent []string) {
	// For safety, do not attempt to move the cursor up here — callers should
	// print the QR in-place once. This function now behaves the same as
	// RenderStringWithSide to avoid overwriting unrelated terminal content.
	RenderStringWithSide(s, inverseColor, sideContent)
}

// UpdateProgressLine overwrites the single line immediately below the QR
// (the line reserved by printing a blank line after the QR). It safely
// clears the line and prints the provided progress text.
func UpdateProgressLine(progress string) {
	// Move cursor up 1 line, clear it, print progress, and leave cursor after line
	// [1A = move cursor up 1, [2K = erase entire line
	fmt.Printf("\033[1A\033[2K%s\n", progress)
}

// RenderImage returns a QR code as an image.Image
func RenderImage(s string) image.Image {
	q, err := qrcode.New(s, qrcode.Medium)
	if err != nil {
		log.Fatal(err)
	}
	return q.Image(256)
}
