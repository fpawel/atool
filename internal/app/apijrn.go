package app

import (
	"context"
	"github.com/fpawel/atool/internal/pkg/logjrn"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/atool/internal/workgui"
)

type journalSvc struct{}

var _ api.JournalService = &journalSvc{}

func (journalSvc) DeleteDays(_ context.Context, days []string) error {
	return workgui.Journal.DeleteDays(days)
}

func (journalSvc) ListEntriesIDsOfDay(_ context.Context, day string) ([]int64, error) {
	xs, err := workgui.Journal.GetEntriesIDsOfDay(day)
	if len(xs) == 0 {
		xs = make([]int64, 0)
	}
	return xs, err
}

func (journalSvc) GetEntryByID(_ context.Context, entryID int64) (r *apitypes.JournalEntry, err error) {
	x := &logjrn.Entry{
		EntryID: entryID,
	}
	if err := workgui.Journal.GetEntryByID(x); err != nil {
		return nil, err
	}
	return &apitypes.JournalEntry{
		EntryID:  x.EntryID,
		StoredAt: x.StoredAt.Format("15:04:05"),
		Indent:   int64(x.Indent),
		Ok:       x.Ok,
		Text:     x.Text,
		Stack:    x.Stack,
	}, nil
}

func (journalSvc) ListDays(context.Context) (days []string, err error) {
	days, err = workgui.Journal.ListDays()
	return
}
func (journalSvc) ListEntriesOfDay(_ context.Context, strDay string) ([]*apitypes.JournalEntry, error) {
	entries, err := workgui.Journal.GetEntriesOfDay(strDay)
	if err != nil {
		return nil, err
	}
	var result []*apitypes.JournalEntry
	for _, x := range entries {
		result = append(result, &apitypes.JournalEntry{
			EntryID:  x.EntryID,
			StoredAt: x.StoredAt.Format("15:04:05"),
			Indent:   int64(x.Indent),
			Ok:       x.Ok,
			Text:     x.Text,
			Stack:    x.Stack,
		})
	}
	if len(result) == 0 {
		result = make([]*apitypes.JournalEntry, 0)
	}
	return result, nil
}
