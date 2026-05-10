# Ulanzi TC001 Smart Clock

A Go-based MQTT publisher for the [AWTRIX 3](https://blueforcer.github.io/awtrix3/) firmware on the Ulanzi TC001 pixel clock.

## Architecture

```
┌───────────────────────────────────────────────────────────┐
│  Go Clock Publisher  ──MQTT──►  Mosquitto  ──MQTT──►  TC001│
│  (cmd/clock)                   (broker)         (AWTRIX)  │
└───────────────────────────────────────────────────────────┘
```

---

## PixelClock

### ----Requirements for Ulanzi CLock----

### 1. Flash AWTRIX onto the TC001

Follow the [AWTRIX web flasher](https://blueforcer.github.io/awtrix3/#/quickstart):
1. Connect TC001 via USB.
2. Open Chrome → navigate to the flasher → click **Install**.
3. After flashing, the device creates a WiFi hotspot (`AWTRIX_XXXXXX`).
4. Connect to it, enter your home WiFi credentials.
5. Find the device IP on your router or via the AWTRIX app.

### 2. Configure AWTRIX MQTT

In the AWTRIX web UI (`http://<device-ip>`):

- **Settings → MQTT**
  - Broker host: your machine's LAN IP (e.g. `192.168.1.50`)
  - Port: `1883`
  - Topic prefix: `awtrix` (default, or change and update `AWTRIX_PREFIX` env var)
  - Leave auth blank if using the default Mosquitto config.

#### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `MQTT_BROKER_URL` | `tcp://localhost:1883` | MQTT broker address |
| `AWTRIX_PREFIX` | `awtrix` | Device topic prefix (match AWTRIX setting) |
| `TICK_SECONDS` | `1` | How often to push the clock (seconds) |
| `MQTT_CLIENT_ID` | `pixelclock-publisher` | MQTT client identifier |
| `MQTT_USERNAME` | `` | Optional broker auth |
| `MQTT_PASSWORD` | `` | Optional broker auth |



### ----For Local Testing----

### 3. Start the MQTT Broker vi Makefile

```bash
make broker
# Or manually:
docker compose -f docker/docker-compose.yml up -d
```

### 4. Run the Clock Publisher

```bash
# Default — connects to localhost:1883
make run

# Custom broker (e.g. broker on another machine)
MQTT_BROKER_URL=tcp://192.168.1.50:1883 make run
```

### 5. Verify

Watch the MQTT traffic with:
```bash
docker run --rm --network host eclipse-mosquitto:2.0 \
  mosquitto_sub -h localhost -t "awtrix/#" -v
```

You should see a `awtrix/custom/clock` message every second.

---

## Customising the Clock

Edit the `style` block in `cmd/clock/main.go`:

```go
style := publisher.ClockStyle{
    TwelveHour:  false,         // true = 12h, false = 24h
    ShowSeconds: true,          // include seconds
    Color:       []int{0, 200, 255}, // [R, G, B] — cyan
    Background:  []int{0, 0, 0},    // black bg
    Rainbow:     false,         // true = rainbow cycle
}
```

---

## Project Structure

```
ulanzi-clock/
├── cmd/
│   └── clock/
│       └── main.go          # Entrypoint — clock ticker
├── internal/
│   └── publisher/
│       └── publisher.go     # MQTT client + AWTRIX payload builder
├── config/
│   └── config.go            # Env-based config
├── docker/
│   ├── docker-compose.yml   # Mosquitto broker
│   └── mosquitto.conf       # Broker config
├── k8s/                     # K8s Resource Manifests & Kustomize
│   ├── mqtt/
│   │   └── base/
│   │       ├── configmap.yaml
│   │       ├── deployment.yaml
│   │       ├── kustomization.yaml
│   │       ├── pvc.yaml
│   │       ├── service.yaml
│   │       └── overlays/
│   │           ├── dev/
│   │           │    ├── kustomization.yaml
│   │           │    └── patch-service.yaml
│   │           └── prod/
│   │               ├── kustomization.yaml
│   │               └── patch-service.yaml
│   └── pixelclock/
│       └── base/
│           ├── configmap.yaml
│           ├── deployment.yaml
│           ├── kustomization.yaml
│           ├── secret.yaml
│           └── overlays/
│               └──dev/
│               │   ├── kustomization.yaml
│               │   ├── patch-config.yaml
│               │   └── patch-secret.yaml
│               └── prod/
│                   ├── kustomization.yaml
│                   ├── patch-config.yaml
│                   └── patch-secret.yaml
├── Makefile
└── go.mod
```

---

## Phase 2 — Pixel Art UI (planned)

- Web UI to draw pixel art on a 32×8 grid
- Go backend receives the art, converts to AWTRIX `draw` commands
- Publishes to broker → appears on device in real time

AWTRIX supports drawing primitives via the `/draw` topic — lines, rects, pixels, text — all composable into a full frame.
