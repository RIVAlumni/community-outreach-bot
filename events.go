package main

import (
	"os"

	"go.mau.fi/whatsmeow/types/events"
)

type RIVAClientEvent struct {
    RClient                   *RIVAClient
    DB                        *RIVAClientDB
    Log                       *RIVAClientLog
    SequentialMessageHandlers []SequentialMessageHandlerFunc
    ParallelMessageHandlers   []ParallelMessageHandlerFunc
}

func (ce *RIVAClientEvent) RegisterSequentialHandler(handler SequentialMessageHandlerFunc) {
    ce.SequentialMessageHandlers = append(ce.SequentialMessageHandlers, handler)
}

func (ce *RIVAClientEvent) RegisterParallelHandler(handler ParallelMessageHandlerFunc) {
    ce.ParallelMessageHandlers = append(ce.ParallelMessageHandlers, handler)
}

func (*RIVAClientEvent) New(rClient *RIVAClient, db *RIVAClientDB) *RIVAClientEvent {
    ce := &RIVAClientEvent{
        RClient:                   rClient,
        DB:                        db,
        Log:                       NewRIVAClientLog("RIVABotEvent", "INFO"),
        SequentialMessageHandlers: make([]SequentialMessageHandlerFunc, 0),
        ParallelMessageHandlers:   make([]ParallelMessageHandlerFunc, 0),
    }

    ce.RegisterSequentialHandler(FilterOldMessagesHandler)
    ce.RegisterSequentialHandler(FilterUnsupportedMessagesHandler)
    ce.RegisterSequentialHandler(LogNewMessageHandler)
    ce.RegisterSequentialHandler(SendGreetingMessageHandler)
    ce.RegisterSequentialHandler(AutoEditOutgoingMessageHandler)

    return ce
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

func (ce *RIVAClientEvent) EventCallOffer (evt *events.CallOffer) {
    ce.Log.Infof("Auto-rejecting call from %s (ID: %s)", evt.From, evt.CallID)

    if err := ce.RClient.WMClient.RejectCall(evt.From, evt.CallID); err != nil {
        ce.Log.Errorf("Failed to reject call from %s: %v", evt.From, err)
    }
}

func (ce *RIVAClientEvent) EventCallOfferNotice (evt *events.CallOfferNotice) {
    ce.Log.Infof("Auto-rejecting group call from %s (ID: %s)", evt.From, evt.CallID)

    if err := ce.RClient.WMClient.RejectCall(evt.From, evt.CallID); err != nil {
        ce.Log.Errorf("Failed to reject call from %s: %v", evt.From, err)
    }
}

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
    msg := (*RIVAClientMessage).New(nil, ce.RClient, evt)

    var currentSequenceHandlerIndex int = 0
    var sequencePipelineStopped bool = false
    var runNextSequenceStep func()

    pipelineStopAction := func() {
        sequencePipelineStopped = true
    }

    runNextSequenceStep = func() {
        if sequencePipelineStopped {
            return
        }

        if currentSequenceHandlerIndex >= len(ce.SequentialMessageHandlers) {
            for i, pHandler := range ce.ParallelMessageHandlers {
                go func(idx int, handler ParallelMessageHandlerFunc, msg RIVAClientMessage) {
                    if err := handler(ce.RClient, msg); err != nil {
                        ce.Log.Errorf("Error from parallel handler #%d for message id %s: %v", idx+1, msg.ID, err)
                    }
                }(i, pHandler, msg)
            }

            return
        }

        handlerToExecute := ce.SequentialMessageHandlers[currentSequenceHandlerIndex]
        currentSequenceHandlerIndex++

        actionReturnedByHandler := handlerToExecute(ce.RClient, msg, runNextSequenceStep, pipelineStopAction)
        actionReturnedByHandler()
    }

    runNextSequenceStep()
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
    ce.Log.Infof("Pairing successful: %+v", evt)
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
