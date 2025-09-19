package main

import (
	"flag"
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/pijalu/printer-patcher/config"
	"github.com/pijalu/printer-patcher/github"
	"github.com/pijalu/printer-patcher/tools"
)

// SourceCacheEntry represents a cached source list entry
type SourceCacheEntry struct {
	Sources   []string
	Timestamp time.Time
}

// ScriptCacheEntry represents a cached script entry
type ScriptCacheEntry struct {
	Content   string
	Timestamp time.Time
}

// Global cache for source lists and scripts
var (
	sourceCache = make(map[string]SourceCacheEntry)
	scriptCache = make(map[string]ScriptCacheEntry)
	cacheMutex  sync.RWMutex
	cacheTTL    = 24 * time.Hour
)

func main() {
	// Create local config provider by default
	configProvider := config.NewLocalConfigProvider()
	_, err := configProvider.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		return
	}

	// Command line flags
	headlessMode := flag.Bool("headless", false, "Run in headless mode")
	ipAddress := flag.String("ip", "", "Printer IP address (for headless mode)")
	actionName := flag.String("action", "", "Action to execute (for headless mode)")
	source := flag.String("source", "local", "Source for actions (local, main, or release tag)")

	flag.Parse()

	// If headless mode, run without GUI
	if *headlessMode {
		// Create the appropriate config provider based on source
		var provider config.ConfigProvider
		if *source == "local" {
			provider = config.NewLocalConfigProvider()
		} else {
			// Check if the source is in the format [owner/name] branch
			var owner, name, branch string
			if _, err := fmt.Sscanf(*source, "[%s/%s] %s", &owner, &name, &branch); err == nil {
				// Remove the trailing ']' from the name
				name = strings.TrimSuffix(name, "]")
				fmt.Printf("Creating provider for repo %s/%s, branch %s\n", owner, name, branch)
				provider, err = config.NewRemoteConfigProviderWithRepo(branch, owner, name)
			} else {
				// Fallback to default repository
				fmt.Printf("Creating provider for default repo, branch %s\n", *source)
				provider, err = config.NewRemoteConfigProvider(*source)
			}

			if err != nil {
				fmt.Printf("Error creating remote provider: %v\n", err)
				return
			}
		}
		runHeadless(*ipAddress, provider, *actionName)
		return
	}

	// Otherwise, run GUI mode
	runGUI(configProvider)
}

// getCachedSources retrieves sources from cache if available and not expired
func getCachedSources() ([]string, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	if entry, exists := sourceCache["sources"]; exists {
		if time.Since(entry.Timestamp) < cacheTTL {
			return entry.Sources, true
		}
	}
	return nil, false
}

// setCachedSources stores sources in cache
func setCachedSources(sources []string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	sourceCache["sources"] = SourceCacheEntry{
		Sources:   sources,
		Timestamp: time.Now(),
	}
}

// getCachedScript retrieves a script from cache if available and not expired
func getCachedScript(scriptPath string) (string, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	if entry, exists := scriptCache[scriptPath]; exists {
		if time.Since(entry.Timestamp) < cacheTTL {
			return entry.Content, true
		}
	}
	return "", false
}

// setCachedScript stores a script in cache
func setCachedScript(scriptPath, content string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	scriptCache[scriptPath] = ScriptCacheEntry{
		Content:   content,
		Timestamp: time.Now(),
	}
}

// fetchSources fetches sources from GitHub APIs
func fetchSources() ([]string, error) {
	// Get repository configuration
	repoConfig, err := config.GetRepoConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading repo config: %v", err)
	}

	var allSources []string
	allSources = append(allSources, "local") // Always include local as an option

	// For each repository, fetch branches and releases
	for _, repo := range repoConfig.Repositories {
		fmt.Printf("Fetching sources for repository: %s/%s\n", repo.Owner, repo.Name)
		client := github.NewClientWithRepo(repo.Owner, repo.Name)
		branches, err := client.GetBranchNames()
		if err != nil {
			fmt.Printf("Error fetching branches for %s/%s: %v\n", repo.Owner, repo.Name, err)
			continue
		}

		// Add repository identifier to each branch/release name
		for _, branch := range branches {
			if branch != "local" { // Skip local as it's already added
				sourceName := fmt.Sprintf("[%s/%s] %s", repo.Owner, repo.Name, branch)
				allSources = append(allSources, sourceName)
			}
		}
	}

	fmt.Printf("All available sources: %v\n", allSources)
	return allSources, nil
}

// loadSources loads sources either from cache or by fetching from APIs
func loadSources(callback func([]string, error)) {
	// Check cache first
	if sources, cached := getCachedSources(); cached {
		fmt.Println("Using cached sources")
		go callback(sources, nil)
		return
	}

	// Fetch from APIs
	go func() {
		sources, err := fetchSources()
		if err != nil {
			callback(nil, err)
			return
		}

		// Cache the results
		setCachedSources(sources)
		callback(sources, nil)
	}()
}

