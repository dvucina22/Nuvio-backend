package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/transaction-service/internal/repository"
	"github.com/transaction-service/pkg/models"
)

type CardService struct {
	repo repository.BankCardRepository
	key  []byte
}

func NewCardService(repo repository.BankCardRepository, encryptionKey []byte) *CardService {
	return &CardService{
		repo: repo,
		key:  encryptionKey,
	}
}

func (s *CardService) AddCard(ctx context.Context, userID string, req *models.AddCardRequest) (*models.BankCard, error) {
	if req == nil {
		return nil, models.ErrMissingFields
	}

	if !s.validateCardNumber(req.CardNumber) {
		return nil, models.ErrInvalidCardNumber
	}

	if !s.validateExpiration(req.ExpirationMonth, req.ExpirationYear) {
		return nil, models.ErrCardExpired
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, models.ErrInvalidUserId
	}

	last4 := req.CardNumber[len(req.CardNumber)-4:]
	brand := s.detectCardBrand(req.CardNumber)

	encPAN, iv, err := s.encryptPAN([]byte(req.CardNumber))
	if err != nil {
		return nil, models.ErrEncryptionFailed
	}

	card := &models.BankCard{
		PANEncrypted:    encPAN,
		IV:              iv,
		LastFourDigits:  last4,
		CardBrand:       brand,
		ExpirationMonth: req.ExpirationMonth,
		ExpirationYear:  req.ExpirationYear,
		FullnameOnCard:  req.FullnameOnCard,
		CardName:        req.CardName,
	}

	if err := s.repo.Create(ctx, userUUID, card, req.IsPrimary); err != nil {
		return nil, models.ErrDatabaseOperation
	}

	card.PANEncrypted = nil
	card.IV = nil
	card.IsPrimary = req.IsPrimary

	return card, nil
}

func (s *CardService) GetCards(ctx context.Context, userID string) ([]*models.BankCard, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, models.ErrInvalidUserId
	}

	cards, err := s.repo.FindByUserID(ctx, userUUID)
	if err != nil {
		return nil, models.ErrDatabaseOperation
	}

	return cards, nil
}

func (s *CardService) GetCard(ctx context.Context, userID string, cardID int) (*models.BankCard, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, models.ErrInvalidUserId
	}

	card, err := s.repo.FindByID(ctx, userUUID, cardID)
	if err != nil {
		return nil, models.ErrDatabaseOperation
	}
	if card == nil {
		return nil, models.ErrCardNotFound
	}

	return card, nil
}

func (s *CardService) DeleteCard(ctx context.Context, userID string, cardID int) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return models.ErrInvalidUserId
	}

	exists, err := s.repo.UserHasCard(ctx, userUUID, cardID)
	if err != nil {
		return models.ErrDatabaseOperation
	}
	if !exists {
		return models.ErrCardNotFound
	}

	if err := s.repo.Delete(ctx, userUUID, cardID); err != nil {
		return models.ErrDatabaseOperation
	}

	return nil
}

func (s *CardService) SetPrimaryCard(ctx context.Context, userID string, cardID int) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return models.ErrInvalidUserId
	}

	exists, err := s.repo.UserHasCard(ctx, userUUID, cardID)
	if err != nil {
		return models.ErrDatabaseOperation
	}
	if !exists {
		return models.ErrCardNotFound
	}

	if err := s.repo.SetPrimary(ctx, userUUID, cardID); err != nil {
		return models.ErrDatabaseOperation
	}

	return nil
}

func (s *CardService) GetCardForPayment(ctx context.Context, userID string, cardID int) (*models.PaymentCard, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, models.ErrInvalidUserId
	}

	card, err := s.repo.FindByIDWithSecrets(ctx, userUUID, cardID)
	if err != nil {
		return nil, models.ErrDatabaseOperation
	}
	if card == nil {
		return nil, models.ErrCardNotFound
	}

	if len(card.PANEncrypted) == 0 || len(card.IV) == 0 {
		return nil, models.ErrEncryptionFailed
	}

	pan, err := s.decryptPAN(card.PANEncrypted, card.IV)
	if err != nil {
		return nil, models.ErrEncryptionFailed
	}

	return &models.PaymentCard{
		PAN:             pan,
		ExpirationMonth: card.ExpirationMonth,
		ExpirationYear:  card.ExpirationYear,
		Brand:           card.CardBrand,
		Last4:           card.LastFourDigits,
	}, nil
}

func (s *CardService) encryptPAN(pan []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	iv := make([]byte, gcm.NonceSize())
	_, err = rand.Read(iv)
	if err != nil {
		return nil, nil, err
	}
	enc := gcm.Seal(nil, iv, pan, nil)
	return enc, iv, nil
}

func (s *CardService) validateCardNumber(cardNumber string) bool {
	cardNumber = strings.ReplaceAll(cardNumber, " ", "")
	if len(cardNumber) < 13 || len(cardNumber) > 19 {
		return false
	}
	sum := 0
	alt := false
	for i := len(cardNumber) - 1; i >= 0; i-- {
		n := int(cardNumber[i] - '0')
		if alt {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		alt = !alt
	}
	return sum%10 == 0
}

func (s *CardService) validateExpiration(month, year int) bool {
	if month < 1 || month > 12 {
		return false
	}
	now := time.Now()
	if year < now.Year() || (year == now.Year() && month < int(now.Month())) {
		return false
	}
	return true
}

func (s *CardService) detectCardBrand(cardNumber string) string {
	if strings.HasPrefix(cardNumber, "4") {
		return "visa"
	}
	if strings.HasPrefix(cardNumber, "5") {
		return "mastercard"
	}
	if strings.HasPrefix(cardNumber, "34") || strings.HasPrefix(cardNumber, "37") {
		return "amex"
	}
	return "unknown"
}

func (s *CardService) decryptPAN(enc, iv []byte) (string, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plain, err := gcm.Open(nil, iv, enc, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
