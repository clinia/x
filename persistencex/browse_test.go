package persistencex

import (
	"context"
	"testing"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type mockFetcher struct {
	mock.Mock

	data []string
}

func (mf *mockFetcher) Fetch(ctx context.Context, req ListRequest) (*ListResponse[string], error) {
	args := mf.Called(ctx, req)
	return args.Get(0).(*ListResponse[string]), args.Error(1)
}

type browseTestSuite struct {
	suite.Suite
	mf mockFetcher
}

// BeforeTest implements suite.BeforeTest.
func (ts *browseTestSuite) BeforeTest(suiteName string, testName string) {
	ts.mf = mockFetcher{}
}

// TearDownSubTest implements suite.TearDownSubTest.
func (ts *browseTestSuite) TearDownSubTest() {
	ts.mf.AssertExpectations(ts.T())
}

var (
	_ suite.BeforeTest      = (*browseTestSuite)(nil)
	_ suite.TearDownSubTest = (*browseTestSuite)(nil)
)

func TestBrowse(t *testing.T) {
	suite.Run(t, new(browseTestSuite))
}

func (ts *browseTestSuite) TestBrowseSuccesses() {
	data := make([]string, 100)
	for i := 0; i < 100; i++ {
		data[i] = ksuid.New().String()
	}

	ts.Run("should make a single call when all the data fits in a single request", func() {
		ctx := context.Background()
		ts.mf.On("Fetch", ctx, mock.MatchedBy(func(req ListRequest) bool {
			return req.Page == 0 && req.PerPage == 100
		})).Return(&ListResponse[string]{
			Data: data,
			Meta: NewListResponseMeta(0, 100, 100),
		}, nil).Once()

		r, err := Browse(ts.mf.Fetch, ctx, ListRequest{
			Page:    0,
			PerPage: 100,
		})
		ts.NoError(err)
		ts.Equal(data, r)
	})

	ts.Run("should make multiple calls when the data doesn't fit in a single request", func() {
		ctx := context.Background()
		ts.mf.On("Fetch", ctx, mock.MatchedBy(func(req ListRequest) bool {
			return req.Page == 0 && req.PerPage == 50
		})).Return(&ListResponse[string]{
			Data: data[0:50],
			Meta: NewListResponseMeta(0, 50, 100),
		}, nil).Once()
		ts.mf.On("Fetch", ctx, mock.MatchedBy(func(req ListRequest) bool {
			return req.Page == 1 && req.PerPage == 50
		})).Return(&ListResponse[string]{
			Data: data[50:],
			Meta: NewListResponseMeta(1, 50, 100),
		}, nil).Once()

		r, err := Browse(ts.mf.Fetch, ctx, ListRequest{
			Page:    0,
			PerPage: 50,
		})

		ts.NoError(err)
		ts.Equal(data, r)
	})
}