// preloadActionScripts downloads and caches all scripts for an action
func preloadActionScripts(configProvider config.ConfigProvider, action *config.Action) error {
	fmt.Printf("Preloading %d scripts for action '%s'\n", len(action.Steps), action.Title)

	scriptCount := 0
	for _, step := range action.Steps {
		if strings.HasSuffix(step.Script, ".sh") {
			scriptCount++
		}
	}

	if scriptCount == 0 {
		fmt.Println("No scripts to preload")
		return nil
	}

	preloadedCount := 0
	failedCount := 0

	for _, step := range action.Steps {
		if strings.HasSuffix(step.Script, ".sh") {
			fmt.Printf("Preloading script %d/%d: %s\n", preloadedCount+failedCount+1, scriptCount, step.Script)

			// Check if already cached
			if _, cached := getCachedScript(step.Script); cached {
				fmt.Printf("Script %s already cached\n", step.Script)
				preloadedCount++
				continue
			}

			// Load the script
			content, err := configProvider.LoadStep(step.Script)
			if err != nil {
				fmt.Printf("Warning: Could not preload script '%s': %v\n", step.Script, err)
				failedCount++
				continue
			}

			// Cache the script
			setCachedScript(step.Script, content)
			fmt.Printf("Script %s cached successfully\n", step.Script)
			preloadedCount++
		}
	}

	fmt.Printf("Preloaded %d/%d scripts, %d failed\n", preloadedCount, scriptCount, failedCount)

	// If all scripts failed to preload, return an error
	if scriptCount > 0 && preloadedCount == 0 {
		return fmt.Errorf("failed to preload any scripts")
	}

	return nil
}

