package pubsubx

import (
	"github.com/clinia/x/errorx"
	"github.com/twmb/franz-go/pkg/kadm"
)

type TruncateReponse struct {
	Topic        string `json:"topic"`
	Partition    int32  `json:"partition"`
	OffsetBefore int64  `json:"offset_before"`
	OffsetAfter  int64  `json:"offset_after"`
	Err          error  `json:"error,omitempty"`
}

// NewTruncateResponse creates a slice of TruncateTopicResult from the given previous offsets and delete records responses.
// It requires the 2 lists are sorted by topic and partition
func NewTruncateResponse(previousOffsets []kadm.Offset, deleteRecordsResponses []kadm.DeleteRecordsResponse) ([]TruncateReponse, error) {
	if len(previousOffsets) != len(deleteRecordsResponses) {
		return nil, errorx.InternalErrorf("previousOffsets and deleteRecordsResponses must have the same length")
	}

	result := make([]TruncateReponse, 0, len(previousOffsets))
	for i, offset := range previousOffsets {
		if offset.Topic != deleteRecordsResponses[i].Topic || offset.Partition != deleteRecordsResponses[i].Partition {
			return nil, errorx.InternalErrorf("previousOffsets and deleteRecordsResponses must be sorted by topic and partition")
		}

		dr := deleteRecordsResponses[i]
		result = append(result, TruncateReponse{
			Topic:        offset.Topic,
			Partition:    offset.Partition,
			OffsetBefore: offset.At,
			OffsetAfter:  dr.LowWatermark,
			Err:          dr.Err,
		})
	}

	return result, nil
}
