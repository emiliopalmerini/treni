package bot_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/emiliopalmerini/treni/internal/bot"
	"github.com/emiliopalmerini/treni/internal/domain"
)

func depAt(category string, min int, dest string) domain.Departure {
	return domain.Departure{
		TrainCategory: category,
		ScheduledTime: time.Date(2026, 4, 24, 14, 0, 0, 0, time.UTC).Add(time.Duration(min) * time.Minute),
		Destination:   dest,
	}
}

func TestQueryHandler_multipleTerminiShowsDirectionPicker(t *testing.T) {
	fc := &fakeQuerySvc{
		stations: []domain.Station{{Code: "S01234", Name: "Desio"}},
		departures: []domain.Departure{
			depAt("S9", 5, "Saronno"),
			depAt("S9", 10, "Albairate"),
			depAt("S9", 15, "Saronno"),
		},
	}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnText(bot.NewQueryHandler(fc, 60*time.Minute))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "S9 Desio"))

	if len(sender.messages()) != 0 {
		t.Errorf("expected no plain messages, got %d: %v", len(sender.messages()), sender.messages())
	}
	kbs := sender.sentKeyboards()
	if len(kbs) != 1 {
		t.Fatalf("got %d keyboards, want 1", len(kbs))
	}
	kb := kbs[0]
	// 2 termini + 1 "all" button
	if len(kb.Buttons) != 3 {
		t.Fatalf("got %d buttons, want 3", len(kb.Buttons))
	}
	// First two buttons should encode d: prefix
	for i, b := range kb.Buttons[:2] {
		if !strings.HasPrefix(b.Data, "d:S9:S01234:") {
			t.Errorf("button %d data missing d: prefix: %q", i, b.Data)
		}
	}
	// Last button should be "all" (star)
	if !strings.HasSuffix(kb.Buttons[2].Data, ":*") {
		t.Errorf("last button should be all (*), got %q", kb.Buttons[2].Data)
	}
	// Label assertion: termini appear somewhere in button text
	joined := kb.Buttons[0].Text + "|" + kb.Buttons[1].Text
	if !strings.Contains(joined, "Saronno") || !strings.Contains(joined, "Albairate") {
		t.Errorf("buttons missing terminus labels: %q", joined)
	}
}

func TestQueryHandler_singleTerminusStillListsDirectly(t *testing.T) {
	fc := &fakeQuerySvc{
		stations: []domain.Station{{Code: "S01234", Name: "Desio"}},
		departures: []domain.Departure{
			depAt("S9", 5, "Saronno"),
			depAt("S9", 15, "Saronno"),
		},
	}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnText(bot.NewQueryHandler(fc, 60*time.Minute))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "S9 Desio"))

	if len(sender.sentKeyboards()) != 0 {
		t.Errorf("should not send keyboard when single terminus")
	}
	if len(sender.messages()) != 1 {
		t.Fatalf("got %d messages, want 1", len(sender.messages()))
	}
}

func TestDirectionHandler_filtersAndEdits(t *testing.T) {
	fc := &fakeQuerySvc{
		departures: []domain.Departure{
			depAt("S9", 5, "Saronno"),
			depAt("S9", 10, "Albairate"),
			depAt("S9", 15, "Saronno"),
		},
	}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnCallback("d:", bot.NewDirectionHandler(fc, 60*time.Minute))

	_ = d.Handle(context.Background(), sender,
		newCallbackUpdate(42, 101, "cb-dir", "d:S9:S01234:Saronno"))

	edits := sender.editedMessages()
	if len(edits) != 1 {
		t.Fatalf("got %d edits, want 1", len(edits))
	}
	text := edits[0].Text
	if !strings.Contains(text, "Saronno") {
		t.Errorf("edit missing Saronno: %q", text)
	}
	if strings.Contains(text, "Albairate") {
		t.Errorf("edit should not contain filtered-out terminus: %q", text)
	}

	ans := sender.answered()
	if len(ans) != 1 {
		t.Errorf("expected 1 callback answer")
	}
}

func TestDirectionHandler_allShowsEverything(t *testing.T) {
	fc := &fakeQuerySvc{
		departures: []domain.Departure{
			depAt("S9", 5, "Saronno"),
			depAt("S9", 10, "Albairate"),
		},
	}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnCallback("d:", bot.NewDirectionHandler(fc, 60*time.Minute))

	_ = d.Handle(context.Background(), sender,
		newCallbackUpdate(42, 101, "cb-all", "d:S9:S01234:*"))

	edits := sender.editedMessages()
	if len(edits) != 1 {
		t.Fatalf("got %d edits, want 1", len(edits))
	}
	text := edits[0].Text
	if !strings.Contains(text, "Saronno") || !strings.Contains(text, "Albairate") {
		t.Errorf("'all' should include both termini, got: %q", text)
	}
}

func TestDirectionHandler_prefixMatchHandlesTruncation(t *testing.T) {
	fc := &fakeQuerySvc{
		departures: []domain.Departure{
			depAt("S9", 5, "Milano Porta Garibaldi"),
			depAt("S9", 10, "Albairate"),
		},
	}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnCallback("d:", bot.NewDirectionHandler(fc, 60*time.Minute))

	// Simulate truncated terminus ("Milano Porta" instead of full name)
	_ = d.Handle(context.Background(), sender,
		newCallbackUpdate(42, 101, "cb-trunc", "d:S9:S01234:Milano Porta"))

	edits := sender.editedMessages()
	if len(edits) != 1 {
		t.Fatalf("got %d edits, want 1", len(edits))
	}
	if !strings.Contains(edits[0].Text, "Garibaldi") {
		t.Errorf("prefix match should catch full terminus, got: %q", edits[0].Text)
	}
	if strings.Contains(edits[0].Text, "Albairate") {
		t.Errorf("filter should exclude non-matching terminus")
	}
}

func TestStationPickCallback_multipleTerminiEditsToDirectionPicker(t *testing.T) {
	fc := &fakeQuerySvc{
		departures: []domain.Departure{
			depAt("S9", 5, "Saronno"),
			depAt("S9", 10, "Albairate"),
		},
	}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnCallback("q:", bot.NewCallbackHandler(fc, 60*time.Minute))

	_ = d.Handle(context.Background(), sender,
		newCallbackUpdate(42, 101, "cb-q", "q:S9:S01234"))

	edits := sender.editedMessages()
	if len(edits) != 1 {
		t.Fatalf("got %d edits, want 1", len(edits))
	}
	if len(edits[0].Buttons) == 0 {
		t.Error("station-pick callback with multi-terminus should edit with direction buttons")
	}
}