func runHeadless(ipAddress string, configProvider config.ConfigProvider, actionName string) {
	fmt.Println("Running in headless mode...")
	fmt.Printf("Using source: %s\n", configProvider.GetSourceName())

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
	configData, err := configProvider.LoadConfig()
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

	// Preload all scripts for the action
	fmt.Println("Preloading scripts...")
	err = preloadActionScripts(configProvider, action)
	if err != nil {
		fmt.Printf("Error: Failed to preload scripts: %v\n", err)
		// Continue anyway as we can try to load them during execution
	}

	// Create SSH client
	sshClient := tools.NewSSHConfig(ipAddress, 22, configData.Username, configData.Password, "")

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
			// Try to get from cache first
			if cachedContent, cached := getCachedScript(step.Script); cached {
				scriptContent = cachedContent
				fmt.Printf("Using cached script: %s\n", step.Script)
			} else {
				// Load the script from the script directory
				content, err := configProvider.LoadStep(step.Script)
				if err == nil {
					scriptContent = content
					// Cache for future use
					setCachedScript(step.Script, content)
				} else {
					fmt.Printf("Error: Could not load script '%s': %v\n", step.Script, err)
					fmt.Println("Continuing with original script path...")
					// Continue with original script path - SSH client might be able to handle it
				}
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

func runGUI(configProvider config.ConfigProvider) {
	myApp := app.New()
	myWindow := myApp.NewWindow("3D Printer Patcher")
	myWindow.SetContent(createScreen(myWindow, configProvider))
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.CenterOnScreen()
	myWindow.ShowAndRun()
}

func createScreen(window fyne.Window, configProvider config.ConfigProvider) fyne.CanvasObject {
	selectedAction := -1
	currentConfig, err := configProvider.LoadConfig()
	if err != nil {
		// Instead of returning nil, create an error screen
		errorLabel := widget.NewLabel(fmt.Sprintf("Failed to load configuration: %v", err))
		errorLabel.Wrapping = fyne.TextWrapWord
		errorLabel.Alignment = fyne.TextAlignCenter

		reloadButton := widget.NewButton("Retry", func() {
			window.SetContent(createScreen(window, configProvider))
		})

		return container.NewCenter(container.NewVBox(
			widget.NewLabelWithStyle("Configuration Error", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			errorLabel,
			reloadButton,
		))
	}

	// Store the current source to avoid redundant updates
	// For remote providers, we need to construct the full name
	currentSource := configProvider.GetSourceName()
	if currentSource != "local" {
		// Check if it's a remote provider that has repository info
		if remoteProvider, ok := configProvider.(*config.RemoteConfigProvider); ok {
			currentSource = fmt.Sprintf("[%s] %s", remoteProvider.GetRepoIdentifier(), remoteProvider.GetSourceName())
		}
	}

	// Create source selection dropdown
	sourceSelect := widget.NewSelect([]string{"local"}, func(value string) {
		fmt.Printf("Source selected: %s\n", value)
		// Only reload if the source actually changed
		if value == currentSource {
			fmt.Printf("Source unchanged, skipping reload\n")
			return
		}

		// Handle source change - create new provider and refresh action list
		var newProvider config.ConfigProvider
		var err error
		if value == "local" {
			newProvider = config.NewLocalConfigProvider()
		} else {
			// Parse the source format: [owner/name] branch
			// Extract owner, name, and branch from the value
			var owner, name, branch string
			parts := strings.SplitN(value, "] ", 2)
			if len(parts) == 2 {
				// Remove the leading '[' from the first part
				repoPart := strings.TrimPrefix(parts[0], "[")
				repoParts := strings.SplitN(repoPart, "/", 2)
				if len(repoParts) == 2 {
					owner = repoParts[0]
					name = repoParts[1]
					branch = parts[1]
					newProvider, err = config.NewRemoteConfigProviderWithRepo(branch, owner, name)
				} else {
					// Fallback to default repository
					newProvider, err = config.NewRemoteConfigProvider(value)
				}
			} else {
				// Fallback to default repository
				newProvider, err = config.NewRemoteConfigProvider(value)
			}

			if err != nil {
				dialog.ShowError(fmt.Errorf("failed to create remote provider for %s: %v", value, err), window)
				return
			}
		}

		// Reload the screen with the new provider
		window.SetContent(createScreen(window, newProvider))
	})

	// Create refresh button
	refreshButton := widget.NewButton("Refresh Sources", func() {
		// Clear cache and reload sources
		cacheMutex.Lock()
		delete(sourceCache, "sources")
		cacheMutex.Unlock()

		loadSources(func(sources []string, err error) {
			if err != nil {
				dialog.ShowError(fmt.Errorf("failed to load sources: %v", err), window)
				return
			}

			fyne.Do(func() {
				sourceSelect.Options = sources
				sourceSelect.Refresh()
				// Try to keep the current selection if it's still valid
				found := false
				for _, option := range sources {
					if option == currentSource {
						sourceSelect.SetSelected(currentSource)
						found = true
						break
					}
				}
				// If current source is not in the new list, select the first option
				if !found && len(sources) > 0 {
					sourceSelect.SetSelected(sources[0])
				}
			})
		})
	})

	// Populate source options with GitHub branches/releases from all repositories
	loadSources(func(sources []string, err error) {
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to load sources: %v", err), window)
			return
		}

		fyne.Do(func() {
			sourceSelect.Options = sources
			sourceSelect.Refresh()
			// Make sure the current selection is still selected
			found := false
			for _, option := range sources {
				if option == currentSource {
					sourceSelect.SetSelected(currentSource)
					found = true
					break
				}
			}
			// If current source is not in the new list, select the first option
			if !found && len(sources) > 0 {
				sourceSelect.SetSelected(sources[0])
			}
		})
	})

	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("Enter printer IP address...")

	executeButton := widget.NewButton("Execute Selected Action", func() {
		showActionExecutionDialog(window, ipEntry.Text, currentConfig, selectedAction, configProvider)
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
			if currentConfig != nil {
				return len(currentConfig.Actions)
			}
			return 0
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if currentConfig != nil && id >= 0 && id < len(currentConfig.Actions) {
				item.(*widget.Label).SetText(fmt.Sprintf("%s: %s", currentConfig.Actions[id].Title, currentConfig.Actions[id].Description))
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

	content := container.NewBorder(
		/* top */
		container.NewVBox(
			widget.NewLabel("Enter your printer's connection details:"),
			ipEntry,
			widget.NewSeparator(),
			widget.NewLabel("Select source for actions:"),
			container.NewBorder(nil, nil, nil, refreshButton, sourceSelect),
		),
		/* Bottom */
		container.NewVBox(
			executeButton,
			widget.NewButton("Quit", func() {
				window.Close()
			}),
		),
		/* left */ nil,
		/* right */ nil,
		/* Content */
		container.NewVBox( // Top content
			widget.NewLabelWithStyle("Actions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			actionList,
		),
	)

	// Use the full container instead of centering
	return container.NewVScroll(content)
}

func showActionExecutionDialog(window fyne.Window, ip string, cfg *config.Config, selectedAction int, configProvider config.ConfigProvider) {
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
		// Preload all scripts for the action
		fyne.Do(func() {
			stepStatus.SetText("Preloading scripts...")
		})

		err := preloadActionScripts(configProvider, &cfg.Actions[selectedAction])
		if err != nil {
			fyne.Do(func() {
				output.SetText(output.Text + fmt.Sprintf("\n⚠ Error: Failed to preload scripts: %v", err))
				// Continue anyway as we can try to load them during execution
			})
		}

		fyne.Do(func() {
			stepStatus.SetText(fmt.Sprintf("Connecting to  %s", ip))
		})
		sshClient := tools.NewSSHConfig(ip, 22, cfg.Username, cfg.Password, "")
		if err := sshClient.Connect(); err != nil {
			fyne.DoAndWait(func() {
				dialog.ShowError(fmt.Errorf("cannot connect to printer: %v", err), window)
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
				// Try to get from cache first
				if cachedContent, cached := getCachedScript(step.Script); cached {
					scriptContent = cachedContent
					fyne.Do(func() {
						output.SetText(output.Text + fmt.Sprintf("\n📦 Using cached script: %s", step.Script))
					})
				} else {
					// Load from provider if not cached
					content, err := configProvider.LoadStep(step.Script)
					if err == nil {
						scriptContent = content
						// Cache for future use
						setCachedScript(step.Script, content)
					} else {
						fyne.Do(func() {
							output.SetText(output.Text + fmt.Sprintf("\n⚠ Error: Could not load script '%s': %v", step.Script, err))
							// Continue with original script path - SSH client might be able to handle it
						})
					}
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
