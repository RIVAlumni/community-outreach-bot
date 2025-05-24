package main

import (
    "os"
    "os/signal"
    "context"
    "syscall"

    _ "github.com/mattn/go-sqlite3"
    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/store/sqlstore"
    waLog "go.mau.fi/whatsmeow/util/log"

    "github.com/mdp/qrterminal/v3"
)

func main() {
    ctx := context.Background()

    mainLog := waLog.Stdout("RIVABot::Log", "INFO", true)
    dbLog := waLog.Stdout("Sqlite3::Log", "ERROR", true)
    wmLog := waLog.Stdout("WhatsMeow::Log", "INFO", true)

    container, err := sqlstore.New(ctx, "sqlite3", "file:rivaclient.db?_foreign_keys=on", dbLog)
    if err != nil {
        mainLog.Errorf("Failed to connect to database: %v", err)
        panic(err)
    }

    deviceStore, err := container.GetFirstDevice(ctx)
    if err != nil {
        mainLog.Errorf("Failed to get device from store: %v", err)
        panic(err)
    }

    client := whatsmeow.NewClient(deviceStore, wmLog)

    rc := NewRIVAClient(client, mainLog)
    client.AddEventHandler(rc.EventHandler)

    if client.Store.ID != nil {
        mainLog.Infof("Existing session found. Attempting to connect...")
        if err := client.Connect(); err != nil {
            mainLog.Errorf("Failed to connect with existing session: %v", err)
            mainLog.Infof("This might be due to a session issue. Consider deleting the database file and restarting.")
            panic(err)
        }
    } else {
        mainLog.Infof("No existing session found. Preparing QR code for login...")
        qrChan, _ := client.GetQRChannel(context.Background())

        if err := client.Connect(); err != nil {
            mainLog.Errorf("Failed to connect for QR scan: %v", err)
            panic(err)
        }

        for evt := range qrChan {
            switch evt.Event {
            case "code":
                mainLog.Infof("QR code to scan: %s", evt.Code)
                qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
                mainLog.Infof("Scan the QR code above with the WhatsApp app.")
            case "timeout":
                mainLog.Errorf("QR code scan timed out. Please try again.")
                panic("QR timeout")
            case "error":
                mainLog.Errorf("Error during QR scan process.")
                panic("QR error")
            default:
                mainLog.Infof("QR Login event: %s", evt.Event)
            }
        }

        mainLog.Infof("QR scan process finished.")
    }

    // Listen to CTRL+C
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)

    mainLog.Infof("Client is running. Press CTRL+C to disconnect and exit.")
    <-c

    mainLog.Infof("Disconnecting client...")
    client.Disconnect()
    mainLog.Infof("Client disconnected. Exiting.")
}
