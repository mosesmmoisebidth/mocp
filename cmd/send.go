package cmd

import (
	"fmt"
	"os"

	"github.com/claudiodangelis/qrcp/body"
	"github.com/claudiodangelis/qrcp/config"
	"github.com/claudiodangelis/qrcp/logger"
	"github.com/claudiodangelis/qrcp/qr"
	"github.com/eiannone/keyboard"

	"github.com/claudiodangelis/qrcp/server"
	"github.com/spf13/cobra"
)

func sendCmdFunc(command *cobra.Command, args []string) error {
	log := logger.New(app.Flags.Quiet)
	server.ShowStartupBanner()
	payload, err := body.FromArgs(args, app.Flags.Zip)
	if err != nil {
		server.ShowError(err)
		return err
	}
	// Determine size
	fi, err := os.Stat(payload.Path)
	if err != nil {
		server.ShowError(err)
		return err
	}
	server.ShowFileInfo(payload.Filename, fi.Size())
	// Initialize config and get interface selection
	cfg := config.New(app)

	// Choose interface before starting server
	iface, err := config.ChooseInterface(app.Flags)
	if err != nil {
		server.ShowError(err)
		return err
	}
	cfg.Interface = iface

	srv, err := server.New(&cfg)
	if err != nil {
		server.ShowError(err)
		return err
	}

	// Sets the body
	srv.Send(payload)
	server.ShowQRCode()
	qr.RenderStringWithSideOverwrite(srv.SendURL, cfg.Reversed, nil)
	// Reserve a line below the QR area for in-place progress updates
	fmt.Println()
	if app.Flags.Browser {
		srv.DisplayQR(srv.SendURL)
	}
	if err := keyboard.Open(); err == nil {
		defer func() {
			keyboard.Close()
		}()
		go func() {
			for {
				char, key, _ := keyboard.GetKey()
				if string(char) == "q" || key == keyboard.KeyCtrlC {
					srv.Shutdown()
				}
			}
		}()
	} else {
		log.Print(fmt.Sprintf("Warning: keyboard not detected: %v", err))
	}
	if err := srv.Wait(); err != nil {
		return err
	}
	return nil
}

var sendCmd = &cobra.Command{
	Use:     "transfer",
	Short:   "Transfer a file(s) or directories from this host",
	Long:    "Transfer a file(s) or directories from this host",
	Aliases: []string{"send", "s"},
	Example: `# Transfer /path/file.gif. Webserver listens on a random port
mocp transfer /path/file.gif
# Shorter version:
mocp /path/file.gif
# Zip file1.gif and file2.gif, then transfer the zip package
mocp /path/file1.gif /path/file2.gif
# Zip the content of directory, then transfer the zip package
mocp /path/directory
# Transfer file.gif by creating a webserver on port 8080
mocp --port 8080 /path/file.gif
`,
	Args: cobra.MinimumNArgs(1),
	RunE: sendCmdFunc,
}
