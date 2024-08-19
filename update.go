package telebot

import "strings"

// Update object represents an incoming update.
type Update struct {
	ID int `json:"update_id"`

	Message                 *Message                 `json:"message,omitempty"`
	EditedMessage           *Message                 `json:"edited_message,omitempty"`
	ChannelPost             *Message                 `json:"channel_post,omitempty"`
	EditedChannelPost       *Message                 `json:"edited_channel_post,omitempty"`
	MessageReaction         *MessageReaction         `json:"message_reaction"`
	MessageReactionCount    *MessageReactionCount    `json:"message_reaction_count"`
	Callback                *Callback                `json:"callback_query,omitempty"`
	Query                   *Query                   `json:"inline_query,omitempty"`
	InlineResult            *InlineResult            `json:"chosen_inline_result,omitempty"`
	ShippingQuery           *ShippingQuery           `json:"shipping_query,omitempty"`
	PreCheckoutQuery        *PreCheckoutQuery        `json:"pre_checkout_query,omitempty"`
	Poll                    *Poll                    `json:"poll,omitempty"`
	PollAnswer              *PollAnswer              `json:"poll_answer,omitempty"`
	MyChatMember            *ChatMemberUpdate        `json:"my_chat_member,omitempty"`
	ChatMember              *ChatMemberUpdate        `json:"chat_member,omitempty"`
	ChatJoinRequest         *ChatJoinRequest         `json:"chat_join_request,omitempty"`
	Boost                   *BoostUpdated            `json:"chat_boost"`
	BoostRemoved            *BoostRemoved            `json:"removed_chat_boost"`
	BusinessConnection      *BusinessConnection      `json:"business_connection"`
	BusinessMessage         *Message                 `json:"business_message"`
	EditedBusinessMessage   *Message                 `json:"edited_business_message"`
	DeletedBusinessMessages *BusinessMessagesDeleted `json:"deleted_business_messages"`
}

// ProcessUpdate processes a single incoming update.
// A started bot calls this function automatically.
func (b *Bot) ProcessUpdate(u Update) {
	handler := b.SelectHandlerForUpdate(&u)
	if handler == nil {
		handler = b.getHandler(OnAny)
	}
	if handler == nil {
		// noop handler with global middleware
		handler = b.fallbackHandler
	}
	c := b.NewContext(u)
	b.runHandler(handler, c)
}

