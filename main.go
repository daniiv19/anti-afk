package main

import (
	"strconv"
	"strings"
	"sync"
	"time"

	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/goearth/shockwave/in"
	"xabbo.b7c.io/goearth/shockwave/out"
)

// Initialize the extension with metadata
var ext = g.NewExt(g.ExtInfo{
	Title:       "AFK Tracker",
	Description: "Anti AFK. An extension that tracks how long you have been AFK.",
	Author:      "danii / v19",
	Version:     "0.5",
})

var (
	afkTimer          *time.Timer
	lastActionTime    time.Time
	afkDuration       = 5 * time.Minute
	setupMutex        sync.Mutex
	afkActive         bool
	sendingAfkMessage bool // Flag to indicate if the script is sending an AFK message
)

// Entry point of the application
func main() {
	ext.Initialized(onInitialized)
	ext.Connected(onConnected)
	ext.Disconnected(onDisconnected)
	ext.Intercept(out.CHAT, out.WHISPER, out.DICE_OFF, out.THROW_DICE, out.MOVE).With(resetAfkTimer)
	ext.Intercept(in.CHAT).With(ignoreAfkMessages)
	ext.Run()
}

func onInitialized(e g.InitArgs) {
	lastActionTime = time.Now()
	startAfkTimer()
}

func onConnected(e g.ConnectArgs) {
}

func onDisconnected() {
	if afkTimer != nil {
		afkTimer.Stop()
	}
}

func resetAfkTimer(e *g.Intercept) {
	setupMutex.Lock()
	defer setupMutex.Unlock()
	if sendingAfkMessage {
		return
	}
	lastActionTime = time.Now()
	if !afkActive {
		startAfkTimer()
	}
}

func startAfkTimer() {
	if afkTimer != nil {
		afkTimer.Stop()
	}
	afkTimer = time.AfterFunc(afkDuration, func() {
		setupMutex.Lock()
		defer setupMutex.Unlock()
		afkActive = true
		sendAfkMessages()
		afkActive = false
		startAfkTimer() // Continue the timer for the next interval
	})
}

func sendAfkMessages() {
	sendingAfkMessage = true
	time.Sleep(2 * time.Second) // Pause before sending the message

	elapsed := time.Since(lastActionTime)
	message := formatAfkMessage(elapsed)
	ext.Send(in.CHAT, message)

	time.Sleep(2 * time.Second) // Pause after sending the message
	sendingAfkMessage = false
}

func formatAfkMessage(elapsed time.Duration) string {
	totalMinutes := int(elapsed.Minutes())
	days := totalMinutes / (60 * 24)
	hours := (totalMinutes % (60 * 24)) / 60
	minutes := totalMinutes % 60

	var messageParts []string
	if days > 0 {
		messageParts = append(messageParts, strconv.Itoa(days)+" day"+pluralize(days))
	}
	if hours > 0 {
		messageParts = append(messageParts, strconv.Itoa(hours)+" hour"+pluralize(hours))
	}
	if minutes > 0 || (days == 0 && hours == 0) {
		messageParts = append(messageParts, strconv.Itoa(minutes)+" minute"+pluralize(minutes))
	}

	return "I have been AFK for " + strings.Join(messageParts, " ")
}

func pluralize(value int) string {
	if value == 1 {
		return ""
	}
	return "s"
}

func ignoreAfkMessages(e *g.Intercept) {
	msg := e.Packet.ReadString()
	if strings.Contains(msg, "I'm AFK") || strings.Contains(msg, "I have been AFK for") {
		e.Block()
	}
}
