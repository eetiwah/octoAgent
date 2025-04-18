# octoAgent

A Go-based agent framework for interacting with 3D printers via OctoPrint, part of the broader *NeedToKnow* decentralized swarm architecture.

## Prerequisites

- **OctoPrint Service**: Required for API communication.
  - # Installation**: Install OctoPrint on your system (e.g., Ubuntu NUC):
    cd [local directory]
    git clone https://github.com/OctoPrint/OctoPrint.git
    cd OctoPrint
    python3 -m venv venv
    pip install .
    source venv/bin/activate

    # Enable serial port logging -> required for positioning to work
    ~/.octoprint/config.yaml
    add:
    serial: -> this should already be present
      log: true -> add this

    # start the service
    [local directory]/OctoPrint/venv/bin/octoprint serve

    # URL and apiKey
    open browser
    localhost:5000
    create account
    Goto User Settings under account name
    API key is located there - first time, you will need to create the key
    Add key and URL to .env file

## Tor executable
  make sure the file permissions on the tor executable is set appropriately run
  cd into linux/dependencies
  ./tor --version to test. If errors out, permissions problem!

  make sure the file permissions for profiles within tor/octoAgent is set appropriately for accessing
  chmod -R go+rx main/tor/linux/octoAgent/profiles maybe needed

## Position Tracking
  - Function**: `NewPositionTracker(logPath string) (*PositionTracker, error)`
  - Data**: Tracks Z from `serial.log` (e.g., `G1 Z0.2`, `G0 Z0.6`).
  - Output**: `Position{Z: 0.200}` via `GetPosition`.
