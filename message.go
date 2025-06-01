package main

import (
    "time"
    "context"
    "strings"

    "go.mau.fi/whatsmeow/types"
    "go.mau.fi/whatsmeow/types/events"
    "google.golang.org/protobuf/proto"

    waProto "go.mau.fi/whatsmeow/binary/proto"
)

type RIVAClientMessageDirection string
const (
    DirectionIncoming RIVAClientMessageDirection = "INCOMING"
    DirectionOutgoing RIVAClientMessageDirection = "OUTGOING"
)

type RIVAClientMessage struct {
    RClient    *RIVAClient
    ID         string                     // Unique ID of the message
    From       types.JID                  // Sender's JID
    FromPN     string                     // Sender's Phone Number or LID
    FromNonAD  types.JID                  // Sender's JID without device part
    To         types.JID                  // Recipient's JID
    ToPN       string                     // Recipient's Phone Number or LID
    ToNonAD    types.JID                  // Recipient's JID without device part
    Direction  RIVAClientMessageDirection // Direction of message
    IsGroup    bool                       // If message came from a group chat
    Content    string                     // Text content of the message
    Timestamp  time.Time                  // Timestamp of the message
    RawMessage *events.Message            // Raw WhatsMeow message event
}

// TODO: Change to Reply
func (msg *RIVAClientMessage) SendGreetingMessage(recipientJID types.JID) error {
    buildMsg := &waProto.Message{Conversation: proto.String(rBotGreetingCooldownMessage)}

    sanitisedJID := recipientJID.ToNonAD()

    _, err := msg.RClient.WMClient.SendMessage(context.Background(), sanitisedJID, buildMsg)
    if err != nil {
        msg.RClient.Log.MainLog.Errorf("Failed to send greeting message to %s: %v", recipientJID, err)
        return err
    }

    msg.RClient.Log.MainLog.Infof("Greeting message sent to %s", recipientJID)
    return nil
}

func (_ *RIVAClientMessage) New(rClient *RIVAClient, evt *events.Message) RIVAClientMessage {
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

    msg := RIVAClientMessage{
        RClient:    rClient,
        ID:         evt.Info.ID,
        From:       evt.Info.Sender,
        To:         evt.Info.Chat,
        IsGroup:    evt.Info.IsGroup,
        Content:    content,
        Timestamp:  evt.Info.Timestamp,
        RawMessage: evt,
    }
    msg.FromPN = msg.getPhoneNumberFromJID(msg.From)
    msg.FromNonAD = msg.From.ToNonAD()
    msg.ToPN = msg.getPhoneNumberFromJID(msg.To)
    msg.ToNonAD = msg.To.ToNonAD()
    msg.Direction = msg.getMessageDirection()

    if msg.Direction == DirectionIncoming && rClient.WMClient.Store.ID != nil {
        msg.To = rClient.WMClient.Store.GetJID()
    }

    return msg
}

func (msg *RIVAClientMessage) IsSentByMe() bool {
    if msg.RClient.WMClient.Store == nil || msg.RClient.WMClient.Store.ID == nil {
        return false
    }

    return msg.getPhoneNumberFromJID(msg.From) == msg.RClient.WMClient.Store.ID.User
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

