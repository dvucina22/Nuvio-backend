package service

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
)

type CloudinaryService struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryService() (*CloudinaryService, error) {
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		return nil, err
	}

	return &CloudinaryService{cld: cld}, nil
}

func generateSignature(params map[string]string, apiSecret string) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var canonicalParts []string
	for _, k := range keys {
		if params[k] != "" {
			canonicalParts = append(canonicalParts, k+"="+params[k])
		}
	}

	canonical := strings.Join(canonicalParts, "&") + apiSecret

	h := sha1.New()
	h.Write([]byte(canonical))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *CloudinaryService) GenerateUploadSignature(publicID string) (map[string]interface{}, error) {

	timestamp := time.Now().Unix()

	params := map[string]string{
		"timestamp":     fmt.Sprintf("%d", timestamp),
		"upload_preset": "profile_pictures",
		"folder":        "user_profiles",
	}

	if publicID != "" {
		params["public_id"] = publicID
	}

	signature := generateSignature(params, os.Getenv("CLOUDINARY_API_SECRET"))

	return map[string]interface{}{
		"signature":    signature,
		"timestamp":    timestamp,
		"cloudName":    os.Getenv("CLOUDINARY_CLOUD_NAME"),
		"apiKey":       os.Getenv("CLOUDINARY_API_KEY"),
		"uploadPreset": "profile_pictures",
		"folder":       "user_profiles",
		"publicId":     publicID,
	}, nil
}
