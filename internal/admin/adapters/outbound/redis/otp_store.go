package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"sage-backend/internal/admin/domain"
	"sage-backend/internal/admin/ports/outbound"

	"github.com/redis/go-redis/v9"
)

type OTPPayload struct {
	AdminID    string `json:"admin_id"`
	OTP        string `json:"otp"`
	RememberMe bool   `json:"remember_me"`
	ClientIP   string `json:"client_ip"`
	Attempts   int    `json:"attempts"`
}

type RedisOTPStore struct {
	client *redis.Client
}

func NewRedisOTPStore(client *redis.Client) outbound.OTPStore {
	return &RedisOTPStore{client: client}
}

var _ outbound.OTPStore = (*RedisOTPStore)(nil)

func otpKey(adminID string) string {
	return fmt.Sprintf("admin:otp:%s", adminID)
}

func (s *RedisOTPStore) SaveOTP(ctx context.Context, data outbound.OTPPendingData, ttl time.Duration) error {
	payload := OTPPayload{
		AdminID:    data.AdminID,
		OTP:        data.OTP,
		RememberMe: data.RememberMe,
		ClientIP:   data.ClientIP,
		Attempts:   0,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, otpKey(data.AdminID), raw, ttl).Err()
}

func (s *RedisOTPStore) VerifyAndConsumeOTP(ctx context.Context, adminID string, otp string) (*outbound.OTPPendingData, error) {
	key := otpKey(adminID)
	raw, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrInvalidOTP
		}
		return nil, err
	}

	var payload OTPPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	if payload.OTP != otp {
		payload.Attempts++
		if payload.Attempts >= 3 {
			_ = s.client.Del(ctx, key).Err()
			return nil, domain.ErrInvalidOTP
		}
		// update attempts with remaining TTL
		ttl, _ := s.client.TTL(ctx, key).Result()
		if ttl > 0 {
			if updated, err := json.Marshal(payload); err == nil {
				_ = s.client.Set(ctx, key, updated, ttl).Err()
			}
		}
		return nil, domain.ErrInvalidOTP
	}

	// Delete immediately upon successful verification
	_ = s.client.Del(ctx, key).Err()

	return &outbound.OTPPendingData{
		AdminID:    payload.AdminID,
		OTP:        payload.OTP,
		RememberMe: payload.RememberMe,
		ClientIP:   payload.ClientIP,
	}, nil
}

func (s *RedisOTPStore) ClearOTP(ctx context.Context, adminID string) error {
	return s.client.Del(ctx, otpKey(adminID)).Err()
}
