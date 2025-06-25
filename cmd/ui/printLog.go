package ui

import (
	"fmt"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/cmd/entity"
)

// PrintLog is used to print a log message to the log window
func (u *UI) PrintLog(param entity.Log) {
	// Get last log message
	lastLog := u.Layout.LogList.GetText(false)

	warnaLog := "white"
	switch param.Type {
	case entity.LOG_ERROR:
		warnaLog = "red"
	case entity.LOG_WARNING:
		warnaLog = "yellow"
	}

	formatLog := fmt.Sprintf("[green][%s] [%s]%s [white]\n", time.Now().Format(time.RFC822), warnaLog, param.Content)
	u.Layout.LogList.SetWordWrap(true).SetText(lastLog + formatLog)

	// Scroll log window to bottom
	u.Layout.LogList.ScrollToEnd()
}

func (u *UI) PrintOutput(param entity.Output) {
	var (
		// metadata  string
		newBuffer string
	)

	out := u.Layout.OutputPanel
	// _, _, width, _ := out.TextArea.GetRect()

	// timeHeader := time.Now().Format("15:04:05 02/01/2006")

	// if param.WithHeader {
	// 	if u.GRPC != nil && len(u.GRPC.Conn.ParseMetadata()) > 0 {
	// 		for _, meta := range u.GRPC.Conn.ParseMetadata() {
	// 			metadata += "  ► " + meta + "\n"
	// 		}

	// 		metaHeader := strings.Repeat(string(tcell.RuneCkBoard), 2) + "[ Request Metadata ]" + (strings.Repeat(string(tcell.RuneCkBoard), width-47)) + "[ " + timeHeader + " ]" + strings.Repeat(string(tcell.RuneCkBoard), 2) + "\n\n"
	// 		newBuffer = metaHeader + metadata + "\n"
	// 	}

	// 	payloadHeader := strings.Repeat(string(tcell.RuneCkBoard), 2) + "[ Request Payload ]" + (strings.Repeat(string(tcell.RuneCkBoard), width-46)) + "[ " + timeHeader + " ]" + strings.Repeat(string(tcell.RuneCkBoard), 2) + "\n\n"
	// 	newBuffer += payloadHeader + u.GRPC.Conn.RequestPayload

	// 	responseHeader := "\n\n" + strings.Repeat(string(tcell.RuneCkBoard), 2) + "[ Response Payload ]" + (strings.Repeat(string(tcell.RuneCkBoard), width-47)) + "[ " + timeHeader + " ]" + strings.Repeat(string(tcell.RuneCkBoard), 2) + "\n"
	// 	newBuffer += responseHeader + param.Content
	// } else {
	// 	newBuffer = param.Content
	// }

	newBuffer = param.Content

	out.TextArea.SetText(newBuffer, param.CursorAtEnd)
}
