package usecase

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image/png"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/identity/ports/outbound"

	"github.com/pquerna/otp/totp"
)

type TwoFactorAuth struct {
	userRepo  outbound.UserRepository
	appConfig *config.APIConfig
}

func NewTwoFactorAuth(userRepo outbound.UserRepository, appConfig *config.APIConfig) *TwoFactorAuth {
	return &TwoFactorAuth{
		userRepo:  userRepo,
		appConfig: appConfig,
	}
}

func (t *TwoFactorAuth) Generate2FA(ctx context.Context, email string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Sage",
		AccountName: email,
	})
	if err != nil {
		return "", "", err
	}

	faSecret, err := utils.Encrypt(key.Secret(), []byte(t.appConfig.AppEncryptionKey))
	if err != nil {
		return "", "", err
	}

	img, err := key.Image(200, 200)
	if err != nil {
		return "", "", err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", "", err
	}
	qr := base64.StdEncoding.EncodeToString(buf.Bytes())

	return faSecret, qr, nil
}

func (t *TwoFactorAuth) Enabled2FA(ctx context.Context, code string, secret string, userID string) error {
	decryptedSecret, err := utils.Decrypt(secret, []byte(t.appConfig.AppEncryptionKey))
	if err != nil {
		return fmt.Errorf("Error decrypting secret: %v", err)
	}
	if !totp.Validate(code, decryptedSecret) {
		return fmt.Errorf("Invalid code")
	}

	if err := t.userRepo.Enable2FA(ctx, secret, userID); err != nil {
		return fmt.Errorf("Error enabling 2FA: %v", err)
	}

	return nil
}

func (t *TwoFactorAuth) Verify2FA(ctx context.Context, code, userID string) error {
	secret, err := t.userRepo.GetTOTPSecret(ctx, userID)
	if err != nil {
		return err
	}

	decryptedSecret, err := utils.Decrypt(secret, []byte(t.appConfig.AppEncryptionKey))
	if err != nil {
		return fmt.Errorf("Error decrypting secret: %v", err)
	}

	if !totp.Validate(code, decryptedSecret) {
		return fmt.Errorf("Invalid 2FA code")
	}

	return nil
}
