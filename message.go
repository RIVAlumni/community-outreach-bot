package main

import (
	"fmt"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type RIVAClientMessageDirection string
const (
    DirectionIncoming RIVAClientMessageDirection = "INCOMING"
    DirectionOutgoing RIVAClientMessageDirection = "OUTGOING"
)

type RIVAClientMessageType string
const (
    TypeText        RIVAClientMessageType = "TEXT"
    TypeImage       RIVAClientMessageType = "IMAGE"
    TypeVideo       RIVAClientMessageType = "VIDEO"
    TypeAudio       RIVAClientMessageType = "AUDIO"
    TypeSticker     RIVAClientMessageType = "STICKER"
    TypeDocument    RIVAClientMessageType = "DOCUMENT"
    TypeUnsupported RIVAClientMessageType = "UNSUPPORTED"
)

type RIVAClientMessage struct {
    RClient    *RIVAClient
    ID         string                     // Unique ID of the message
    Type       RIVAClientMessageType      // Message type
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

func (*RIVAClientMessage) New(rClient *RIVAClient, evt *events.Message) RIVAClientMessage {
    var msgType RIVAClientMessageType
    var msgContent string

    switch {
    case evt.Message.GetConversation() != "":
        msgType = TypeText
        msgContent = evt.Message.GetConversation()
    case evt.Message.GetExtendedTextMessage() != nil:
        msgType = TypeText
        msgContent = evt.Message.GetExtendedTextMessage().GetText()
    case evt.Message.GetImageMessage() != nil:
        msgType = TypeImage
        imgMsg := evt.Message.GetImageMessage()

        msgContent = imgMsg.GetCaption()
        if msgContent != "" {
            msgContent = "[IMAGE]"
        }
    case evt.Message.GetVideoMessage() != nil:
        msgType = TypeVideo
        vidMsg := evt.Message.GetVideoMessage()

        msgContent = vidMsg.GetCaption()
        if msgContent != "" {
            msgContent = "[VIDEO]"
        }
    case evt.Message.GetAudioMessage() != nil:
        msgType = TypeAudio
        msgContent = "[AUDIO]"
        // TODO: Use audio transformer models to transcribe audio
    case evt.Message.GetDocumentMessage() != nil:
        msgType = TypeDocument
        docMsg := evt.Message.GetDocumentMessage()
        msgContent = fmt.Sprintf("[DOCUMENT] %s (%s)", docMsg.GetFileName(), docMsg.GetFileName())
    case evt.Message.GetStickerMessage() != nil:
        msgType = TypeSticker
        msgContent = "[STICKER]"
    default:
        msgType = TypeUnsupported
        msgContent = "[UNSUPPORTED]"
    }

    msg := RIVAClientMessage{
        RClient:    rClient,
        ID:         evt.Info.ID,
        Type:       msgType,
        From:       evt.Info.Sender,
        To:         evt.Info.Chat,
        IsGroup:    evt.Info.IsGroup,
        Content:    msgContent,
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

func (msg *RIVAClientMessage) IsNewsletter() bool {
    if msg.From.Server == types.NewsletterServer {
        return true
    }

    return false
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

