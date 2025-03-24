package main

import (
	"strings"

	"github.com/dooshek/voicify/pkg/pluginapi"
)

var (
	logger    = pluginapi.NewLogger()
	clipboard = pluginapi.NewClipboard()
	window    = pluginapi.NewWindow()
)

// VSCodePlugin is a plugin for VSCode
type VSCodePlugin struct{}

// Action is the VSCode action
type Action struct {
	transcription string
}

// Initialize initializes the VSCode plugin
func (p *VSCodePlugin) Initialize() error {
	logger.Info("VSCode plugin initialized")
	return nil
}

// GetMetadata returns metadata about the plugin
func (p *VSCodePlugin) GetMetadata() pluginapi.PluginMetadata {
	return pluginapi.PluginMetadata{
		Name:        "vscode",
		Version:     "1.0.0",
		Description: "Plugin for Visual Studio Code",
		Author:      "Voicify Team",
	}
}

// GetActions returns a list of actions provided by this plugin
func (p *VSCodePlugin) GetActions(transcription string) []pluginapi.PluginAction {
	return []pluginapi.PluginAction{
		&Action{transcription: transcription},
	}
}

// Execute executes the VSCode action
func (a *Action) Execute(transcription string) error {
	logger.Debugf("Checking if VSCode should execute action for transcription: %s", transcription)
	focusedWindow, err := window.GetFocusedWindow()
	if err != nil {
		logger.Error("Error getting focused window", err)
		return err
	}

	logger.Debugf("Checking window title: %s", focusedWindow.Title)
	if !strings.Contains(focusedWindow.Title, "VSC") {
		logger.Debug("VSCode is not open, skipping action")
		return nil
	}

	clipboard.PasteWithReturn(transcription)

	return nil
}

// GetMetadata returns metadata about the action
func (a *Action) GetMetadata() pluginapi.ActionMetadata {
	return pluginapi.ActionMetadata{
		Name:        "vscode",
		Description: "wykonanie akcji w edytorze VSCode",
		LLMCommands: &[]string{
			"vscode",
			"edytor",
			"visual studio code",
		},
		Priority: 2,
	}
}

// CreatePlugin creates a new instance of the VSCode plugin
// This function is loaded by the plugin manager
func CreatePlugin() pluginapi.VoicifyPlugin {
	return &VSCodePlugin{}
}

// This is required for Go plugins
var (
	// Export the plugin creation function
	_ = CreatePlugin
)
