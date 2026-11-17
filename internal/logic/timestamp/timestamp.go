package timestamp

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"

	"qingdao_backend/internal/client"
	"qingdao_backend/internal/types"
)

var ErrInvalidInput = errors.New("invalid timestamp input")

const MaxTSRBytes = 512 * 1024

type Service struct{ upstream client.TimestampProvider }

func New(upstream client.TimestampProvider) *Service { return &Service{upstream: upstream} }

func normalizeDigest(rid string) (string, error) {
	if len(rid) != 64 {
		return "", ErrInvalidInput
	}
	if _, err := hex.DecodeString(rid); err != nil {
		return "", ErrInvalidInput
	}
	return strings.ToLower(rid), nil
}

func (s *Service) Apply(ctx context.Context, algorithm string, req types.ApplyRequest) (*types.ApplyResult, error) {
	rid, err := normalizeDigest(req.RID)
	if err != nil {
		return nil, err
	}
	switch algorithm {
	case "SHA256":
		return s.upstream.RSA(ctx, rid)
	case "SM3":
		return s.upstream.SM(ctx, rid)
	default:
		return nil, ErrInvalidInput
	}
}

func (s *Service) Verify(ctx context.Context, req types.VerifyRequest) (*types.VerifyResult, error) {
	rid, err := normalizeDigest(req.RID)
	if err != nil {
		return nil, err
	}
	req.RID = rid
	if req.Algorithm != "SHA256" && req.Algorithm != "SM3" {
		return nil, ErrInvalidInput
	}
	if req.TSR == "" || len(req.TSR) > base64.StdEncoding.EncodedLen(MaxTSRBytes) {
		return nil, ErrInvalidInput
	}
	decoded, err := base64.StdEncoding.Strict().DecodeString(req.TSR)
	if err != nil || len(decoded) == 0 || len(decoded) > MaxTSRBytes {
		return nil, ErrInvalidInput
	}
	result, err := s.upstream.Verify(ctx, req)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("empty upstream verification result")
	}
	// Parsing a timestamp alone must never be treated as successful verification.
	if result.Valid && (!strings.EqualFold(result.RID, rid) || result.Algorithm != req.Algorithm) {
		result.Valid = false
		result.Reason = "DIGEST_MISMATCH"
	}
	return result, nil
}
