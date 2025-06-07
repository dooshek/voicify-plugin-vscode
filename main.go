package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/dooshek/voicify/pkg/pluginapi"
	"github.com/go-vgo/robotgo"
)

// Logger implementation
type LogLevel int

const (
	// Debug level for detailed information
	LevelDebug LogLevel = iota
	// Info level for general information
	LevelInfo
	// Warn level for warnings
	LevelWarn
	// Error level for errors
	LevelError
)

var (
	currentLevel LogLevel  = LevelInfo
	output       io.Writer = os.Stdout
	mu           sync.Mutex
)

// Logger provides logging functionality for plugins
type Logger struct{}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return &Logger{}
}

// Debug logs a debug message
func (l *Logger) Debug(message string) {
	if currentLevel <= LevelDebug {
		writeLog("DEBUG", message)
	}
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	if currentLevel <= LevelDebug {
		writeLog("DEBUG", fmt.Sprintf(format, args...))
	}
}

// Info logs an info message
func (l *Logger) Info(message string) {
	if currentLevel <= LevelInfo {
		writeLog("INFO", message)
	}
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	if currentLevel <= LevelInfo {
		writeLog("INFO", fmt.Sprintf(format, args...))
	}
}

// Warn logs a warning message
func (l *Logger) Warn(message string) {
	if currentLevel <= LevelWarn {
		writeLog("WARN", message)
	}
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	if currentLevel <= LevelWarn {
		writeLog("WARN", fmt.Sprintf(format, args...))
	}
}

// Error logs an error message with an optional error
func (l *Logger) Error(message string, err error) {
	if currentLevel <= LevelError {
		errMsg := message
		if err != nil {
			errMsg = fmt.Sprintf("%s: %v", message, err)
		}
		writeLog("ERROR", errMsg)
	}
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	if currentLevel <= LevelError {
		writeLog("ERROR", fmt.Sprintf(format, args...))
	}
}

// writeLog writes a log message with the specified level
func writeLog(level, message string) {
	mu.Lock()
	defer mu.Unlock()
	timestamp := time.Now().Format("15:04:05")
	fmt.Fprintf(output, "%s [%s] %s\n", timestamp, level, message)
}

// SetLogLevel sets the global log level for this logger
func SetLogLevel(level LogLevel) {
	mu.Lock()
	defer mu.Unlock()
	currentLevel = level
}

// SetOutput sets the output writer for this logger
func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	output = w
}

// Window implementation
type Window struct {
	Title string
}

// NewWindow creates a new window instance
func NewWindow() *Window {
	return &Window{}
}

// WindowInfo contains information about the focused window
type WindowInfo struct {
	Title   string
	AppName string
}

// GetFocusedWindow gets the currently focused window
func (w *Window) GetFocusedWindow() (*Window, error) {
	// Get window ID
	windowID, err := exec.Command("xdotool", "getactivewindow").Output()
	if err != nil {
		return nil, err
	}

	// Get window name
	windowName, err := exec.Command("xdotool", "getwindowname", strings.TrimSpace(string(windowID))).Output()
	if err != nil {
		return nil, err
	}

	return &Window{
		Title: strings.TrimSpace(string(windowName)),
	}, nil
}

// Clipboard implementation
type Clipboard struct{}

// NewClipboard creates a new clipboard instance
func NewClipboard() *Clipboard {
	return &Clipboard{}
}

// CopyToClipboard copies text to the clipboard
func (c *Clipboard) CopyToClipboard(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		// macOS
		cmd = exec.Command("pbcopy")
	case "linux":
		// Linux
		if _, err := exec.LookPath("xclip"); err != nil {
			return fmt.Errorf("xclip is not installed")
		}
		cmd = exec.Command("xclip", "-selection", "clipboard")
		pipeReader, pipeWriter := io.Pipe()
		cmd.Stdin = pipeReader

		go func() {
			defer pipeWriter.Close()
			pipeWriter.Write([]byte(text))
		}()
	}

	return cmd.Run()
}

// PasteWithReturn pastes text and adds a newline
func (c *Clipboard) PasteWithReturn(text string) error {
	// Copy text to clipboard
	if err := c.CopyToClipboard(text); err != nil {
		return err
	}

	if isX11() {
		// Use robotgo for X11
		robotgo.KeyTap("v", "ctrl")
		robotgo.KeyTap("enter")
	} else {
		// Wayland: Use XWayland compatibility layer (most reliable on Fedora)
		cmd := exec.Command("xdotool", "key", "ctrl+v")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to paste: %v", err)
		}

		// Press Enter
		time.Sleep(50 * time.Millisecond)
		cmd = exec.Command("xdotool", "key", "Return")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to press enter: %v", err)
		}
	}

	return nil
}

// isX11 checks if the current session is running X11
func isX11() bool {
	session := os.Getenv("XDG_SESSION_TYPE")
	return strings.ToLower(session) == "x11"
}

// VSCodePlugin is a plugin for VSCode
type VSCodePlugin struct{}

// Action is the VSCode action
type Action struct {
	transcription string
}

// Initialize initializes the VSCode plugin
func (p *VSCodePlugin) Initialize() error {
	logger.Debug("VSCode plugin initialized")
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
	window := NewWindow()
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

	clipboard := NewClipboard()
	clipboard.PasteWithReturn(transcription)

	return nil
}

// GetMetadata returns metadata about the action
func (a *Action) GetMetadata() pluginapi.ActionMetadata {
	return pluginapi.ActionMetadata{
		Name:        "vscode",
		Description: "wykonanie akcji w edytorze VSCode",
		Priority:    2,
	}
}

// CreatePlugin creates a new instance of the VSCode plugin
// This function is loaded by the plugin manager
func CreatePlugin() pluginapi.VoicifyPlugin {
	return &VSCodePlugin{}
}

// This ensures the function name is in the binary's exported symbols
var _ = CreatePlugin

var logger = NewLogger()
