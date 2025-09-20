package main

import (
	"flag"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/pijalu/printer-patcher/config"
	"github.com/pijalu/printer-patcher/tools"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		return
	}

	// Command line flags
	headlessMode := flag.Bool("headless", false, "Run in headless mode")
	ipAddress := flag.String("ip", "", "Printer IP address (for headless mode)")
	actionName := flag.String("action", "", "Action to execute (for headless mode)")

	flag.Parse()

	// If headless mode, run without GUI
	if *headlessMode {
		runHeadless(*ipAddress, config.Username, config.Password, *actionName)
		return
	}

	// Otherwise, run GUI mode
	runGUI(config)
}

func runHeadless(ipAddress, username, password, actionName string) {
	fmt.Println("Running in headless mode...")

	// Validate required parameters
	if ipAddress == "" {
		fmt.Println("Error: IP address is required for headless mode")
		return
	}

	if actionName == "" {
		fmt.Println("Error: Action name is required for headless mode")
		return
	}

	// Load configuration
	configData, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// Find the requested action
	var action *config.Action
	for _, a := range configData.Actions {
		if a.Title == actionName {
			action = &a
			break
		}
	}

	if action == nil {
		fmt.Printf("Error: Action '%s' not found\n", actionName)
		return
	}

	// Create SSH client
	sshClient := tools.NewSSHConfig(ipAddress, 22, username, password, "")

	// Connect to printer
	fmt.Printf("Connecting to %s...\n", ipAddress)
	err = sshClient.Connect()
	if err != nil {
		fmt.Printf("Error connecting to printer: %v\n", err)
		return
	}
	defer sshClient.Close()

	fmt.Println("Connected successfully!")

	// Execute steps
	fmt.Printf("Executing action: %s\n", action.Title)
	totalSteps := len(action.Steps)
	successfulSteps := 0

	for i, step := range action.Steps {
		fmt.Printf("Step %d/%d: %s\n", i+1, totalSteps, step.Title)
		// Load script content if it's a file path
		scriptContent := step.Script
		if strings.HasSuffix(step.Script, ".sh") {
			// Try to load the script from the script directory
			content, err := config.LoadStep(step.Script)
			if err == nil {
				scriptContent = content
			} else {
				fmt.Printf("Warning: Could not load script '%s': %v\n", step.Script, err)
			}
		}

		// Execute step
		result, err := sshClient.Execute(scriptContent)
		if err != nil {
			fmt.Printf("Error in step '%s': %v\n", step.Title, err)
			break
		}

		// Validate output
		if tools.ValidateOutput(strings.TrimSpace(result), step.Expected) {
			fmt.Printf("✓ Step '%s' completed successfully\n", step.Title)
			successfulSteps++
		} else {
			fmt.Printf("✗ Step '%s' failed validation. Expected: '%s', Got: '%s'\n",
				step.Title, step.Expected, strings.TrimSpace(result))
			break
		}
	}

	// Show final result
	if successfulSteps == totalSteps {
		fmt.Printf("Success: All %d steps completed successfully!\n", totalSteps)
	} else {
		fmt.Printf("Error: Only %d/%d steps completed successfully\n", successfulSteps, totalSteps)
	}
}

func runGUI(config *config.Config) {
	myApp := app.New()
	myWindow := myApp.NewWindow("3D Printer Patcher")
	myWindow.SetContent(createScreen(myWindow, config))
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.CenterOnScreen()
	myWindow.ShowAndRun()
}

