package main

import (
	"time"
)

type SequentialMessageHandlerFunc func(
    rc   *RIVAClient,
    msg  RIVAClientMessage,
    next func(),
    stop func(),
) func()

type ParallelMessageHandlerFunc func(
    rc  *RIVAClient,
    msg RIVAClientMessage,
) error

func FilterOldMessagesHandler(rc *RIVAClient, msg RIVAClientMessage, next func(), stop func()) func() {
    if !rc.LastSuccessfulConnectionTime.IsZero() && msg.Timestamp.Before(rc.LastSuccessfulConnectionTime) {
        rc.Log.MainLog.Infof("Ignoring old message: %+v", msg)
        return stop
    }
    
    return next
}

func FilterUnsupportedMessagesHandler(rc *RIVAClient, msg RIVAClientMessage, next func(), stop func()) func() {
    if msg.Type == TypeUnsupported {
        rc.Log.MainLog.Infof("Ignoring unsupported message: %+v", msg)
        return stop
    }

    if msg.IsNewsletter() {
        rc.Log.MainLog.Infof("Ignoring newsletter message: %+v", msg)
        return stop
    }

    return next
}

func LogNewMessageHandler(rc *RIVAClient, msg RIVAClientMessage, next func(), stop func()) func() {
    rc.Log.MainLog.Infof("New message: %+v", msg)
    return next
}

func SendGreetingMessageHandler(rc *RIVAClient, msg RIVAClientMessage, next func(), stop func()) func() {
    if !msg.IsSentByMe() && !msg.IsGroup {
        rc.Log.MainLog.Infof("SendGreetingMessageHandler: Processing message: %+v", msg)

        fromJID := msg.FromNonAD
        isNewsletter := msg.IsNewsletter()

        switch {
        case isNewsletter:
            rc.Log.MainLog.Infof("SendGreetingMessageHandler: Sender %s is a newsletter. Skipping greeting", fromJID)
        default:
            lastInteraction, found, err := rc.DB.GetLastInteractionTime(fromJID)
            if err != nil {
                rc.Log.MainLog.Errorf("SendGreetingMessageHandler: Error getting last interaction time for %s: %v", fromJID, err)
                rc.Log.MainLog.Errorf("SendGreetingMessageHandler: Skipping greeting logic: %+v", msg)
                return next
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
                rc.Log.MainLog.Infof("SendGreetingMessageHandler: Greeting cooldown: %+v", msg)
            }

            if shouldSendGreeting {
                if err := rc.SendGreetingMessage(fromJID); err != nil {
                    rc.Log.MainLog.Errorf("SendGreetingMessageHandler: Failed to send greeting for %s: %v", fromJID, err)
                } else {
                    rc.Log.MainLog.Infof("SendGreetingMessageHandler: Sending greeting: %+v", msg)
                }
            }

        }

        if err := rc.DB.UpdateLastInteractionTime(fromJID, msg.Timestamp); err != nil {
            rc.Log.MainLog.Errorf("SendGreetingMessageHandler: Failed to update last interaction for %s: %v", fromJID, err)
        } else {
            rc.Log.MainLog.Infof("SendGreetingMessageHandler: Updating last interaction time for %s to %s", fromJID, msg.Timestamp.Format(time.RFC3339))
        }
    }

    return next
}

func AutoEditOutgoingMessageHandler(rc *RIVAClient, msg RIVAClientMessage, next func(), stop func()) func() {
    if msg.IsSentByMe() {
        rc.Log.MainLog.Infof("AutoEditOutgoingMessageHandler: Processing message: %+v", msg)
        go func() {
            if err := rc.EditIncludeHeaderFooterMessage(msg); err != nil {
                rc.Log.MainLog.Errorf("Error during auto-edit attempt for message %s: %v", msg.ID, err)
            }
        }()
    }

    return next
}

