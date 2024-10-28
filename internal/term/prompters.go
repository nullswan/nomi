package term

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	hook "github.com/robotn/gohook"
)

func PromptForBool(label string, defaultVal bool) bool {
	items := []string{"Yes", "No"}
	defaultIndex := 0
	if !defaultVal {
		defaultIndex = 1
	}
	prompt := promptui.Select{
		Label:        label,
		Items:        items,
		CursorPos:    defaultIndex,
		HideHelp:     false,
		HideSelected: false,
	}
	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
		os.Exit(1)
	}
	return result == "Yes"
}

func PromptForString(
	label string,
	defaultVal string,
	validate func(string) error,
) string {
	prompt := promptui.Prompt{
		Label:       label,
		Default:     defaultVal,
		Validate:    validate,
		HideEntered: false,
	}
	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
		os.Exit(1)
	}
	return result
}

func PromptSelectString(
	label string,
	items []string,
) string {
	prompt := promptui.Select{
		Label:        label,
		Items:        items,
		CursorPos:    0,
		HideHelp:     false,
		HideSelected: false,
	}
	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
		os.Exit(1)
	}
	return result
}

func PromptForKey(label string) uint16 {
	fmt.Println(label)
	s := hook.Start()

	var keyCode uint16
	for e := range s {
		if e.Kind == hook.KeyDown || e.Kind == hook.KeyHold {
			keyCode = e.Rawcode
			break
		}
	}

	confirmed := PromptForBool(
		fmt.Sprintf(
			"Confirm key code %d?",
			keyCode,
		),
		false,
	)
	if confirmed {
		hook.End()
		return keyCode
	}
	hook.End()
	return PromptForKey(label)
}
