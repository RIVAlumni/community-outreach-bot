package main

import (
    "os"
    "time"
    "strings"

    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/types"
    "go.mau.fi/whatsmeow/types/events"
    waLog "go.mau.fi/whatsmeow/util/log"
)

const (
    GREETING_COOLDOWN_HOURS = 24
    GREETING_MESSAGE = `
    Thank you for contacting RIVA. An assistant will respond shortly.
    `
)

type RIVAClientEvent struct {
    WMClient *whatsmeow.Client
    DB       *RIVAClientDB
    Log      waLog.Logger
}

func NewRIVAClientEvent(wmClient *whatsmeow.Client, db *RIVAClientDB, logger waLog.Logger) *RIVAClientEvent {
    return &RIVAClientEvent{
        WMClient: wmClient,
        DB:       db,
        Log:      logger,
    }
}

func (ce *RIVAClientEvent) getPhoneNumberFromJID(jid types.JID) string {
    if jid.User == "" {
        return ""
    }

    parts := strings.Split(jid.User, ":")
    return parts[0]
}

func (ce *RIVAClientEvent) EventAppState(evt *events.AppState) {}

func (ce *RIVAClientEvent) EventAppStateSyncComplete(evt *events.AppStateSyncComplete) {}

func (ce *RIVAClientEvent) EventArchive(evt *events.Archive) {}

func (ce *RIVAClientEvent) EventBlocklist(evt *events.Blocklist) {}

func (ce *RIVAClientEvent) EventBlocklistAction(evt *events.BlocklistAction) {}

func (ce *RIVAClientEvent) EventBlocklistChange(evt *events.BlocklistChange) {}

func (ce *RIVAClientEvent) EventBlocklistChangeAction(evt *events.BlocklistChangeAction) {}

func (ce *RIVAClientEvent) EventBusinessName(evt *events.BusinessName) {}

func (ce *RIVAClientEvent) EventCallAccept(evt *events.CallAccept) {}

func (ce *RIVAClientEvent) EventCallOffer (evt *events.CallOffer) {}

func (ce *RIVAClientEvent) EventCallOfferNotice (evt *events.CallOfferNotice) {}

func (ce *RIVAClientEvent) EventCallPreAccept (evt *events.CallPreAccept) {}

func (ce *RIVAClientEvent) EventCallReject (evt *events.CallReject) {}

func (ce *RIVAClientEvent) EventCallRelayLatency (evt *events.CallRelayLatency) {}

func (ce *RIVAClientEvent) EventCallTerminate (evt *events.CallTerminate) {}

func (ce *RIVAClientEvent) EventCallTransport (evt *events.CallTransport) {}

func (ce *RIVAClientEvent) EventChatPresence (evt *events.ChatPresence) {}

func (ce *RIVAClientEvent) EventClearChat (evt *events.ClearChat) {}

func (ce *RIVAClientEvent) EventClientOutdated (evt *events.ClientOutdated) {}

func (ce *RIVAClientEvent) EventConnectFailure (evt *events.ConnectFailure) {}

func (ce *RIVAClientEvent) EventConnectFailureReason (evt *events.ConnectFailureReason) {}

func (ce *RIVAClientEvent) EventConnected (evt *events.Connected) {
    ce.Log.Infof("Successfully connected and authenticated to WhatsApp.")
}

func (ce *RIVAClientEvent) EventContact (evt *events.Contact) {}

func (ce *RIVAClientEvent) EventDecryptFailMode (evt *events.DecryptFailMode) {}

func (ce *RIVAClientEvent) EventDeleteChat (evt *events.DeleteChat) {}

func (ce *RIVAClientEvent) EventDeleteForMe (evt *events.DeleteForMe) {}

func (ce *RIVAClientEvent) EventDisconnected (evt *events.Disconnected) {
    ce.Log.Infof("Disconnected from WhatsApp. Connection closed by WhatsApp.")
}

func (ce *RIVAClientEvent) EventFBMessage (evt *events.FBMessage) {}

func (ce *RIVAClientEvent) EventGroupInfo (evt *events.GroupInfo) {}

func (ce *RIVAClientEvent) EventHistorySync (evt *events.HistorySync) {}

func (ce *RIVAClientEvent) EventIdentityChange (evt *events.IdentityChange) {}

func (ce *RIVAClientEvent) EventJoinedGroup (evt *events.JoinedGroup) {}

func (ce *RIVAClientEvent) EventKeepAliveRestored (evt *events.KeepAliveRestored) {}

func (ce *RIVAClientEvent) EventKeepAliveTimeout (evt *events.KeepAliveTimeout) {}

func (ce *RIVAClientEvent) EventLabelAssociationChat (evt *events.LabelAssociationChat) {}

func (ce *RIVAClientEvent) EventLabelAssociationMessage (evt *events.LabelAssociationMessage) {}

func (ce *RIVAClientEvent) EventLabelEdit (evt *events.LabelEdit) {}

func (ce *RIVAClientEvent) EventLoggedOut (evt *events.LoggedOut) {
    ce.Log.Infof("Logged out. Reason: %s", evt.Reason.String())
    os.Exit(0)
}

func (ce *RIVAClientEvent) EventMarkChatAsRead (evt *events.MarkChatAsRead) {}

func (ce *RIVAClientEvent) EventMediaRetry (evt *events.MediaRetry) {}

