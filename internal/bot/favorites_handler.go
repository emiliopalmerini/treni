package bot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/emiliopalmerini/treni/internal/domain"
	"github.com/emiliopalmerini/treni/internal/state"
	"github.com/go-telegram/bot/models"
)

// FavoritesStore is the storage port. The bot package depends on this
// interface, not on the concrete JSONStore.
type FavoritesStore interface {
	List(chatID int64) []domain.Favorite
	Get(chatID int64, name string) (domain.Favorite, bool)
	Save(chatID int64, fav domain.Favorite) error
	Delete(chatID int64, name string) (bool, error)
}

const (
	favoritesEmptyText = "No favorites yet. Use /save <name> <FROM>: <TO>."

	favRunCallback    = "fr:"
	favDeleteCallback = "fd:"
)

func NewSaveHandler(favs FavoritesStore) Handler {
	return func(ctx context.Context, s Sender, msg *models.Message) error {
		name, from, to, err := parseSaveCommand(msg.Text)
		if err != nil {
			return s.SendMessage(ctx, msg.Chat.ID, saveErrorReply(err))
		}

		_, existed := favs.Get(msg.Chat.ID, name)
		if err := favs.Save(msg.Chat.ID, domain.Favorite{Name: name, From: from, To: to}); err != nil {
			if errors.Is(err, state.ErrFavoriteLimit) {
				return s.SendMessage(ctx, msg.Chat.ID,
					fmt.Sprintf("Favorite limit reached (%d). Delete one with /unsave <name>.", state.MaxFavoritesPerChat))
			}
			log.Printf("favorites save: %v", err)
			return s.SendMessage(ctx, msg.Chat.ID, "Couldn't save favorite. Try again.")
		}
		verb := "Saved"
		if existed {
			verb = "Updated"
		}
		return s.SendMessage(ctx, msg.Chat.ID, fmt.Sprintf("%s '%s': %s: %s.", verb, name, from, to))
	}
}

func NewUnsaveHandler(favs FavoritesStore) Handler {
	return func(ctx context.Context, s Sender, msg *models.Message) error {
		name, err := parseUnsaveCommand(msg.Text)
		if err != nil {
			return s.SendMessage(ctx, msg.Chat.ID, unsaveErrorReply(err))
		}
		deleted, err := favs.Delete(msg.Chat.ID, name)
		if err != nil {
			log.Printf("favorites delete: %v", err)
			return s.SendMessage(ctx, msg.Chat.ID, "Couldn't delete favorite. Try again.")
		}
		if !deleted {
			return s.SendMessage(ctx, msg.Chat.ID, fmt.Sprintf("No favorite named '%s'.", name))
		}
		return s.SendMessage(ctx, msg.Chat.ID, fmt.Sprintf("Deleted '%s'.", name))
	}
}

func NewFavoritesHandler(favs FavoritesStore) Handler {
	return func(ctx context.Context, s Sender, msg *models.Message) error {
		list := favs.List(msg.Chat.ID)
		if len(list) == 0 {
			return s.SendMessage(ctx, msg.Chat.ID, favoritesEmptyText)
		}
		text, buttons := renderFavoritesList(list)
		return s.SendMessageWithButtons(ctx, msg.Chat.ID, text, buttons)
	}
}

func NewFavoritesRunCallback(favs FavoritesStore, query QueryService, window time.Duration) CallbackHandler {
	return func(ctx context.Context, s Sender, cq *models.CallbackQuery) error {
		name := strings.TrimPrefix(cq.Data, favRunCallback)
		chatID := callbackChatID(cq)
		target := messageTarget{chatID: chatID, messageID: callbackMessageID(cq)}

		fav, ok := favs.Get(chatID, name)
		if !ok {
			_ = target.renderText(ctx, s, "That favorite is gone. /favorites to refresh.")
			return s.AnswerCallback(ctx, cq.ID)
		}
		_ = runRouteInto(ctx, s, query, target, fav.From, fav.To, window)
		return s.AnswerCallback(ctx, cq.ID)
	}
}

