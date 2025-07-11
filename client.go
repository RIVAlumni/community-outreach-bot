package main

import (
	"database/sql"
    "fmt"
	"time"
    "context"

	"go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"

	waProto "go.mau.fi/whatsmeow/binary/proto"
)

type RIVAClient struct {
    WMClient                     *whatsmeow.Client
    Handlers                     *RIVAClientEvent
    DB                           *RIVAClientDB
    Log                          *RIVAClientLog
    LastSuccessfulConnectionTime time.Time
}

func (*RIVAClient) New(wmClient *whatsmeow.Client, db *sql.DB) *RIVAClient {
    rc := &RIVAClient{
        WMClient:                     wmClient,
        Log:                          NewRIVAClientLog("RIVABotClient", "INFO"),
        LastSuccessfulConnectionTime: time.Time{},
    }

    rc.DB       = (*RIVAClientDB).New(nil, rc, db)
    rc.Handlers = (*RIVAClientEvent).New(nil, rc, rc.DB)
    return rc
}

func (rc *RIVAClient) EditIncludeHeaderFooterMessage(msg RIVAClientMessage) error {
    if msg.HasOrgPrefix() {
        return nil
    }

    newContent := fmt.Sprintf(rBotOrgHeaderFooter, msg.Content)
    newPayload := &waE2E.Message{}
    if msg.Type == TypeTextConv {
        newPayload.Conversation = proto.String(newContent)
    } else if msg.Type == TypeTextExt {
        newPayload.ExtendedTextMessage = &waProto.ExtendedTextMessage{
            Text:        proto.String(newContent),
            ContextInfo: msg.RawMessage.Message.GetExtendedTextMessage().GetContextInfo(),
        }
    }

    resp, err := rc.WMClient.SendMessage(context.Background(),
                                         msg.To,
                                         rc.WMClient.BuildEdit(msg.To,
                                                               msg.ID,
                                                               newPayload))
    if err != nil {
        rc.Log.Errorf("Failed to edit message id %s in chat %s: %v", msg.ID, msg.To, err)
        return err
    }

    rc.Log.Infof("Successfully edited message ID %s in chat %s. New ID: %s, Timestamp: %s", msg.ID, msg.To, resp.ID, resp.Timestamp)
    return nil
}

func (rc *RIVAClient) SendGreetingMessage(recipientJID types.JID) error {
    buildMsg := &waProto.Message{
        Conversation: proto.String(rBotGreetingMessage),
    }

    sanitisedJID := recipientJID.ToNonAD()

    _, err := rc.WMClient.SendMessage(context.Background(), sanitisedJID, buildMsg)
    if err != nil {
        rc.Log.Errorf("Failed to send greeting message to %s: %v", recipientJID, err)
        return err
    }

    rc.Log.Infof("Greeting message sent to %s", recipientJID)
    return nil
}

