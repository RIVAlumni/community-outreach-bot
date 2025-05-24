package main

import (
    "os"
    "strings"

    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/types"
    "go.mau.fi/whatsmeow/types/events"
    waLog "go.mau.fi/whatsmeow/util/log"
)

type RIVAClient struct {
    WMClient *whatsmeow.Client
    Log      waLog.Logger
}

func NewRIVAClient(wmClient *whatsmeow.Client, logger waLog.Logger) *RIVAClient {
    return &RIVAClient{
        WMClient: wmClient,
        Log:      logger,
    }
}

func (rc *RIVAClient) getPhoneNumberFromJID(jid types.JID) string {
    if jid.User == "" {
        return ""
    }

    parts := strings.Split(jid.User, ":")
    return parts[0]
}

func (rc *RIVAClient) isMessageSentByMe(messageInfo *types.MessageInfo) bool {
    return messageInfo.IsFromMe
}

func (rc *RIVAClient) isMessageReceivedByMe(messageInfo *types.MessageInfo) bool {
    return !messageInfo.IsFromMe
}

func (rc *RIVAClient) EventHandler(evt interface{}) {
    switch v := evt.(type) {
    case *events.AppState:
    case *events.AppStateSyncComplete:
    case *events.Archive:
    case *events.Blocklist:
    case *events.BlocklistAction:
    case *events.BlocklistChange:
    case *events.BlocklistChangeAction:
    case *events.BusinessName:
    case *events.CallAccept:                    // Useful for auto-rejecting calls
    case *events.CallOffer:                     // Useful for auto-rejecting calls
    case *events.CallOfferNotice:               // Useful for auto-rejecting calls
    case *events.CallPreAccept:                 // Useful for auto-rejecting calls
    case *events.CallReject:                    // Useful for auto-rejecting calls
    case *events.CallRelayLatency:              // Useful for auto-rejecting calls
    case *events.CallTerminate:                 // Useful for auto-rejecting calls
    case *events.CallTransport:                 // Useful for auto-rejecting calls
    case *events.ChatPresence:
    case *events.ClearChat:
    case *events.ClientOutdated:
    case *events.ConnectFailure:
    case *events.ConnectFailureReason:
    case *events.Connected:
        rc.Log.Infof("Successfully connected and authenticated to WhatsApp.")
    case *events.Contact:
    case *events.DecryptFailMode:
    case *events.DeleteChat:
    case *events.DeleteForMe:
    case *events.Disconnected:
        rc.Log.Infof("Disconnected from WhatsApp. Connection closed by WhatsApp.")
    case *events.FBMessage:
    case *events.GroupInfo:
    case *events.HistorySync:
    case *events.IdentityChange:
    case *events.JoinedGroup:
    case *events.KeepAliveRestored:
    case *events.KeepAliveTimeout:
    case *events.LabelAssociationChat:
    case *events.LabelAssociationMessage:
    case *events.LabelEdit:
    case *events.LoggedOut:
        rc.Log.Infof("Logged out. Reason: %s", v.Reason.String())
        os.Exit(0)
    case *events.MarkChatAsRead:
    case *events.MediaRetry:
    case *events.MediaRetryError:
    case *events.Message:
        rc.Log.Infof("Received a message!")
        rc.Log.Infof("   ID       : %s", v.Info.ID)
        rc.Log.Infof("   Source   : %s", v.Info.Sender)
        rc.Log.Infof("   Timestamp: %s", v.Info.Timestamp)
        rc.Log.Infof("   IsFromMe : %t", v.Info.IsFromMe)
        rc.Log.Infof("   IsGroup  : %t", v.Info.IsGroup)

        if v.Message.GetConversation() != "" {
            rc.Log.Infof("  Content (Conversation): %s", v.Message.GetConversation())
        } else if v.Message.GetExtendedTextMessage() != nil {
            rc.Log.Infof("  Content (Extended Text): %s", v.Message.GetExtendedTextMessage().GetText())
        } else if imageMsg := v.Message.GetImageMessage(); imageMsg != nil {
            rc.Log.Infof("  Content (Image): Caption: %s", imageMsg.GetCaption())
        } else {
            rc.Log.Infof("  Content (Other type): %s", v.Message.String())
        }
        rc.Log.Infof("------------------------------------------------------")
    case *events.Mute:
    case *events.NewsletterJoin:                // Useless for now
    case *events.NewsletterLeave:               // Useless for now
    case *events.NewsletterLiveUpdate:          // Useless for now
    case *events.NewsletterMessageMeta:         // Useless for now
    case *events.NewsletterMuteChange:          // Useless for now
    case *events.OfflineSyncCompleted:
    case *events.OfflineSyncPreview:
    case *events.PairError:
    case *events.PairSuccess:
        rc.Log.Infof("Pairing successful.")
        rc.Log.Infof("  ID          : %s", v.ID)
        rc.Log.Infof("  LID         : %s", v.LID)
        rc.Log.Infof("  BusinessName: %s", v.BusinessName)
        rc.Log.Infof("  Platform    : %s", v.Platform)
    case *events.PermanentDisconnect:
    case *events.Picture:
    case *events.Pin:
    case *events.Presence:
    case *events.PrivacySettings:
    case *events.PushName:
    case *events.PushNameSetting:
    case *events.QR:
    case *events.QRScannedWithoutMultidevice:
    case *events.Receipt:
    case *events.Star:
    case *events.StreamError:
    case *events.StreamReplaced:
    case *events.TempBanReason:                 // IMPORTANT
    case *events.TemporaryBan:                  // IMPORTANT
    case *events.UnarchiveChatsSetting:
    case *events.UnavailableType:
    case *events.UndecryptableMessage:
    case *events.UnknownCallEvent:
    case *events.UserAbout:
    case *events.UserStatusMute:
    }
}

