package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/victorinolavida/securities/internal/domainerrors"
)

type Security struct {
	id           uuid.UUID
	publicID     uuid.UUID
	isin         string
	name         string
	securityType SecurityType
	enable       bool
	serie        string
	issueDate    time.Time
	maturityDate time.Time
	nominalValue decimal.Decimal
	currency     Currency
	fx           decimal.Decimal
	createdAt    time.Time
	updatedAt    time.Time
}

func (s Security) ID() uuid.UUID                 { return s.id }
func (s Security) PublicID() uuid.UUID           { return s.publicID }
func (s Security) Enable() bool                  { return s.enable }
func (s Security) ISIN() string                  { return s.isin }
func (s Security) Name() string                  { return s.name }
func (s Security) SecurityType() SecurityType    { return s.securityType }
func (s Security) Serie() string                 { return s.serie }
func (s Security) IssueDate() time.Time          { return s.issueDate }
func (s Security) MaturityDate() time.Time       { return s.maturityDate }
func (s Security) NominalValue() decimal.Decimal { return s.nominalValue }
func (s Security) Currency() Currency            { return s.currency }
func (s Security) Fx() decimal.Decimal           { return s.fx }
func (s Security) CreatedAt() time.Time          { return s.createdAt }
func (s Security) UpdatedAt() time.Time          { return s.updatedAt }

func NewSecurity(
	isin, name, securityType, serie, currency string, issueDate, maturityDate time.Time, nominalValue, fx decimal.Decimal,
) (*Security, error) {
	now := time.Now()
	publicID := uuid.New()

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	if !maturityDate.After(issueDate) && !maturityDate.Equal(issueDate) {
		return nil, domainerrors.ErrInvalidMaturityDate
	}
	if len(isin) != 12 {
		return nil, domainerrors.ErrInvalidISIN
	}
	if len(name) == 0 {
		return nil, domainerrors.ErrEmptyName
	}
	if nominalValue.LessThanOrEqual(decimal.Zero) {
		return nil, domainerrors.ErrNominalValueLessThanZero
	}
	if fx.LessThanOrEqual(decimal.Zero) {
		return nil, domainerrors.ErrFxLessThanZero
	}
	if len(serie) == 0 {
		return nil, domainerrors.ErrEmptySerie
	}

	curr, err := NewCurrency(currency)
	if err != nil {
		return nil, err
	}

	secType, err := NewSecurityType(securityType)
	if err != nil {
		return nil, err
	}

	return &Security{
		id:           id,
		publicID:     publicID,
		issueDate:    issueDate,
		enable:       true,
		maturityDate: maturityDate,
		isin:         isin,
		name:         name,
		nominalValue: nominalValue,
		fx:           fx,
		serie:        serie,
		currency:     curr,
		securityType: secType,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func ReconstructSecurity(
	id, publicID uuid.UUID,
	isin string,
	name string,
	securityType string,
	enable bool,
	serie string,
	issueDate, maturityDate time.Time,
	nominalValue, fx float64,
	currency string,
	createdAt time.Time,
	updatedAt time.Time,
) (*Security, error) {
	secType, err := NewSecurityType(securityType)
	if err != nil {
		return nil, err
	}
	curr, err := NewCurrency(currency)
	if err != nil {
		return nil, err
	}

	return &Security{
		id:           id,
		publicID:     publicID,
		enable:       enable,
		isin:         isin,
		securityType: secType,
		serie:        serie,
		issueDate:    issueDate,
		maturityDate: maturityDate,
		nominalValue: decimal.NewFromFloat(nominalValue),
		fx:           decimal.NewFromFloat(fx),
		currency:     curr,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}, nil
}
