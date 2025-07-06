package main

import (
    "os"
    "os/signal"
    "fmt"
    "time"
    "context"
    "syscall"
    "database/sql"

    _ "github.com/mattn/go-sqlite3"
    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/store/sqlstore"

    "github.com/mdp/qrterminal/v3"
)

func main() {
    ctx := context.Background()
    logger := NewRIVAClientLog("RIVABotMain", "INFO")

    dbConn, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", rBotSqlFilePath))
    if err != nil {
        logger.Errorf("Failed to open database: %v", err)
        panic(err)
    }

    defer func() {
        if err := dbConn.Close(); err != nil {
            logger.Errorf("Failed to close database connection: %v", err)
        } else {
            logger.Infof("Database connection closed.")
        }
    }()

    ctxDb, cancelDb := context.WithTimeout(ctx, 5 * time.Second)
    defer cancelDb()
    if err := dbConn.PingContext(ctxDb); err != nil {
        logger.Errorf("Failed to ping database: %v", err)
        panic(err)
    }

    logger.Infof("Successfully connected to SQLite database.")

    container := sqlstore.NewWithDB(dbConn, "sqlite3", logger.logger)
    if container == nil {
        logger.Errorf("Failed to create WhatsMeow SQL store container.")
        panic("nil WhatsMeow container")
    }

    if err := container.Upgrade(ctx); err != nil {
        logger.Errorf("Failed to upgrade database: %w", err)
        panic(err)
    }

    deviceStore, err := container.GetFirstDevice(ctx)
    if err != nil {
        logger.Errorf("Failed to get device from store: %v", err)
        panic(err)
    }

    wm := whatsmeow.NewClient(deviceStore, logger.logger)

    client := (*RIVAClient).New(nil, wm, dbConn)
    wm.AddEventHandler(client.EventHandler)

    if wm.Store.ID != nil {
        logger.Infof("Existing session found. Attempting to connect...")
        if err := wm.Connect(); err != nil {
            logger.Errorf("Failed to connect with existing session: %v", err)
            logger.Infof("This might be due to a session issue. Consider deleting the database file and restarting.")
            panic(err)
        }
    } else {
        logger.Infof("No existing session found. Preparing QR code for login...")
        qrChan, _ := wm.GetQRChannel(context.Background())

        if err := wm.Connect(); err != nil {
            logger.Errorf("Failed to connect for QR scan: %v", err)
            panic(err)
        }

        for evt := range qrChan {
            switch evt.Event {
            case "code":
                logger.Infof("QR code to scan: %s", evt.Code)
                qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
                logger.Infof("Scan the QR code above with the WhatsApp app.")
            case "timeout":
                logger.Errorf("QR code scan timed out. Please try again.")
                panic("QR timeout")
            case "error":
                logger.Errorf("Error during QR scan process.")
                panic("QR error")
            default:
                logger.Infof("QR Login event: %s", evt.Event)
            }
        }

        logger.Infof("QR scan process finished.")
    }

    // Listen to CTRL+C
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)

    logger.Infof("Client is running. Press CTRL+C to disconnect and exit.")
    <-c

    logger.Infof("Disconnecting client...")
    wm.Disconnect()
    logger.Infof("Client disconnected. Exiting.")
}
