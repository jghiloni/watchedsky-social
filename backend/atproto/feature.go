package atproto

import (
	"encoding/json"

	"github.com/jghiloni/watchedsky-social/backend/features"
	"github.com/jghiloni/watchedsky-social/backend/utils"
)

func (a Alert) ToFeature() features.Feature {
	f := features.Feature{
		ID: a.Id,
		Properties: map[string]any{
			"affectedZones": a.AffectedZones,
			"certainty":     a.Certainty,
			"description":   a.Description,
			"effective":     a.Effective,
			"ends":          a.Ends,
			"event":         a.Event,
			"expires":       a.Expires,
			"headline":      a.Headline,
			"id":            a.Id,
			"instruction":   a.Instruction,
			"messageType":   a.MessageType,
			"onset":         a.Onset,
			"replacedAt":    a.ReplacedAt,
			"replacedBy":    a.ReplacedBy,
			"sender":        a.Sender,
			"senderName":    a.SenderName,
			"sent":          a.Sent,
			"severity":      a.Severity,
			"status":        a.Status,
			"urgency":       a.Urgency,
			"@type":         "wx:Alert",
		},
	}

	if a.Geometry != nil {
		j, err := a.Geometry.MarshalJSON()
		if err != nil {
			panic(err)
		}

		if err = json.Unmarshal(j, &f.Geometry); err != nil {
			panic(err)
		}
	}

	return f
}

func FromFeature(f features.Feature) Alert {
	a := Alert{
		Id:            f.Properties.StringValue("id"),
		AffectedZones: utils.FromAnySlice[string](f.Properties["affectedZones"].([]any)),
		Certainty:     f.Properties.StringValue("certainty"),
		Description:   f.Properties.StringValue("description"),
		Effective:     f.Properties.StringValue("effective"),
		Event:         f.Properties.StringValue("event"),
		Headline:      f.Properties.StringValue("headline"),
		MessageType:   f.Properties.StringValue("messageType"),
		Sender:        f.Properties.StringValue("sender"),
		SenderName:    f.Properties.StringValue("senderName"),
		Sent:          f.Properties.StringValue("sent"),
		Severity:      f.Properties.StringValue("severity"),
		Status:        f.Properties.StringValue("status"),
		Urgency:       f.Properties.StringValue("urgency"),
	}

	if strEnds := f.Properties.StringValue("ends"); strEnds != "" {
		a.Ends = utils.Ptr(strEnds)
	}

	if strExpires := f.Properties.StringValue("expires"); strExpires != "" {
		a.Expires = utils.Ptr(strExpires)
	}

	if strInstruction := f.Properties.StringValue("instruction"); strInstruction != "" {
		a.Instruction = utils.Ptr(strInstruction)
	}

	if strOnset := f.Properties.StringValue("onset"); strOnset != "" {
		a.Onset = utils.Ptr(strOnset)
	}

	if strReplacedAt := f.Properties.StringValue("replacedAt"); strReplacedAt != "" {
		a.ReplacedAt = utils.Ptr(strReplacedAt)
	}

	if strReplacedBy := f.Properties.StringValue("replacedBy"); strReplacedBy != "" {
		a.ReplacedBy = utils.Ptr(strReplacedBy)
	}

	return a
}
