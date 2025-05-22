package main

import (
    "os"
    "os/signal"
    "fmt"
    "context"
    "syscall"

    _ "github.com/mattn/go-sqlite3"
    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/store/sqlstore"
    "go.mau.fi/whatsmeow/types/events"
    waLog "go.mau.fi/whatsmeow/util/log"
)

func eventHandler(evt interface{}) {
    switch v := evt.(type) {
    case *events.Message:
        fmt.Println("Received a message!", v.Message.GetConversation())
    }
}

func main() {
    dbLog := waLog.Stdout("Database", "DEBUG", true)
    ctx := context.Background()
    container, err := sqlstore.New(ctx, "sqlite3", "file:examplestore.db?_foreign_keys=on", dbLog)
    if err != nil {
        panic(err)
    }

    deviceStore, err := container.GetFirstDevice(ctx)
    if err != nil {
        panic(err)
    }

    clientLog := waLog.Stdout("Client", "DEBUG", true)
    client := whatsmeow.NewClient(deviceStore, clientLog)
    client.AddEventHandler(eventHandler)

    if client.Store.ID == nil {
        // No ID stored, new login
        qrChan, _ := client.GetQRChannel(context.Background())
        err = client.Connect()
        if err != nil {
            panic(err)
        }

        for evt := range qrChan {
            if evt.Event == "code" {
                // Render the QR code here
                // e.g., qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
                // or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
                fmt.Println("QR code: ", evt.Code)
            } else {
                fmt.Println("Login event: ", evt.Event)
            }
        }
    } else {
        // Already logged in, just connect
        err = client.Connect()
        if err != nil {
            panic(err)
        }
    }

    // Listen to CTRL+C
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c

    client.Disconnect()
}