func (ce *RIVAClientEvent) EventMediaRetryError (evt *events.MediaRetryError) {}

func (ce *RIVAClientEvent) EventMessage (evt *events.Message) {
    msg := NewRIVAClientMessage(ce.WMClient, evt)
    ce.Log.Infof("New message: %+v", msg)

    if !msg.IsGroup && !msg.IsSentByMe() {
        chatPartnerJIDUser := msg.From

        lastInteraction, found, err := ce.DB.GetLastInteractionTime(chatPartnerJIDUser)
        if err != nil {
            ce.Log.Errorf("Error getting last interaction time for %s: %v. Skipping greeting logic.", chatPartnerJIDUser, err)
        } else {
            shouldSendGreeting := false
            if !found {
                shouldSendGreeting = true
                ce.Log.Infof("No previous interaction record for %s. Sending greeting.", chatPartnerJIDUser)
            } else {
                if time.Since(lastInteraction).Hours() >= GREETING_COOLDOWN_HOURS {
                    shouldSendGreeting = true
                    ce.Log.Infof("Last interaction with %s was at %s. Sending greeting.", chatPartnerJIDUser, lastInteraction.Format(time.RFC3339))
                } else {
                    ce.Log.Infof("Last interaction with %s was at %s. Greeting cooldown active.", chatPartnerJIDUser, lastInteraction.Format(time.RFC3339))
                }
            }

            if shouldSendGreeting {
                err := msg.SendGreetingMessage(msg.From)
                if err != nil {
                    ce.Log.Errorf("Failed to execute sendGreetingMessage for %s: %v", msg.From, err)
                }
            }
        }

        err = ce.DB.UpdateLastInteractionTime(chatPartnerJIDUser, msg.Timestamp)
        if err != nil {
            ce.Log.Errorf("Failed to update last interaction time for %s after processing message: %v", chatPartnerJIDUser, err)
        }
    }
}

func (ce *RIVAClientEvent) EventMute (evt *events.Mute) {}

func (ce *RIVAClientEvent) EventNewsletterJoin (evt *events.NewsletterJoin) {}

func (ce *RIVAClientEvent) EventNewsletterLeave (evt *events.NewsletterLeave) {}

func (ce *RIVAClientEvent) EventNewsletterLiveUpdate (evt *events.NewsletterLiveUpdate) {}

func (ce *RIVAClientEvent) EventNewsletterMessageMeta (evt *events.NewsletterMessageMeta) {}

func (ce *RIVAClientEvent) EventNewsletterMuteChange (evt *events.NewsletterMuteChange) {}

func (ce *RIVAClientEvent) EventOfflineSyncCompleted (evt *events.OfflineSyncCompleted) {}

func (ce *RIVAClientEvent) EventOfflineSyncPreview (evt *events.OfflineSyncPreview) {}

func (ce *RIVAClientEvent) EventPairError (evt *events.PairError) {}

func (ce *RIVAClientEvent) EventPairSuccess (evt *events.PairSuccess) {
    ce.Log.Infof("Pairing successful.")
    ce.Log.Infof("  ID          : %s", evt.ID)
    ce.Log.Infof("  LID         : %s", evt.LID)
    ce.Log.Infof("  BusinessName: %s", evt.BusinessName)
    ce.Log.Infof("  Platform    : %s", evt.Platform)
}

func (ce *RIVAClientEvent) EventPermanentDisconnect (evt *events.PermanentDisconnect) {}

func (ce *RIVAClientEvent) EventPicture (evt *events.Picture) {}

func (ce *RIVAClientEvent) EventPin (evt *events.Pin) {}

func (ce *RIVAClientEvent) EventPresence (evt *events.Presence) {}

func (ce *RIVAClientEvent) EventPrivacySettings (evt *events.PrivacySettings) {}

func (ce *RIVAClientEvent) EventPushName (evt *events.PushName) {}

func (ce *RIVAClientEvent) EventPushNameSetting (evt *events.PushNameSetting) {}

func (ce *RIVAClientEvent) EventQR (evt *events.QR) {}

func (ce *RIVAClientEvent) EventQRScannedWithoutMultidevice (evt *events.QRScannedWithoutMultidevice) {}

func (ce *RIVAClientEvent) EventReceipt (evt *events.Receipt) {}

func (ce *RIVAClientEvent) EventStar (evt *events.Star) {}

func (ce *RIVAClientEvent) EventStreamError (evt *events.StreamError) {}

func (ce *RIVAClientEvent) EventStreamReplaced (evt *events.StreamReplaced) {}

func (ce *RIVAClientEvent) EventTempBanReason (evt *events.TempBanReason) {}

func (ce *RIVAClientEvent) EventTemporaryBan (evt *events.TemporaryBan) {}

func (ce *RIVAClientEvent) EventUnarchiveChatsSetting (evt *events.UnarchiveChatsSetting) {}

func (ce *RIVAClientEvent) EventUnavailableType (evt *events.UnavailableType) {}

func (ce *RIVAClientEvent) EventUndecryptableMessage (evt *events.UndecryptableMessage) {}

func (ce *RIVAClientEvent) EventUnknownCall (evt *events.UnknownCallEvent) {}

func (ce *RIVAClientEvent) EventUserAbout (evt *events.UserAbout) {}

func (ce *RIVAClientEvent) EventUserStatusMute (evt *events.UserStatusMute) {}

