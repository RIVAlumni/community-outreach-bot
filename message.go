package main

import (
    "time"
    "strings"

    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/types"
    "go.mau.fi/whatsmeow/types/events"
)

type RIVAClientMessageDirection string
const (
    DirectionIncoming RIVAClientMessageDirection = "INCOMING"
    DirectionOutgoing RIVAClientMessageDirection = "OUTGOING"
)

type RIVAClientMessage struct {
    ID          string                     // Unique ID of the message
    From        string                     // Sender's Phone Number or Jabber ID
    To          string                     // Receipt's Phone Number or Jabber ID
    Direction   RIVAClientMessageDirection // Direction of message
    Content     string                     // Text content of the message
    Timestamp   time.Time                  // Timestamp of the message
    RawMessage  *events.Message            // Raw WhatsMeow message event
}

func NewRIVAClientMessage(wmClient *whatsmeow.Client, evt *events.Message) RIVAClientMessage {
    var content string

    switch {
    case evt.Message.GetConversation() != "":
        content = evt.Message.GetConversation()
    case evt.Message.GetExtendedTextMessage() != nil:
        content = evt.Message.GetExtendedTextMessage().GetText()
    case evt.Message.GetImageMessage() != nil:
        imageMsg := evt.Message.GetImageMessage()

        content = "Image message"
        if imageMsg.GetCaption() != "" {
            content += ": " + imageMsg.GetCaption()
        }
    case evt.Message.GetVideoMessage() != nil:
        videoMsg := evt.Message.GetVideoMessage()

        content = "Video message"
        if videoMsg.GetCaption() != "" {
            content += ": " + videoMsg.GetCaption()
        }
    case evt.Message.GetAudioMessage() != nil:
        content = "Audio message"
        // TODO: Use audio transformer models to transcribe audio
    case evt.Message.GetDocumentMessage() != nil:
        docMsg := evt.Message.GetDocumentMessage()
        content = "Document message: " + docMsg.GetTitle()
    default:
        content = "Unsupported message type"
    }

    message := RIVAClientMessage{
        ID:         evt.Info.ID,
        Content:    content,
        Timestamp:  evt.Info.Timestamp,
        RawMessage: evt,
    }
    message.From = message.getPhoneNumberFromJID(evt.Info.Sender)
    message.To = message.getPhoneNumberFromJID(evt.Info.Chat)
    message.Direction = message.getMessageDirection()

    if message.Direction == DirectionIncoming && wmClient.Store.ID != nil {
        message.To = message.getPhoneNumberFromJID(wmClient.Store.GetJID())
    }

    return message
}

func (msg *RIVAClientMessage) getMessageDirection() RIVAClientMessageDirection {
    if msg.RawMessage.Info.IsFromMe {
        return DirectionOutgoing
    }

    return DirectionIncoming
}

func (msg *RIVAClientMessage) getPhoneNumberFromJID(jid types.JID) string {
    if jid.User == "" {
        return ""
    }

    parts := strings.Split(jid.User, ":")
    return parts[0]
}

