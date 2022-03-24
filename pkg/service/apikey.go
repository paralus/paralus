package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/RafayLabs/rcloud-base/internal/dao"
	"github.com/RafayLabs/rcloud-base/internal/models"
	"github.com/RafayLabs/rcloud-base/pkg/crypto"
	rpcv3 "github.com/RafayLabs/rcloud-base/proto/rpc/user"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
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
	Delete(ctx context.Context, req *rpcv3.ApiKeyRequest) (*rpcv3.DeleteUserResponse, error)
	// list api keys
	List(ctx context.Context, req *rpcv3.ApiKeyRequest) (*rpcv3.ApiKeyResponseList, error)
}

// apiKeyService implements ApiKeyService
type apiKeyService struct {
	db *bun.DB
}

// NewApiKeyService return new api key service
func NewApiKeyService(db *bun.DB) ApiKeyService {
	return &apiKeyService{db}
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

	_, err := dao.Create(ctx, s.db, apikey)
	if err != nil {
		return nil, err
	}
	return apikey, nil
}

func (s *apiKeyService) Delete(ctx context.Context, req *rpcv3.ApiKeyRequest) (*rpcv3.DeleteUserResponse, error) {
	_, err := s.db.NewUpdate().Model(&models.ApiKey{}).
		Set("trash = ?", true).
		Where("account_id = ?", req.Username).
		Where("key = ?", req.Id).Exec(ctx)
	return nil, err
}

func (s *apiKeyService) List(ctx context.Context, req *rpcv3.ApiKeyRequest) (*rpcv3.ApiKeyResponseList, error) {
	var apikeys []models.ApiKey
	resp, err := dao.GetX(ctx, s.db, "account_id", req.Username, &apikeys)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if apikeys, ok := resp.(*[]models.ApiKey); ok {
		apiKeyResp := &rpcv3.ApiKeyResponseList{
			Items: make([]*rpcv3.ApiKeyResponse, len(*apikeys)),
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
