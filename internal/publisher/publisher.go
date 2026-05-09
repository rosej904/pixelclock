package publisher

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// AWTRIXPayload represents a custom app payload for the AWTRIX firmware.
// Docs: https://blueforcer.github.io/awtrix3/#/api
type AWTRIXPayload struct {
	// Text to display. Supports placeholders.
	Text string `json:"text"`

	// Text color as [R, G, B] array. Omit for white.
	Color []int `json:"color,omitempty"`

	// Background color as [R, G, B] array. Omit for black.
	Background []int `json:"background,omitempty"`

	// Rainbow mode — cycles text through colors.
	Rainbow bool `json:"rainbow,omitempty"`

	// How long (ms) the app stays visible. 0 = forever (good for clock).
	Duration int `json:"duration,omitempty"`

	// Scroll speed override. Default is fine for clocks (text fits screen).
	ScrollSpeed int `json:"scrollSpeed,omitempty"`

	// Center text horizontally.
	Center bool `json:"center,omitempty"`

	// Lifetime (seconds): AWTRIX removes the app if not refreshed within this window.
	// Set slightly above your tick rate so the app stays alive.
	Lifetime int `json:"lifetime,omitempty"`

	// Icons (phase 2 — leave nil for now).
	// Icon string `json:"icon,omitempty"`
}

// ClockStyle controls the visual style of the clock.
type ClockStyle struct {
	// Use 12h format (true) or 24h (false).
	TwelveHour bool

	// Show seconds in the text.
	ShowSeconds bool

	// Text color [R,G,B]. Nil = white.
	Color []int

	// Background color [R,G,B]. Nil = black.
	Background []int

	// Rainbow cycle.
	Rainbow bool
}

// DefaultClockStyle is a clean white-on-black digital clock.
var DefaultClockStyle = ClockStyle{
	TwelveHour:  false,
	ShowSeconds: true,
	Color:       []int{0, 200, 255}, // cyan
}

// Publisher wraps an MQTT client and knows how to send AWTRIX payloads.
type Publisher struct {
	client       mqtt.Client
	devicePrefix string
}

// New creates a connected Publisher. It blocks until the connection is established.
func New(brokerURL, clientID, username, password, devicePrefix string) (*Publisher, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(clientID).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetOnConnectHandler(func(c mqtt.Client) {
			log.Printf("✅ Connected to MQTT broker at %s", brokerURL)
		}).
		SetConnectionLostHandler(func(c mqtt.Client, err error) {
			log.Printf("⚠️  MQTT connection lost: %v — will reconnect", err)
		})

	if username != "" {
		opts.SetUsername(username)
		opts.SetPassword(password)
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.WaitTimeout(10*time.Second) && token.Error() != nil {
		return nil, fmt.Errorf("mqtt connect: %w", token.Error())
	}
	if !client.IsConnected() {
		return nil, fmt.Errorf("mqtt: failed to connect within timeout")
	}

	return &Publisher{client: client, devicePrefix: devicePrefix}, nil
}

// PublishCustomApp pushes a payload to a named custom app slot on the device.
// AWTRIX custom app topic: <prefix>/custom/<appName>
func (p *Publisher) PublishCustomApp(appName string, payload AWTRIXPayload) error {
	topic := fmt.Sprintf("%s/custom/%s", p.devicePrefix, appName)

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	token := p.client.Publish(topic, 0, false, data)
	token.Wait()
	return token.Error()
}

// PublishNotification sends a one-shot notification (dismisses itself after Duration ms).
func (p *Publisher) PublishNotification(payload AWTRIXPayload) error {
	topic := fmt.Sprintf("%s/notify", p.devicePrefix)

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}

	token := p.client.Publish(topic, 0, false, data)
	token.Wait()
	return token.Error()
}

// ClockText formats the current time as a string based on style.
func ClockText(t time.Time, style ClockStyle) string {
	if style.TwelveHour {
		if style.ShowSeconds {
			return t.Format("3:04:05 PM")
		}
		return t.Format("3:04 PM")
	}
	if style.ShowSeconds {
		return t.Format("15:04:05")
	}
	return t.Format("15:04")
}

// BuildClockPayload builds an AWTRIXPayload for the current time.
func BuildClockPayload(t time.Time, style ClockStyle) AWTRIXPayload {
	return AWTRIXPayload{
		Text:       ClockText(t, style),
		Color:      style.Color,
		Background: style.Background,
		Rainbow:    style.Rainbow,
		Center:     true,
		Duration:   0,
		Lifetime:   5, // Remove app if not refreshed within 5s
	}
}

// Disconnect cleanly closes the MQTT connection.
func (p *Publisher) Disconnect() {
	p.client.Disconnect(500)
}
