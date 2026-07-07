package resource

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type Item map[string]any

type ListRequest struct {
	ResourceID string      `json:"resource_id"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	Query      string      `json:"query,omitempty"`
	ParentID   *string     `json:"parent_id,omitempty"`
	Sort       []SortField `json:"sort,omitempty"`
}

type ListResponse struct {
	Items    []Item `json:"items"`
	Total    int    `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	HasMore  bool   `json:"has_more"`
}

type Store interface {
	List(ctx context.Context, req ListRequest) (ListResponse, error)
	Get(ctx context.Context, resourceID, id string) (Item, error)
	Create(ctx context.Context, resourceID string, values Item) (Item, error)
	Update(ctx context.Context, resourceID, id string, values Item) (Item, error)
	Delete(ctx context.Context, resourceID, id string) error
}

type ResourceDefinition struct {
	Schema Resource
	Store  Store
}

type Registry struct {
	mu        sync.RWMutex
	resources map[string]ResourceDefinition
}

func NewRegistry(definitions ...ResourceDefinition) (*Registry, error) {
	registry := &Registry{resources: map[string]ResourceDefinition{}}
	for _, definition := range definitions {
		if err := registry.Register(definition); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

func (r *Registry) Register(definition ResourceDefinition) error {
	definition.Schema.Normalize()
	if err := definition.Schema.Validate(); err != nil {
		return err
	}
	if definition.Store == nil {
		return fmt.Errorf("resource.%s.store: required", definition.Schema.ID)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.resources == nil {
		r.resources = map[string]ResourceDefinition{}
	}
	if _, exists := r.resources[definition.Schema.ID]; exists {
		return fmt.Errorf("resource.%s: already registered", definition.Schema.ID)
	}
	r.resources[definition.Schema.ID] = definition
	return nil
}

func (r *Registry) Get(resourceID string) (ResourceDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	definition, ok := r.resources[strings.TrimSpace(resourceID)]
	return definition, ok
}

func (r *Registry) Schemas() []Resource {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Resource, 0, len(r.resources))
	for _, definition := range r.resources {
		out = append(out, definition.Schema)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

type Service struct {
	registry *Registry
}

func NewService(registry *Registry) (*Service, error) {
	if registry == nil {
		return nil, fmt.Errorf("resource.registry: required")
	}
	return &Service{registry: registry}, nil
}

func (s *Service) Schemas(ctx context.Context) ([]Resource, error) {
	_ = ctx
	return s.registry.Schemas(), nil
}

func (s *Service) Schema(ctx context.Context, resourceID string) (Resource, error) {
	_ = ctx
	definition, ok := s.registry.Get(resourceID)
	if !ok {
		return Resource{}, fmt.Errorf("resource.%s: not found", resourceID)
	}
	return definition.Schema, nil
}

func (s *Service) List(ctx context.Context, req ListRequest) (ListResponse, error) {
	definition, ok := s.registry.Get(req.ResourceID)
	if !ok {
		return ListResponse{}, fmt.Errorf("resource.%s: not found", req.ResourceID)
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = definition.Schema.List.PageSize
	}
	return definition.Store.List(ctx, req)
}

func (s *Service) Get(ctx context.Context, resourceID, id string) (Item, error) {
	definition, ok := s.registry.Get(resourceID)
	if !ok {
		return nil, fmt.Errorf("resource.%s: not found", resourceID)
	}
	return definition.Store.Get(ctx, resourceID, id)
}

func (s *Service) Create(ctx context.Context, resourceID string, values Item) (Item, error) {
	definition, ok := s.registry.Get(resourceID)
	if !ok {
		return nil, fmt.Errorf("resource.%s: not found", resourceID)
	}
	return definition.Store.Create(ctx, resourceID, values)
}

func (s *Service) Update(ctx context.Context, resourceID, id string, values Item) (Item, error) {
	definition, ok := s.registry.Get(resourceID)
	if !ok {
		return nil, fmt.Errorf("resource.%s: not found", resourceID)
	}
	return definition.Store.Update(ctx, resourceID, id, values)
}

func (s *Service) Delete(ctx context.Context, resourceID, id string) error {
	definition, ok := s.registry.Get(resourceID)
	if !ok {
		return fmt.Errorf("resource.%s: not found", resourceID)
	}
	return definition.Store.Delete(ctx, resourceID, id)
}