func createScreen(window fyne.Window, config *config.Config) fyne.CanvasObject {
	selectedAction := -1

	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("Enter printer IP address...")

	executeButton := widget.NewButton("Execute Selected Action", func() {
		showActionExecutionDialog(window, ipEntry.Text, config, selectedAction)
	})
	executeButton.Disable()

	// Track execute button
	ipEntry.OnChanged = func(value string) {
		if len(value) > 0 && selectedAction != -1 {
			executeButton.Enable()
		} else {
			executeButton.Disable()
		}
	}

	actionList := widget.NewList(
		func() int {
			if config != nil {
				return len(config.Actions)
			}
			return 0
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if config != nil && id >= 0 && id < len(config.Actions) {
				item.(*widget.Label).SetText(fmt.Sprintf("%s: %s", config.Actions[id].Title, config.Actions[id].Description))
			}
		},
	)

	actionList.OnSelected = func(id widget.ListItemID) {
		selectedAction = int(id)
		if ipEntry.Text != "" {
			executeButton.Enable()
		}
	}
	actionList.Select(widget.ListItemID(0))

	// Create bottom buttons - only show Quit button on non-mobile platforms
	bottomButtons := container.NewVBox(executeButton)

	// Only add Quit button on non-mobile platforms
	if !fyne.CurrentDevice().IsMobile() {
		bottomButtons.Add(widget.NewButton("Quit", func() {
			window.Close()
		}))
	}

	content := container.NewBorder(
		/* top */
		container.NewVBox(
			widget.NewLabel("Enter your printer's connection details:"),
			ipEntry,
		),
		/* Bottom */
		bottomButtons,
		/* left */ nil,
		/* right */ nil,
		/* Content */
		container.NewBorder(
			container.NewVBox(
				widget.NewLabelWithStyle("Actions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewSeparator(),
			),
			nil,
			nil,
			nil,
			actionList,
		),
	)

	// Use the full container instead of centering
	return container.NewVScroll(content)
}

func showActionExecutionDialog(window fyne.Window, ip string, cfg *config.Config, selectedAction int) {
	if cfg == nil || selectedAction < 0 || selectedAction >= len(cfg.Actions) {
		dialog.ShowError(fmt.Errorf("no action selected"), window)
		return
	}
	action := cfg.Actions[selectedAction]

	output := widget.NewLabel("")
	output.SetText(fmt.Sprintf("Executing: %s\n", action.Title))

	stepProgress := widget.NewProgressBar()
	stepStatus := widget.NewLabel("Ready to execute...")
	stepStatus.Alignment = fyne.TextAlignCenter

	stepProgress.Show()
	stepStatus.Show()
	stepProgress.SetValue(0)

	totalSteps := len(action.Steps)
	successfulSteps := 0

	execFn := func() {
		fyne.Do(func() {
			stepStatus.SetText(fmt.Sprintf("Connecting to  %s", ip))
		})
		sshClient := tools.NewSSHConfig(ip, 22, cfg.Username, cfg.Password, "")
		if err := sshClient.Connect(); err != nil {
			fyne.DoAndWait(func() {
				dialog.ShowError(fmt.Errorf("cannot connected to printer: %v", err), window)
			})
			return
		}
		defer sshClient.Close()
		fyne.Do(func() {
			stepStatus.SetText(fmt.Sprintf("Connected to  %s", ip))
		})

		for i, step := range action.Steps {
			progress := float64(i) / float64(totalSteps)
			scriptContent := step.Script
			if strings.HasSuffix(step.Script, ".sh") {
				content, err := config.LoadStep(step.Script)
				if err == nil {
					scriptContent = content
				} else {
					fyne.Do(func() {
						output.SetText(output.Text + fmt.Sprintf("\n⚠ Warning: Could not load script '%s': %v", step.Script, err))
					})
				}
			}

			// Update screen
			fyne.Do(func() {
				stepProgress.SetValue(progress)
				stepStatus.SetText(fmt.Sprintf("Step %d/%d: %s", i+1, totalSteps, step.Title))
				output.SetText(output.Text + fmt.Sprintf("\n▶ Executing step: %s", step.Title))
			})

			result, err := sshClient.Execute(scriptContent)

			if err != nil {
				fyne.Do(func() {
					output.SetText(output.Text + fmt.Sprintf("\n📋 Output:\n%s", result))
					output.SetText(output.Text + fmt.Sprintf("\n❌ Error in step '%s': %v", step.Title, err))
				})
				break
			}

			if tools.ValidateOutput(strings.TrimSpace(result), step.Expected) {
				fyne.Do(func() {
					output.SetText(output.Text + fmt.Sprintf("\n✓ Step completed successfully"))
				})
				successfulSteps++
			} else {
				fyne.Do(func() {
					if result != "" {
						output.SetText(output.Text + fmt.Sprintf("\n📋 Output:\n%s", result))
					}
					output.SetText(output.Text + fmt.Sprintf("\n✗ Step failed validation. Expected: '%s', Got: '%s'",
						step.Expected, strings.TrimSpace(result)))
				})
				break
			}
		}

		fyne.Do(func() {
			stepProgress.SetValue(1.0)
			stepStatus.SetText(fmt.Sprintf("Completed: %d/%d steps successful", successfulSteps, totalSteps))

			if successfulSteps == totalSteps {
				dialog.ShowInformation("Success", fmt.Sprintf("All %d steps completed successfully!", totalSteps), window)
			} else {
				dialog.ShowError(fmt.Errorf("only %d/%d steps completed successfully", successfulSteps, totalSteps), window)
			}
		})
	}

	dialogContent := container.NewBorder(
		container.NewVBox(
			stepStatus,
			stepProgress,
		),
		nil,
		nil,
		nil,
		container.NewScroll(output),
	)

	executionDialog := dialog.NewCustom(fmt.Sprintf("Action: %s", action.Title), "Close", dialogContent, window)
	executionDialog.Resize(fyne.NewSize(800, 600))
	executionDialog.Show()

	// Run in it's own thread
	go execFn()
}
