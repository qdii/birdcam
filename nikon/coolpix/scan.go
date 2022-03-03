package coolpix

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
)

var (
	fClientDevice = flag.String("client_name", "BirdCam", "Identify using this name (# are replaced with random digits).")
)

func Main() {
	flag.Parse()

	ctx := context.Background()

	fmt.Print("Searching for Bluetooth device: ")
	dev, err := linux.NewDevice()
	if err != nil {
		fmt.Println("default device not found:", err)
		return
	}
	ble.SetDefaultDevice(dev)
	fmt.Println("OK!")

	fmt.Print("Connecting: ")
	cam, err := Connect(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("OK!")

	fmt.Print("Exchanging MTU: ")
	if err := cam.SetMaxMTU(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("OK!")
	}

	fmt.Print("Authenticating: ")
	if err = cam.Authenticate(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("OK!")

	fmt.Print("Reading BatteryLevel: ")
	if b, err := cam.BatteryLevel(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("OK! BatteryLevel =", b)
	}

	fmt.Print("Writing ClientDeviceName: ")
	if err = cam.WriteString(ClientDeviceName, *fClientDevice, 32); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("OK!")
	}

	fmt.Print("Reading ServerDeviceName: ")
	if s, err := cam.ServerDeviceName(); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("OK! ServerDeviceName =", s)
	}

	fmt.Print("Writing CurrentTime: ")
	if err = cam.WriteCurrentTime(); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("OK!")
	}

	fmt.Print("Reading ConnectionConfiguration: ")
	if buf, err := cam.ReadBytes(ConnectionConfiguration); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("OK! ConnectionConfiguration:", buf)
	}

	fmt.Print("Enabling WiFi: ")
	if err := cam.SetWiFi(true); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("OK!")
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	quitting := false

	for {
		select {
		case <-time.After(time.Second / 10):
			if !quitting {
				fmt.Printf("\rCONN %s [%s] / TX %d / RSSI %d", A900ModelName, cam.Addr(), cam.TxPower, cam.ReadRSSI())
			}
		case <-quit:
			quitting = true
			fmt.Println()
			fmt.Print("Ctrl+C, clearing subscriptions: ")
			if err := cam.ClearSubscriptions(); err != nil {
				fmt.Print("failed:", err)
				return
			}
			fmt.Print("ok; closing connection: ")
			if err := cam.CancelConnection(); err != nil {
				fmt.Print("failed:", err)
				return
			}
			fmt.Print("ok")
		case <-cam.Disconnected():
			fmt.Println()
			fmt.Println("Disconnected.")
			return
		}
	}
}
