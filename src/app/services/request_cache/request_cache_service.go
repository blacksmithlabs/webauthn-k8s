package request_cache

import (
	"context"
	"encoding/json"
	"fmt"

	"blacksmithlabs.dev/webauthn-k8s/app/cache"
	"blacksmithlabs.dev/webauthn-k8s/app/config"
	"github.com/go-webauthn/webauthn/webauthn"
)

type RequestCacheService struct {
	ctx    context.Context
	client *cache.CacheClient
	Nil    error
}

type RequestInfo struct {
	UserId      int64
	SessionData *webauthn.SessionData
}

var (
	cacheTimeout = config.GetSessionTimeout()
)

func New(ctx context.Context) *RequestCacheService {
	client := cache.ConnectCache()
	return &RequestCacheService{
		ctx:    ctx,
		client: client,
		Nil:    cache.Nil,
	}
}

func (s *RequestCacheService) GetRequestCache(requestId string) (*RequestInfo, error) {
	requestJson, err := s.client.Get(s.ctx, requestId).Bytes()
	if err == cache.Nil {
		return nil, s.Nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get request cache: %w", err)
	}

	requestInfo := RequestInfo{}
	if err = json.Unmarshal(requestJson, &requestInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request cache: %v", err)
	}

	return &requestInfo, nil
}

func (s *RequestCacheService) SetRequestCache(requestId string, requestInfo *RequestInfo) error {
	requestJson, err := json.Marshal(requestInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal request cache: %w", err)
	}

	if err = s.client.SetEx(s.ctx, requestId, requestJson, cacheTimeout).Err(); err != nil {
		return fmt.Errorf("failed to set request cache: %w", err)
	}

	return nil
}

func (s *RequestCacheService) DeleteRequestCache(requestId string) error {
	if err := s.client.Del(s.ctx, requestId).Err(); err != nil {
		return fmt.Errorf("failed to delete request cache: %w", err)
	}

	return nil
}
