package main

import (
	"time"

	"go.mau.fi/whatsmeow/types"
)

type SequentialMessageHandlerFunc func(
    RClient      *RIVAClient,
    Message      RIVAClientMessage,
    next         func(),
    stop         func(),
)

type ParallelMessageHandlerFunc func(
    RClient      *RIVAClient,
    Message      RIVAClientMessage,
) error

func FilterOldMessagesHandler(rc *RIVAClient, message RIVAClientMessage, next func(), stop func()) {
    if !rc.LastSuccessfulConnectionTime.IsZero() && message.Timestamp.Before(rc.LastSuccessfulConnectionTime) {
        rc.Log.MainLog.Infof("Ignoring old message: %+v", message)
        stop()
        return
    }
    next()
}

func FilterUnsupportedMessagesHandler(rc *RIVAClient, message RIVAClientMessage, next func(), stop func()) {
    if message.Type == TypeUnsupported {
        rc.Log.MainLog.Infof("Ignoring unsupported message: %+v", message)
        stop()
        return
    }

    if message.IsNewsletter() {
        rc.Log.MainLog.Infof("Ignoring newsletter message: %+v", message)
        stop()
        return
    }

    next()
}

func LogNewMessageHandler(rc *RIVAClient, message RIVAClientMessage, next func(), stop func()) {
    rc.Log.MainLog.Infof("New message: %+v", message)
    next()
}

func SendGreetingMessageHandler(rc *RIVAClient, message RIVAClientMessage, next func(), stop func()) {
    if !message.IsSentByMe() && !message.IsGroup {
        rc.Log.MainLog.Infof("SendGreetingMessageHandler: Processing message: %+v", message)

        fromJID := message.FromNonAD
        isNewsletter := message.IsNewsletter()

        switch {
        case isNewsletter:
            rc.Log.MainLog.Infof("SendGreetingMessageHandler: Sender %s is a newsletter. Skipping greeting", fromJID)
        default:
            lastInteraction, found, err := rc.DB.GetLastInteractionTime(fromJID)
            if err != nil {
                rc.Log.MainLog.Errorf("SendGreetingMessageHandler: Error getting last interaction time for %s: %v", fromJID, err)
                rc.Log.MainLog.Errorf("SendGreetingMessageHandler: Skipping greeting logic: %+v", message)
                next()
                return
            }
            
            shouldSendGreeting := false
            if !found {
                shouldSendGreeting = true
                rc.Log.MainLog.Infof("SendGreetingMessageHandler: No last interaction record for %s", fromJID)
            } else if found && time.Since(lastInteraction).Hours() >= rBotGreetingCooldownHours {
                shouldSendGreeting = true
                rc.Log.MainLog.Infof("SendGreetingMessageHandler: Last interaction with %s was at %s", fromJID, lastInteraction.Format(time.RFC3339))
            } else {
                rc.Log.MainLog.Infof("SendGreetingMessageHandler: Last interaction with %s was at %s", fromJID, lastInteraction.Format(time.RFC3339))
                rc.Log.MainLog.Infof("SendGreetingMessageHandler: Greeting cooldown: %+v", message)
            }

            if shouldSendGreeting {
                if err := rc.SendGreetingMessage(fromJID); err != nil {
                    rc.Log.MainLog.Errorf("SendGreetingMessageHandler: Failed to send greeting for %s: %v", fromJID, err)
                } else {
                    rc.Log.MainLog.Infof("SendGreetingMessageHandler: Sending greeting: %+v", message)
                }
            }

        }

        if err := rc.DB.UpdateLastInteractionTime(fromJID, message.Timestamp); err != nil {
            rc.Log.MainLog.Errorf("SendGreetingMessageHandler: Failed to update last interaction for %s: %v", fromJID, err)
        } else {
            rc.Log.MainLog.Infof("SendGreetingMessageHandler: Updating last interaction time for %s to %s", fromJID, message.Timestamp.Format(time.RFC3339))
        }
    }
    next()
}

func AutoEditOutgoingMessageHandler(rc *RIVAClient, message RIVAClientMessage, next func(), stop func()) {
    if message.IsSentByMe() {
        rc.Log.MainLog.Infof("AutoEditOutgoingMessageHandler: Processing message: %+v", message)
        go func() {
            time.Sleep(1 * time.Second)
            rc.Log.MainLog.Warnf("AutoEditOutgoingMessageHandler: Stub fired")
            // Complete message edit via helper function
        }()
    }
    next()
}

