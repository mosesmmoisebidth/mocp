package server

import (
	"fmt"
	"github.com/claudiodangelis/qrcp/style"
	"time"
)

// ShowStartupBanner displays the initial MOCP banner and info
func ShowStartupBanner() {
	fmt.Println(style.MainLogo)
	fmt.Println(style.InfoBox("MOCP Transfer", "Ready to transfer files securely and quickly"))
}

// ShowTransferProgress displays an animated progress bar
func ShowTransferProgress(current, total int64, filename string) {
	prefix := fmt.Sprintf("Transferring %s%s%s", style.BrightYellow, filename, style.Reset)
	// Print a single-line progress bar that updates in place
	fmt.Print(style.AnimatedProgressBar(current, total, prefix))
}

// ShowWaitingStatus shows a spinner while waiting for connection
func ShowWaitingStatus() {
	msg := "Waiting for connection... (Press Ctrl+C to cancel)"
	fmt.Print(style.LoadingSpinner(msg))
}

// ShowTransferComplete shows success message and support info
func ShowTransferComplete(filename string) {
	fmt.Printf("\n%s\n", style.SuccessMessage("Transfer Completed Successfully!"))
	fmt.Printf("%s\n", style.RetroBox(fmt.Sprintf("Successfully transferred: %s", filename)))

	// Show support message occasionally (33% chance)
	if time.Now().UnixNano()%3 == 0 {
		fmt.Println(style.Coffee())
	}
}

// ShowQRCode displays QR code with enhanced styling
func ShowQRCode() {
	fmt.Println(style.InfoBox("QR Code", "Scan with your device to begin transfer"))
}

// ShowError displays an error message with styling
func ShowError(err error) {
	fmt.Println(style.ErrorMessage(err.Error()))
}

// ShowInterfaceSelection shows available network interfaces
func ShowInterfaceSelection(interfaces map[string]string) {
	fmt.Println(style.InfoBox("Network Interfaces", "Choose an interface to use for transfer:"))
	// Print each interface with a unique color based on its name
	for name, ip := range interfaces {
		color := style.ColorForInterface(name)
		fmt.Printf("%s→ %s%s%s (%s%s%s)%s\n",
			style.BrightCyan,
			color, name, style.Reset,
			style.BrightGreen, ip,
			style.Reset,
			style.Reset)
	}
}

// ShowFileInfo displays information about the file being transferred
func ShowFileInfo(filename string, size int64) {
	info := fmt.Sprintf("File: %s\nSize: %s", filename, style.FormatSize(size))
	fmt.Println(style.InfoBox("Transfer Details", info))
}
