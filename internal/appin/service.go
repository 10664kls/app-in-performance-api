package appin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	azidentity "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	graph "github.com/microsoftgraph/msgraph-sdk-go"
	core "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/sites"
	"go.uber.org/zap"
)

type Service struct {
	client *graph.GraphServiceClient
	zlog   *zap.Logger
	siteID string
	listID string
}

type Config struct {
	Zlog     *zap.Logger
	TenantID string
	ClientID string
	Secret   string
	SiteID   string
	ListID   string
	Scopes   []string
}

func NewService(_ context.Context, config *Config) (*Service, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	cred, err := azidentity.NewClientSecretCredential(config.TenantID, config.ClientID, config.Secret, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	client, err := graph.NewGraphServiceClientWithCredentials(cred, config.Scopes)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &Service{
		client: client,
		siteID: config.SiteID,
		listID: config.ListID,
		zlog:   config.Zlog,
	}, nil
}

type ListAppInResult struct {
	AppIns []*AppIn `json:"appIns"`
}

func (s *Service) ListAppIns(ctx context.Context, q *Query) (*ListAppInResult, error) {
	as, err := s.listAppIns(ctx, q)
	if err != nil {
		return nil, err
	}

	return &ListAppInResult{
		AppIns: as,
	}, nil
}

func (s *Service) GetOverview(ctx context.Context, q *Query) (*Overview, error) {
	as, err := s.listAppIns(ctx, q)
	if err != nil {
		return nil, err
	}

	return newOverview(as), nil
}

func (s *Service) listAppIns(ctx context.Context, q *Query) ([]*AppIn, error) {
	zlog := s.zlog.With(
		zap.String("method", "listAppIns"),
		zap.Any("query", q),
	)

	as := make([]*AppIn, 0)
	config := newReqConfig(q)

	res, err := s.client.Sites().
		BySiteId(s.siteID).
		Lists().
		ByListId(s.listID).
		Items().
		Get(ctx, config)
	if err != nil {
		zlog.Error("failed to get list items", zap.Error(err))
		return nil, err
	}

	pager, err := core.NewPageIterator[*models.ListItem](res, s.client.GetAdapter(), models.CreateListItemCollectionResponseFromDiscriminatorValue)
	if err != nil {
		zlog.Error("failed to create page iterator", zap.Error(err))
		return nil, err
	}

	if err := pager.Iterate(ctx, func(l *models.ListItem) bool {
		if l.GetFields() == nil {
			return false
		}

		byt, err := json.Marshal(l.GetFields().GetAdditionalData())
		if err != nil {
			zlog.
				With(
					zap.Any("fields", l.GetFields().GetAdditionalData()),
				).Error("failed to marshal fields", zap.Error(err))
			return false
		}

		a := new(rawAppIn)
		if err := json.Unmarshal(byt, a); err != nil {
			zlog.
				With(
					zap.String("byt", string(byt)),
				).Error("failed to unmarshal fields", zap.Error(err))
			return false
		}

		as = append(as, newAppInFromRawAppIn(a))
		return true
	}); err != nil {
		zlog.Error("failed to iterate page", zap.Error(err))
		return nil, err
	}

	return as, nil
}

func newQueryParams(q *Query) *sites.ItemListsItemItemsRequestBuilderGetQueryParameters {
	return &sites.ItemListsItemItemsRequestBuilderGetQueryParameters{
		Expand: []string{
			`fields($select=Created,Title,LOFacility,ServiceType,CustomerType,Gender,ENGfullname,Status,CompletedDateTime,Creditamount,Instalmentperiod,AssignedTo,Author)`,
		},
		Filter: to.Ptr(q.String()),
		Orderby: []string{
			"fields/Created desc",
		},
		Top: to.Ptr[int32](500),
	}
}

func newReqConfig(q *Query) *sites.ItemListsItemItemsRequestBuilderGetRequestConfiguration {
	return &sites.ItemListsItemItemsRequestBuilderGetRequestConfiguration{
		QueryParameters: newQueryParams(q),
	}
}

type Query struct {
	CreatedAfter  time.Time `json:"createdAfter" query:"createdAfter"`
	CreatedBefore time.Time `json:"createdBefore" query:"createdBefore"`
	Product       string    `json:"product" query:"product"`
}

func (q *Query) String() string {
	var s string
	if q.Product != "" {
		s = "fields/ServiceType eq '" + q.Product + "' and "
	}

	if !q.CreatedAfter.IsZero() && !q.CreatedBefore.IsZero() {
		s += fmt.Sprintf(`fields/Created ge '%s' and fields/Created le '%s'`, q.CreatedAfter.Format(time.RFC3339), q.CreatedBefore.Format(time.RFC3339))
	}

	if !q.CreatedAfter.IsZero() && q.CreatedBefore.IsZero() {
		s += fmt.Sprintf(`fields/Created ge '%s'`, q.CreatedAfter.Format(time.RFC3339))
	}

	if q.CreatedAfter.IsZero() && !q.CreatedBefore.IsZero() {
		s += fmt.Sprintf(`fields/Created le '%s'`, q.CreatedBefore.Format(time.RFC3339))
	}

	// if createdAfter and createdBefore are both zero then we want to query for the last month
	if q.CreatedAfter.IsZero() && q.CreatedBefore.IsZero() {
		monthAgo := time.Now().AddDate(0, -1, 0)
		s += fmt.Sprintf("fields/Created ge '%s'", monthAgo.Format(time.RFC3339))
	}

	return strings.TrimSuffix(s, " and ")
}

type rawAppIn struct {
	LONumber           string     `json:"LOFacility"`
	Product            string     `json:"ServiceType"`
	Type               string     `json:"CustomerType"`
	Prename            string     `json:"Gender"`
	DisplayName        string     `json:"Title"`
	DisplayNameEnglish string     `json:"ENGfullname"`
	Status             string     `json:"Status"`
	FinanceAmount      string     `json:"Creditamount"`
	Term               string     `json:"Instalmentperiod"`
	Executor           string     `json:"AssignedTo"`
	CreatedBy          string     `json:"Author"`
	CompletedAt        *time.Time `json:"CompletedDateTime"`
	CreatedAt          time.Time  `json:"Created"`
}

type AppIn struct {
	Number             string     `json:"number"`
	Product            string     `json:"product"`
	Type               string     `json:"type"`
	Prename            string     `json:"prename"`
	DisplayName        string     `json:"displayName"`
	DisplayNameEnglish string     `json:"displayNameEnglish"`
	Status             string     `json:"status"`
	FinanceAmount      string     `json:"financeAmount"`
	Term               string     `json:"term"`
	Executor           string     `json:"executor"`
	CreatedBy          string     `json:"createdBy"`
	CompletedAt        *time.Time `json:"completedAt"`
	CreatedAt          time.Time  `json:"createdAt"`
}

func newAppInFromRawAppIn(a *rawAppIn) *AppIn {
	return &AppIn{
		Number:             a.LONumber,
		Product:            a.Product,
		Type:               a.Type,
		Prename:            a.Prename,
		DisplayName:        a.DisplayName,
		DisplayNameEnglish: a.DisplayNameEnglish,
		Status:             a.Status,
		FinanceAmount:      a.FinanceAmount,
		Term:               a.Term,
		Executor:           a.Executor,
		CreatedBy:          a.CreatedBy,
		CompletedAt:        a.CompletedAt,
		CreatedAt:          a.CreatedAt,
	}
}
