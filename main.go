package main

import (
    "os"
    "os/signal"
    "time"
    "context"
    "syscall"
    "database/sql"

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

    dbConn, err := sql.Open("sqlite3", "file:rivaclient.db?_foreign_keys=on")
    if err != nil {
        dbLog.Errorf("Failed to open database: %v", err)
        panic(err)
    }

    defer func() {
        if err := dbConn.Close(); err != nil {
            dbLog.Errorf("Failed to close database connection: %v", err)
        } else {
            mainLog.Infof("Database connection closed.")
        }
    }()

    ctxDb, cancelDb := context.WithTimeout(ctx, 5 * time.Second)
    defer cancelDb()
    if err := dbConn.PingContext(ctxDb); err != nil {
        dbLog.Errorf("Failed to ping database: %v", err)
        panic(err)
    }

    mainLog.Infof("Successfully connected to SQLite database.")

    container := sqlstore.NewWithDB(dbConn, "sqlite3", dbLog)
    if container == nil {
        mainLog.Errorf("Failed to create WhatsMeow SQL store container.")
        panic("nil WhatsMeow container")
    }

    deviceStore, err := container.GetFirstDevice(ctx)
    if err != nil {
        mainLog.Errorf("Failed to get device from store: %v", err)
        panic(err)
    }

    wm := whatsmeow.NewClient(deviceStore, wmLog)

    client := NewRIVAClient(wm, dbConn, mainLog)
    wm.AddEventHandler(client.EventHandler)

    if wm.Store.ID != nil {
        mainLog.Infof("Existing session found. Attempting to connect...")
        if err := wm.Connect(); err != nil {
            mainLog.Errorf("Failed to connect with existing session: %v", err)
            mainLog.Infof("This might be due to a session issue. Consider deleting the database file and restarting.")
            panic(err)
        }
    } else {
        mainLog.Infof("No existing session found. Preparing QR code for login...")
        qrChan, _ := wm.GetQRChannel(context.Background())

        if err := wm.Connect(); err != nil {
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
    wm.Disconnect()
    mainLog.Infof("Client disconnected. Exiting.")
}
