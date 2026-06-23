package language

import (
	"context"
	"database/sql"
)

type Options struct{}

type Service struct {
	cfg     LanguageConfig
	enabled map[string]struct{}
	store   *Store
}

func NewService(db *sql.DB, cfg LanguageConfig, _ Options) (*Service, error) {
	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if err := cfg.ensureDefaultInLanguages(); err != nil {
		return nil, err
	}

	return &Service{
		cfg:     cfg,
		enabled: cfg.EnabledSet(),
		store:   NewStore(db, cfg.Schema),
	}, nil
}

func (s *Service) Config() LanguageConfig {
	return s.cfg
}

func (s *Service) DefaultLanguage() string {
	return s.cfg.DefaultLanguage
}

func (s *Service) ValidateCode(code string) error {
	normalized := NormalizeCode(code)
	if normalized == "" {
		return ErrUnknownCatalogCode.With("code", code)
	}

	if err := validateCatalogCode(normalized); err != nil {
		return err
	}

	if _, ok := s.enabled[normalized]; !ok {
		return ErrUnknownLanguage.With("code", normalized)
	}

	return nil
}

func (s *Service) Store() *Store {
	return s.store
}

func (s *Service) ListLanguages(ctx context.Context) ([]Language, error) {
	items, err := s.store.ListLanguages(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]Language, 0, len(items))
	for _, item := range items {
		if _, ok := s.enabled[item.IDLanguage]; ok {
			out = append(out, item)
		}
	}

	return out, nil
}

func (s *Service) GetLanguage(ctx context.Context, code string) (Language, error) {
	if err := s.ValidateCode(code); err != nil {
		return Language{}, err
	}

	item, err := s.store.GetLanguage(ctx, code)
	if err != nil {
		if IsNotFound(err) {
			return Language{}, ErrLanguageNotFound.With("code", NormalizeCode(code))
		}
		return Language{}, err
	}

	return item, nil
}