func NewFavoritesDeleteCallback(favs FavoritesStore) CallbackHandler {
	return func(ctx context.Context, s Sender, cq *models.CallbackQuery) error {
		name := strings.TrimPrefix(cq.Data, favDeleteCallback)
		chatID := callbackChatID(cq)
		deleted, err := favs.Delete(chatID, name)
		if err != nil {
			log.Printf("favorites delete (callback): %v", err)
			return s.AnswerCallback(ctx, cq.ID)
		}
		if !deleted {
			// Silent: the list is already stale. Just dismiss the spinner.
			return s.AnswerCallback(ctx, cq.ID)
		}
		target := messageTarget{chatID: chatID, messageID: callbackMessageID(cq)}
		remaining := favs.List(chatID)
		if len(remaining) == 0 {
			_ = target.renderText(ctx, s, favoritesEmptyText)
			return s.AnswerCallback(ctx, cq.ID)
		}
		text, buttons := renderFavoritesList(remaining)
		_ = target.renderWithButtons(ctx, s, text, buttons)
		return s.AnswerCallback(ctx, cq.ID)
	}
}

// NewAliasHandler lets a plain message `<name>` resolve to a saved
// route and then run the query. Messages containing `:` bypass the
// lookup so the ADR-014 query grammar stays authoritative.
func NewAliasHandler(favs FavoritesStore, next Handler) Handler {
	return func(ctx context.Context, s Sender, msg *models.Message) error {
		text := strings.TrimSpace(msg.Text)
		if strings.Contains(text, ":") {
			return next(ctx, s, msg)
		}
		fav, ok := favs.Get(msg.Chat.ID, text)
		if !ok {
			return next(ctx, s, msg)
		}
		// Rewrite the message into the canonical query grammar and
		// hand it to the existing query handler.
		aliased := *msg
		aliased.Text = fav.From + ": " + fav.To
		return next(ctx, s, &aliased)
	}
}

// runRouteInto is a narrower version of the query-handler flow used by
// the fr: callback: resolve FROM, handle 0/1/many matches, render or
// edit the target message with departures.
func runRouteInto(ctx context.Context, s Sender, svc QueryService, target messageTarget, from, to string, window time.Duration) error {
	stations, err := svc.SearchStations(ctx, from)
	if err != nil {
		log.Printf("SearchStations %q: %v", from, err)
		return target.renderText(ctx, s, upstreamDownMsg)
	}
	if len(stations) == 0 {
		return target.renderText(ctx, s, "No station found for '"+from+"'.")
	}
	if len(stations) > 1 {
		return sendStationPicker(ctx, s, target, stations, to)
	}
	return renderDepartures(ctx, s, svc, target, stations[0], to, window)
}

func renderFavoritesList(list []domain.Favorite) (string, []Button) {
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })

	var sb strings.Builder
	sb.WriteString("Saved routes (this chat):\n\n")
	buttons := make([]Button, 0, len(list)*2)
	for _, fav := range list {
		sb.WriteString(fmt.Sprintf("%s — %s: %s\n", fav.Name, fav.From, fav.To))
		buttons = append(buttons,
			Button{Text: "▶ " + fav.Name, Data: favRunCallback + fav.Name},
			Button{Text: "🗑 " + fav.Name, Data: favDeleteCallback + fav.Name},
		)
	}
	return sb.String(), buttons
}

func saveErrorReply(err error) string {
	switch {
	case errors.Is(err, ErrInvalidFavoriteName):
		return "Invalid name. Use 1–32 chars, no spaces, no ':'."
	case errors.Is(err, ErrSaveUsage):
		return "Usage: /save <name> <FROM>: <TO>."
	default:
		return "Usage: /save <name> <FROM>: <TO>."
	}
}

func unsaveErrorReply(err error) string {
	switch {
	case errors.Is(err, ErrInvalidFavoriteName):
		return "Invalid name. Use 1–32 chars, no spaces, no ':'."
	case errors.Is(err, ErrUnsaveUsage):
		return "Usage: /unsave <name>."
	default:
		return "Usage: /unsave <name>."
	}
}
