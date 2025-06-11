package logrusx

import "github.com/sirupsen/logrus"

type requestIdHook struct {
	rIdCtxKey   interface{}
	rIdFieldKey string
}

var _ logrus.Hook = (*requestIdHook)(nil)

func NewRequestIdHook(requestIdContextKey interface{}, requestIdFieldKey string) *requestIdHook {
	return &requestIdHook{
		rIdCtxKey:   requestIdContextKey,
		rIdFieldKey: requestIdFieldKey,
	}
}

func (irh *requestIdHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (rih *requestIdHook) Fire(entry *logrus.Entry) error {
	defer func() {
		// Nullify panic to prevent having this hook break a request
		recover() //nolint:errcheck,gosec
	}()
	if entry == nil || entry.Context == nil || entry.Data == nil {
		return nil
	}
	requestId := entry.Context.Value(rih.rIdCtxKey)
	if requestId == nil {
		return nil
	} else {
		entry.Data[rih.rIdFieldKey] = requestId
	}

	return nil
}
