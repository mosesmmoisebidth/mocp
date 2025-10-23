package config

import (
	"errors"
	"fmt"

	"github.com/claudiodangelis/qrcp/application"
	"github.com/claudiodangelis/qrcp/style"
	"github.com/claudiodangelis/qrcp/util"
	"github.com/manifoldco/promptui"
)

// ChooseInterface prompts user to select a network interface
func ChooseInterface(flags application.Flags) (string, error) {
	interfaces, err := util.Interfaces(flags.ListAllInterfaces)
	if err != nil {
		return "", err
	}
	if len(interfaces) == 0 {
		return "", errors.New("no interfaces found")
	}
	if len(interfaces) == 1 && !interactive {
		for name := range interfaces {
			fmt.Printf("%s💻 Only one interface found: %s%s %s%s\n",
				style.BrightCyan, style.BrightYellow, name,
				"(using this one)", style.Reset)
			return name, nil
		}
	}
	// Map for pretty printing
	m := make(map[string]string)
	items := []string{}
	for name, ip := range interfaces {
		label := fmt.Sprintf("%s (%s)", name, ip)
		m[label] = name
		items = append(items, label)
	}
	// Add the "any" interface
	anyIP := "0.0.0.0"
	anyName := "any"
	anyLabel := fmt.Sprintf("%s (%s)", anyName, anyIP)
	m[anyLabel] = anyName
	items = append(items, anyLabel)
	// Print a colored list of options first so the user sees them immediately
	fmt.Println(style.BrightCyan + "Available network interfaces:" + style.Reset)
	for idx, it := range items {
		fmt.Printf("  %s%2d.%s %s\n", style.BrightYellow, idx+1, style.Reset, it)
	}

	prompt := promptui.Select{
		Items: items,
		Label: "Choose interface",
	}
	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return m[result], nil
}
