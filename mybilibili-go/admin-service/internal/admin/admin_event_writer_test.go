package admin

import (
	"context"
	"errors"
	"testing"
)

type fakeEventWriter struct {
	statusEvents  []statusEventCall
	processEvents []processEventCall
	editVersions  []editVersionCall
}

type statusEventCall struct {
	manuscriptID int64
	fromStatus   int32
	toStatus     int32
	action       string
}

type processEventCall struct {
	videoID       int64
	manuscriptID  int64
	fromStatus    int32
	toStatus      int32
	stage         string
	progress      int
}

type editVersionCall struct {
	manuscriptID int64
	userID       int64
	changed      string
}

func (f *fakeEventWriter) RecordStatusEvent(_ context.Context, manuscriptID, _ int64, fromStatus, toStatus int32, action, _ string, _ int64, _ string) error {
	f.statusEvents = append(f.statusEvents, statusEventCall{manuscriptID, fromStatus, toStatus, action})
	return nil
}

func (f *fakeEventWriter) RecordVideoProcessEvent(_ context.Context, videoID, manuscriptID int64, fromStatus, toStatus int32, stage string, progress int) error {
	f.processEvents = append(f.processEvents, processEventCall{videoID, manuscriptID, fromStatus, toStatus, stage, progress})
	return nil
}

func (f *fakeEventWriter) RecordEditVersion(_ context.Context, manuscriptID, userID int64, _, _, changed string) error {
	f.editVersions = append(f.editVersions, editVersionCall{manuscriptID, userID, changed})
	return nil
}

func errWriter() *fakeEventWriter {
	return &fakeEventWriter{}
}

var _ ManuscriptEventWriter = (*fakeEventWriter)(nil)

func TestSetManuscriptStatusWritesStatusEvent(t *testing.T) {
	f := errWriter()
	h := &ManuscriptAdminHandler{db: nil, events: f}

	err := h.eventWriter().RecordStatusEvent(context.Background(), 7, 1, 0, 3, "PUBLISH", "ADMIN", 2, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.statusEvents) != 1 {
		t.Fatalf("want 1 status event, got %d", len(f.statusEvents))
	}
	evt := f.statusEvents[0]
	if evt.manuscriptID != 7 || evt.fromStatus != 0 || evt.toStatus != 3 || evt.action != "PUBLISH" {
		t.Fatalf("unexpected event: %+v", evt)
	}
}

func TestTriggerVideoProcessWritesProcessEvent(t *testing.T) {
	f := errWriter()
	h := &ManuscriptAdminHandler{db: nil, events: f}

	err := h.eventWriter().RecordVideoProcessEvent(context.Background(), 11, 7, 0, 1, "TRANSCODING", 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.processEvents) != 1 {
		t.Fatalf("want 1 process event, got %d", len(f.processEvents))
	}
	evt := f.processEvents[0]
	if evt.videoID != 11 || evt.manuscriptID != 7 || evt.toStatus != 1 || evt.stage != "TRANSCODING" {
		t.Fatalf("unexpected event: %+v", evt)
	}
}

func TestReviewWritesEditVersion(t *testing.T) {
	f := errWriter()
	h := &ManuscriptAdminHandler{db: nil, events: f}

	err := h.eventWriter().RecordEditVersion(context.Background(), 7, 1, "", "approved", "review_status,status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.editVersions) != 1 {
		t.Fatalf("want 1 edit version, got %d", len(f.editVersions))
	}
	if f.editVersions[0].manuscriptID != 7 || f.editVersions[0].changed != "review_status,status" {
		t.Fatalf("unexpected edit version: %+v", f.editVersions[0])
	}
}

var _ = errors.New
