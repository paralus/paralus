package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/pkg/crypto"
	rpcv3 "github.com/paralus/paralus/proto/rpc/user"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ApiKeyService is the interface for api key operations
type ApiKeyService interface {
	// create api key
	Create(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error)
	// get by user
	Get(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error)
	// get by key
	GetByKey(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error)
	// delete api key
	Delete(ctx context.Context, req *rpcv3.ApiKeyRequest) (*rpcv3.UserDeleteApiKeysResponse, error)
	// list api keys
	List(ctx context.Context, req *rpcv3.ApiKeyRequest) (*rpcv3.UserListApiKeysResponse, error)
}

// apiKeyService implements ApiKeyService
type apiKeyService struct {
	db *bun.DB
	al *zap.Logger
}

// NewApiKeyService return new api key service
func NewApiKeyService(db *bun.DB, al *zap.Logger) ApiKeyService {
	return &apiKeyService{db, al}
}

func (s *apiKeyService) Create(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
	apikey := &models.ApiKey{
		Name:       req.Username,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
		Trash:      false,
		AccountID:  uuid.MustParse(req.Id),
		Key:        crypto.GenerateSha1Key(),
		Secret:     crypto.GenerateSha256Secret(),
	}

	entity, err := dao.Create(ctx, s.db, apikey)
	if err != nil {
		return nil, err
	}

	if ak, ok := entity.(*models.Group); ok {
		CreateApiKeyAuditEvent(ctx, s.al, AuditActionCreate, ak.ID.String())
	}
	return apikey, nil
}

func (s *apiKeyService) Delete(ctx context.Context, req *rpcv3.ApiKeyRequest) (*rpcv3.UserDeleteApiKeysResponse, error) {
	_, err := s.db.NewUpdate().Model(&models.ApiKey{}).
		Set("trash = ?", true).
		Where("account_id = ?", req.Username).
		Where("key = ?", req.Id).Exec(ctx)
	if err != nil {
		return &rpcv3.UserDeleteApiKeysResponse{}, err
	}

	CreateApiKeyAuditEvent(ctx, s.al, AuditActionDelete, req.Id)
	return nil, err
}

func (s *apiKeyService) List(ctx context.Context, req *rpcv3.ApiKeyRequest) (*rpcv3.UserListApiKeysResponse, error) {
	var apikeys []models.ApiKey
	resp, err := dao.GetX(ctx, s.db, "account_id", req.Username, &apikeys)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if apikeys, ok := resp.(*[]models.ApiKey); ok {
		apiKeyResp := &rpcv3.UserListApiKeysResponse{
			Items: make([]*rpcv3.ApiKeyResponse, 0),
		}
		for _, apikey := range *apikeys {
			apiKeyResp.Items = append(apiKeyResp.Items, &rpcv3.ApiKeyResponse{
				Name:       apikey.Name,
				CreatedAt:  timestamppb.New(apikey.CreatedAt),
				ModifiedAt: timestamppb.New(apikey.ModifiedAt),
				Key:        apikey.Key,
			})
		}
		return apiKeyResp, nil
	}
	return nil, err
}

func (s *apiKeyService) Get(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
	var apikey models.ApiKey
	_, err := dao.GetByName(ctx, s.db, req.Username, &apikey)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &apikey, err
}

func (s *apiKeyService) GetByKey(ctx context.Context, req *rpcv3.ApiKeyRequest) (*models.ApiKey, error) {
	var apikey models.ApiKey
	_, err := dao.GetX(ctx, s.db, "key", req.Id, &apikey)
	if err != nil {
		return nil, err
	}
	return &apikey, err
}
