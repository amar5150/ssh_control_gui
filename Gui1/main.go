package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/crypto/ssh"
)

// SSHClient represents the SSH connection details.
type SSHClient struct {
	Host     string
	Port     string
	User     string
	Password string
}

// RunCommand executes a command over SSH and returns the output.
func (c *SSHClient) RunCommand(command string) (string, error) {
	config := &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", c.Host, c.Port), config)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdout bytes.Buffer
	session.Stdout = &stdout

	if err := session.Run(command); err != nil {
		return "", err
	}

	return stdout.String(), nil
}

func main() {
	// Create GUI application
	myApp := app.New()
	myWindow := myApp.NewWindow("SSH & JSON Utility")

	// SSH Script Runner Panel
	sshClient := &SSHClient{}

	hostEntry := widget.NewEntry()
	hostEntry.SetPlaceHolder("Enter Host (e.g., 192.168.1.10)")

	portEntry := widget.NewEntry()
	portEntry.SetText("22")

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("Enter Username")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Enter Password")

	scripts := []string{"ls -l", "uname -a", "df -h"}
	scriptSelect := widget.NewSelect(scripts, nil)

	// Dynamic arguments container
	argumentsContainer := container.NewVBox()
	argumentEntries := []*widget.Entry{}

	// Function to add a new argument field
	addArgumentField := func() {
		argIndex := len(argumentEntries) + 1
		label := widget.NewLabel(fmt.Sprintf("Argument %d:", argIndex))
		entry := widget.NewEntry()
		entry.SetPlaceHolder(fmt.Sprintf("Enter Argument %d", argIndex))

		argumentsContainer.Add(container.NewHBox(label, entry))
		argumentEntries = append(argumentEntries, entry)
	}

	// Add the first argument field by default
	addArgumentField()

	addArgumentButton := widget.NewButton("Add Argument", func() {
		addArgumentField()
	})

	output := widget.NewMultiLineEntry()
	output.SetPlaceHolder("Command output will appear here...")
	output.Disable()

	runButton := widget.NewButton("Run Script", func() {
		sshClient.Host = hostEntry.Text
		sshClient.Port = portEntry.Text
		sshClient.User = userEntry.Text
		sshClient.Password = passwordEntry.Text

		selectedScript := scriptSelect.Selected
		if selectedScript == "" {
			output.SetText("Please select a script to run.")
			return
		}

		// Collect all arguments
		var arguments []string
		for _, entry := range argumentEntries {
			if entry.Text != "" {
				arguments = append(arguments, entry.Text)
			}
		}

		finalCommand := selectedScript
		if len(arguments) > 0 {
			finalCommand = fmt.Sprintf("%s %s", selectedScript, strings.Join(arguments, " "))
		}

		result, err := sshClient.RunCommand(finalCommand)
		if err != nil {
			output.SetText(fmt.Sprintf("Error: %s", err))
			return
		}

		output.SetText(result)
	})

	sshPanel := container.NewVBox(
		widget.NewLabel("SSH Connection Details:"),
		widget.NewForm(
			widget.NewFormItem("Host", hostEntry),
			widget.NewFormItem("Port", portEntry),
			widget.NewFormItem("Username", userEntry),
			widget.NewFormItem("Password", passwordEntry),
		),
		widget.NewLabel("Select Script to Run:"),
		scriptSelect,
		widget.NewLabel("Enter Arguments:"),
		argumentsContainer,
		addArgumentButton,
		runButton,
		output,
	)

	// JSON Generator Panel (unchanged)
	jsonEntries := make(map[string]*widget.Entry)
	fieldForm := container.NewVBox()

	addFieldButton := widget.NewButton("Add Field", func() {
		fieldKeyEntry := widget.NewEntry()
		fieldKeyEntry.SetPlaceHolder("Field Name")

		fieldValueEntry := widget.NewEntry()
		fieldValueEntry.SetPlaceHolder("Field Value")

		// Add to form and update entries map
		fieldForm.Add(container.NewHBox(fieldKeyEntry, fieldValueEntry))
		jsonEntries[fieldKeyEntry.Text] = fieldValueEntry
	})

	generateJSONButton := widget.NewButton("Generate JSON", func() {
		data := make(map[string]string)
		for key, entry := range jsonEntries {
			if key == "" || entry.Text == "" {
				continue // Skip empty fields
			}
			data[key] = entry.Text
		}

		file, err := os.Create("output.json")
		if err != nil {
			log.Printf("Failed to create JSON file: %v\n", err)
			return
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(data); err != nil {
			log.Printf("Failed to encode JSON: %v\n", err)
			return
		}

		log.Println("JSON file 'output.json' generated successfully.")
	})

	jsonPanel := container.NewVBox(
		widget.NewLabel("JSON Generator"),
		widget.NewLabel("Add Fields to Generate a JSON File:"),
		fieldForm,
		addFieldButton,
		generateJSONButton,
	)

	// Tab container
	tabs := container.NewAppTabs(
		container.NewTabItem("SSH Script Runner", sshPanel),
		container.NewTabItem("JSON Generator", jsonPanel),
	)

	myWindow.SetContent(tabs)
	myWindow.Resize(fyne.NewSize(500, 500))
	myWindow.ShowAndRun()
}