// SelectHandlerForUpdate selects the best handler for update
// The update may be modified
// Return nil if no handler registered
func (b *Bot) SelectHandlerForUpdate(u *Update) HandlerFunc {
	if u.Message != nil {
		m := u.Message

		if m.PinnedMessage != nil {
			return b.getHandler(OnPinned)
		}

		// Escape malicious messages
		m.Text = strings.TrimLeft(m.Text, "\a")

		if m.Text != "" {
			match := cmdRx.FindAllStringSubmatch(m.Text, -1)
			if match != nil {
				// Syntax: "</command>@<bot> <payload>"
				command, botName := match[0][1], match[0][3]

				if botName != "" && !strings.EqualFold(b.Me.Username, botName) {
					return nil
				}

				m.Command = command
				m.Payload = match[0][5]
				if h := b.getHandler(command); h != nil {
					return h
				}

				if h := b.getHandler(OnCommand); h != nil {
					return h
				}
			}

			// 1:1 satisfaction
			if h := b.getHandler(m.Text); h != nil {
				return h
			}

			return b.getHandler(OnText)
		}

		if h, isMedia := b.getMediaHandler(m); isMedia {
			return h
		}

		if m.Contact != nil {
			return b.getHandler(OnContact)
		}
		if m.Location != nil {
			return b.getHandler(OnLocation)
		}
		if m.Venue != nil {
			return b.getHandler(OnVenue)
		}
		if m.Game != nil {
			return b.getHandler(OnGame)
		}
		if m.Dice != nil {
			return b.getHandler(OnDice)
		}
		if m.Invoice != nil {
			return b.getHandler(OnInvoice)
		}
		if m.Payment != nil {
			return b.getHandler(OnPayment)
		}
		if m.RefundedPayment != nil {
			return b.getHandler(OnRefund)
		}
		if m.TopicCreated != nil {
			return b.getHandler(OnTopicCreated)
		}
		if m.TopicReopened != nil {
			return b.getHandler(OnTopicReopened)
		}
		if m.TopicClosed != nil {
			return b.getHandler(OnTopicClosed)
		}
		if m.TopicEdited != nil {
			return b.getHandler(OnTopicEdited)
		}
		if m.GeneralTopicHidden != nil {
			return b.getHandler(OnGeneralTopicHidden)
		}
		if m.GeneralTopicUnhidden != nil {
			return b.getHandler(OnGeneralTopicUnhidden)
		}
		if m.WriteAccessAllowed != nil {
			return b.getHandler(OnWriteAccessAllowed)
		}

		wasAdded := (m.UserJoined != nil && m.UserJoined.ID == b.Me.ID) ||
			(m.UsersJoined != nil && isUserInList(b.Me, m.UsersJoined))
		if m.GroupCreated || m.SuperGroupCreated || wasAdded {
			return b.getHandler(OnAddedToGroup)
		}

		if m.UserJoined != nil || m.UsersJoined != nil {
			return b.getHandler(OnUserJoined)
		}

		if m.UserLeft != nil {
			b.getHandler(OnUserLeft)
		}

		if m.UserShared != nil {
			b.getHandler(OnUserShared)
		}
		if m.ChatShared != nil {
			b.getHandler(OnChatShared)
		}

		if m.NewGroupTitle != "" {
			b.getHandler(OnNewGroupTitle)
		}
		if m.NewGroupPhoto != nil {
			b.getHandler(OnNewGroupPhoto)
		}
		if m.GroupPhotoDeleted {
			b.getHandler(OnGroupPhotoDeleted)
		}

		if m.GroupCreated {
			b.getHandler(OnGroupCreated)
		}
		if m.SuperGroupCreated {
			b.getHandler(OnSuperGroupCreated)
		}
		if m.ChannelCreated {
			b.getHandler(OnChannelCreated)
		}

		if m.MigrateTo != 0 {
			m.MigrateFrom = m.Chat.ID
			b.getHandler(OnMigration)
		}

		if m.VideoChatStarted != nil {
			b.getHandler(OnVideoChatStarted)
		}
		if m.VideoChatEnded != nil {
			b.getHandler(OnVideoChatEnded)
		}
		if m.VideoChatParticipants != nil {
			b.getHandler(OnVideoChatParticipants)
		}
		if m.VideoChatScheduled != nil {
			b.getHandler(OnVideoChatScheduled)
		}

		if m.WebAppData != nil {
			b.getHandler(OnWebApp)
		}

		if m.ProximityAlert != nil {
			b.getHandler(OnProximityAlert)
		}
		if m.AutoDeleteTimer != nil {
			b.getHandler(OnAutoDeleteTimer)
		}
	}

	if u.EditedMessage != nil {
		b.getHandler(OnEdited)
	}

	if u.ChannelPost != nil {
		m := u.ChannelPost

		if m.PinnedMessage != nil {
			b.getHandler(OnPinned)
		}

		b.getHandler(OnChannelPost)
	}

	if u.EditedChannelPost != nil {
		b.getHandler(OnEditedChannelPost)
	}

	if u.Callback != nil {
		if data := u.Callback.Data; data != "" && data[0] == '\f' {
			match := cbackRx.FindAllStringSubmatch(data, -1)
			if match != nil {
				unique, payload := match[0][1], match[0][3]
				if h := b.getHandler("\f" + unique); h != nil {
					u.Callback.Unique = unique
					u.Callback.Data = payload
					return h
				}
			}
		}

		return b.getHandler(OnCallback)
	}

	if u.Query != nil {
		return b.getHandler(OnQuery)
	}

	if u.InlineResult != nil {
		return b.getHandler(OnInlineResult)
	}

	if u.ShippingQuery != nil {
		return b.getHandler(OnShipping)
	}

	if u.PreCheckoutQuery != nil {
		return b.getHandler(OnCheckout)
	}

	if u.Poll != nil {
		return b.getHandler(OnPoll)
	}
	if u.PollAnswer != nil {
		return b.getHandler(OnPollAnswer)
	}

	if u.MyChatMember != nil {
		return b.getHandler(OnMyChatMember)
	}
	if u.ChatMember != nil {
		return b.getHandler(OnChatMember)
	}
	if u.ChatJoinRequest != nil {
		return b.getHandler(OnChatJoinRequest)
	}

	if u.Boost != nil {
		return b.getHandler(OnBoost)
	}
	if u.BoostRemoved != nil {
		return b.getHandler(OnBoostRemoved)
	}

	if u.BusinessConnection != nil {
		return b.getHandler(OnBusinessConnection)
	}
	if u.BusinessMessage != nil {
		return b.getHandler(OnBusinessMessage)
	}
	if u.EditedBusinessMessage != nil {
		return b.getHandler(OnEditedBusinessMessage)
	}
	if u.DeletedBusinessMessages != nil {
		return b.getHandler(OnDeletedBusinessMessages)
	}

	return nil
}

func (b *Bot) getHandler(end string) HandlerFunc {
	return b.handlers[end]
}

func (b *Bot) getMediaHandler(m *Message) (h HandlerFunc, isMedia bool) {
	switch {
	case m.Photo != nil:
		h = b.getHandler(OnPhoto)
	case m.Voice != nil:
		h = b.getHandler(OnVoice)
	case m.Audio != nil:
		h = b.getHandler(OnAudio)
	case m.Animation != nil:
		h = b.getHandler(OnAnimation)
	case m.Document != nil:
		h = b.getHandler(OnDocument)
	case m.Sticker != nil:
		h = b.getHandler(OnSticker)
	case m.Video != nil:
		h = b.getHandler(OnVideo)
	case m.VideoNote != nil:
		h = b.getHandler(OnVideoNote)
	default:
		return nil, false
	}

	isMedia = true

	if h == nil { // no specific media type handler, try general
		h = b.getHandler(OnMedia)
	}

	return
}

func (b *Bot) runHandler(h HandlerFunc, c Context) {
	f := func() {
		if err := h(c); err != nil {
			b.OnError(err, c)
		}
	}
	if b.synchronous {
		f()
	} else {
		go f()
	}
}

func isUserInList(user *User, list []User) bool {
	for _, user2 := range list {
		if user.ID == user2.ID {
			return true
		}
	}
	return false
}
