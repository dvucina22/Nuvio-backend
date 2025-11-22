package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
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
		return nil, models.ErrInvalidCard
	}
	if !s.validateExpiration(req.ExpirationMonth, req.ExpirationYear) {
		return nil, fmt.Errorf("card expired")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	last4 := req.CardNumber[len(req.CardNumber)-4:]
	brand := s.detectCardBrand(req.CardNumber)
	token := s.generateToken()

	encPAN, iv, err := s.encryptPAN([]byte(req.CardNumber))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt pan: %w", err)
	}

	card := &models.BankCard{
		Token:           token,
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
		return nil, fmt.Errorf("failed to save card: %w", err)
	}

	card.PANEncrypted = nil
	card.IV = nil
	card.IsPrimary = req.IsPrimary

	return card, nil
}

func (s *CardService) GetCards(ctx context.Context, userID string) ([]*models.BankCard, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}
	return s.repo.FindByUserID(ctx, userUUID)
}

func (s *CardService) GetCard(ctx context.Context, userID string, cardID int) (*models.BankCard, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	card, err := s.repo.FindByID(ctx, userUUID, cardID)
	if err != nil {
		return nil, err
	}
	if card == nil {
		return nil, models.ErrCardNotFound
	}

	return card, nil
}

func (s *CardService) DeleteCard(ctx context.Context, userID string, cardID int) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID")
	}
	exists, err := s.repo.UserHasCard(ctx, userUUID, cardID)
	if err != nil {
		return err
	}
	if !exists {
		return models.ErrCardNotFound
	}
	return s.repo.Delete(ctx, userUUID, cardID)
}

func (s *CardService) SetPrimaryCard(ctx context.Context, userID string, cardID int) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID")
	}
	return s.repo.SetPrimary(ctx, userUUID, cardID)
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

func (s *CardService) generateToken() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "card_" + hex.EncodeToString(b)
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