func (rc *RIVAClient) EventHandler(evt interface{}) {
    switch v := evt.(type) {
    case *events.AppState:
        rc.Handlers.EventAppState(v)
    case *events.AppStateSyncComplete:
        rc.Handlers.EventAppStateSyncComplete(v)
    case *events.Archive:
        rc.Handlers.EventArchive(v)
    case *events.Blocklist:
        rc.Handlers.EventBlocklist(v)
    case *events.BlocklistAction:
        rc.Handlers.EventBlocklistAction(v)
    case *events.BlocklistChange:
        rc.Handlers.EventBlocklistChange(v)
    case *events.BlocklistChangeAction:
        rc.Handlers.EventBlocklistChangeAction(v)
    case *events.BusinessName:
        rc.Handlers.EventBusinessName(v)
    case *events.CallAccept:                    // Useful for auto-rejecting calls
        rc.Handlers.EventCallAccept(v)
    case *events.CallOffer:                     // Useful for auto-rejecting calls
        /*
         * FOR 1:1 CALLS
         *
         * Event is fired when a call is received from WhatsApp.
         * We can get the caller JID from v.From and v.CallID
         * and use rc.WMClient.RejectCall(v.From, v.CallID)
         *
         * We should also check if the call originated from us,
         * or if it came externally. By default, we should reject
         * external calls without entertaining them.
         */
        rc.Handlers.EventCallOffer(v)
    case *events.CallOfferNotice:               // Useful for auto-rejecting calls
        /*
         * FOR GROUP CALLS
         *
         * Event is fired when a call is received from WhatsApp.
         * We can get the caller JID from v.From and v.CallID
         * and use rc.WMClient.RejectCall(v.From, v.CallID)
         *
         * We should also check if the call originated from us,
         * or if it came externally. By default, we should reject
         * external calls without entertaining them.
         */
        rc.Handlers.EventCallOfferNotice(v)
    case *events.CallPreAccept:                 // Useful for auto-rejecting calls
        rc.Handlers.EventCallPreAccept(v)
    case *events.CallReject:                    // Useful for auto-rejecting calls
        rc.Handlers.EventCallReject(v)
    case *events.CallRelayLatency:              // Useful for auto-rejecting calls
        rc.Handlers.EventCallRelayLatency(v)
    case *events.CallTerminate:                 // Useful for auto-rejecting calls
        rc.Handlers.EventCallTerminate(v)
    case *events.CallTransport:                 // Useful for auto-rejecting calls
        rc.Handlers.EventCallTransport(v)
    case *events.ChatPresence:
        rc.Handlers.EventChatPresence(v)
    case *events.ClearChat:                     // We might want to raise a warning
        rc.Handlers.EventClearChat(v)
    case *events.ClientOutdated:
        rc.Handlers.EventClientOutdated(v)
    case *events.ConnectFailure:
        rc.Handlers.EventConnectFailure(v)
    case *events.ConnectFailureReason:
        rc.Handlers.EventConnectFailureReason(v)
    case *events.Connected:
        rc.LastSuccessfulConnectionTime = time.Now()
        rc.Handlers.EventConnected(v)
    case *events.Contact:
        rc.Handlers.EventContact(v)
    case *events.DecryptFailMode:
        rc.Handlers.EventDecryptFailMode(v)
    case *events.DeleteChat:
        rc.Handlers.EventDeleteChat(v)
    case *events.DeleteForMe:
        rc.Handlers.EventDeleteForMe(v)
    case *events.Disconnected:
        rc.Handlers.EventDisconnected(v)
    case *events.FBMessage:
        rc.Handlers.EventFBMessage(v)
    case *events.GroupInfo:
        rc.Handlers.EventGroupInfo(v)
    case *events.HistorySync:
        rc.Handlers.EventHistorySync(v)
    case *events.IdentityChange:
        rc.Handlers.EventIdentityChange(v)
    case *events.JoinedGroup:
        rc.Handlers.EventJoinedGroup(v)
    case *events.KeepAliveRestored:
        rc.Handlers.EventKeepAliveRestored(v)
    case *events.KeepAliveTimeout:
        rc.Handlers.EventKeepAliveTimeout(v)
    case *events.LabelAssociationChat:
        rc.Handlers.EventLabelAssociationChat(v)
    case *events.LabelAssociationMessage:
        rc.Handlers.EventLabelAssociationMessage(v)
    case *events.LabelEdit:
        rc.Handlers.EventLabelEdit(v)
    case *events.LoggedOut:
        rc.Handlers.EventLoggedOut(v)
    case *events.MarkChatAsRead:
        rc.Handlers.EventMarkChatAsRead(v)
    case *events.MediaRetry:
        rc.Handlers.EventMediaRetry(v)
    case *events.MediaRetryError:
        rc.Handlers.EventMediaRetryError(v)
    case *events.Message:
        rc.Handlers.EventMessage(v)
    case *events.Mute:
        rc.Handlers.EventMute(v)
    case *events.NewsletterJoin:                // Useless for now
        rc.Handlers.EventNewsletterJoin(v)
    case *events.NewsletterLeave:               // Useless for now
        rc.Handlers.EventNewsletterLeave(v)
    case *events.NewsletterLiveUpdate:          // Useless for now
        rc.Handlers.EventNewsletterLiveUpdate(v)
    case *events.NewsletterMessageMeta:         // Useless for now
        rc.Handlers.EventNewsletterMessageMeta(v)
    case *events.NewsletterMuteChange:          // Useless for now
        rc.Handlers.EventNewsletterMuteChange(v)
    case *events.OfflineSyncCompleted:
        rc.Handlers.EventOfflineSyncCompleted(v)
    case *events.OfflineSyncPreview:
        rc.Handlers.EventOfflineSyncPreview(v)
    case *events.PairError:
        rc.Handlers.EventPairError(v)
    case *events.PairSuccess:
        rc.Handlers.EventPairSuccess(v)
    case *events.PermanentDisconnect:
        rc.Handlers.EventPermanentDisconnect(v)
    case *events.Picture:
        rc.Handlers.EventPicture(v)
    case *events.Pin:
        rc.Handlers.EventPin(v)
    case *events.Presence:
        rc.Handlers.EventPresence(v)
    case *events.PrivacySettings:
        rc.Handlers.EventPrivacySettings(v)
    case *events.PushName:
        rc.Handlers.EventPushName(v)
    case *events.PushNameSetting:
        rc.Handlers.EventPushNameSetting(v)
    case *events.QR:
        rc.Handlers.EventQR(v)
    case *events.QRScannedWithoutMultidevice:
        rc.Handlers.EventQRScannedWithoutMultidevice(v)
    case *events.Receipt:
        rc.Handlers.EventReceipt(v)
    case *events.Star:
        rc.Handlers.EventStar(v)
    case *events.StreamError:
        rc.Handlers.EventStreamError(v)
    case *events.StreamReplaced:
        rc.Handlers.EventStreamReplaced(v)
    case *events.TempBanReason:                 // IMPORTANT
        rc.Handlers.EventTempBanReason(v)
    case *events.TemporaryBan:                  // IMPORTANT
        rc.Handlers.EventTemporaryBan(v)
    case *events.UnarchiveChatsSetting:
        rc.Handlers.EventUnarchiveChatsSetting(v)
    case *events.UnavailableType:
        rc.Handlers.EventUnavailableType(v)
    case *events.UndecryptableMessage:
        rc.Handlers.EventUndecryptableMessage(v)
    case *events.UnknownCallEvent:
        rc.Handlers.EventUnknownCall(v)
    case *events.UserAbout:
        rc.Handlers.EventUserAbout(v)
    case *events.UserStatusMute:
        rc.Handlers.EventUserStatusMute(v)
    }
}

