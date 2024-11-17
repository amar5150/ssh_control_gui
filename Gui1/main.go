package main

import (
	"bytes"
	"fmt"

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
	myWindow := myApp.NewWindow("SSH Script Runner")

	// SSH client details
	sshClient := &SSHClient{}

	// Input fields
	hostEntry := widget.NewEntry()
	hostEntry.SetPlaceHolder("Enter Host (e.g., 192.168.1.10)")

	portEntry := widget.NewEntry()
	portEntry.SetText("22")

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("Enter Username")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Enter Password")

	// Script dropdown
	scripts := []string{"ls -l", "uname -a", "df -h"}
	scriptSelect := widget.NewSelect(scripts, nil)

	// Output area
	output := widget.NewMultiLineEntry()
	output.SetPlaceHolder("Command output will appear here...")
	output.Disable()

	// Run button
	runButton := widget.NewButton("Run Script", func() {
		// Set SSH client details
		sshClient.Host = hostEntry.Text
		sshClient.Port = portEntry.Text
		sshClient.User = userEntry.Text
		sshClient.Password = passwordEntry.Text

		// Run selected script
		selectedScript := scriptSelect.Selected
		if selectedScript == "" {
			output.SetText("Please select a script to run.")
			return
		}

		result, err := sshClient.RunCommand(selectedScript)
		if err != nil {
			output.SetText(fmt.Sprintf("Error: %s", err))
			return
		}

		output.SetText(result)
	})

	// Layout
	form := container.NewVBox(
		widget.NewLabel("SSH Connection Details:"),
		widget.NewForm(
			widget.NewFormItem("Host", hostEntry),
			widget.NewFormItem("Port", portEntry),
			widget.NewFormItem("Username", userEntry),
			widget.NewFormItem("Password", passwordEntry),
		),
		widget.NewLabel("Select Script to Run:"),
		scriptSelect,
		runButton,
		output,
	)

	myWindow.SetContent(form)
	myWindow.Resize(fyne.NewSize(400, 400))
	myWindow.ShowAndRun()
}
