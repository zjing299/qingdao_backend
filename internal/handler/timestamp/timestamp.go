package timestamp

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"qingdao_backend/internal/client"
	logic "qingdao_backend/internal/logic/timestamp"
	"qingdao_backend/internal/types"
)

const maxBodyBytes = 768 * 1024

func Apply(service *logic.Service, algorithm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ApplyRequest
		if !decode(w, r, &req) {
			return
		}
		result, err := service.Apply(r.Context(), algorithm, req)
		respond(w, result, err)
	}
}

func Verify(service *logic.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.VerifyRequest
		if !decode(w, r, &req) {
			return
		}
		result, err := service.Verify(r.Context(), req)
		respond(w, result, err)
	}
}

func decode(w http.ResponseWriter, r *http.Request, target any) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		write(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "请使用 application/json 提交摘要或凭证", nil)
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(target)
	if err == nil {
		var extra any
		if nextErr := decoder.Decode(&extra); nextErr != io.EOF {
			if nextErr != nil {
				err = nextErr
			} else {
				err = logic.ErrInvalidInput
			}
		}
	}
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			write(w, http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE", "请求体过大", nil)
		} else {
			write(w, http.StatusBadRequest, "INVALID_REQUEST", "请求必须为单个 JSON 对象，且只包含接口支持的字段", nil)
		}
		return false
	}
	return true
}

func respond(w http.ResponseWriter, result any, err error) {
	switch {
	case err == nil:
		write(w, http.StatusOK, "OK", "成功", result)
	case errors.Is(err, logic.ErrInvalidInput):
		write(w, http.StatusBadRequest, "INVALID_REQUEST", "摘要须为 64 位十六进制；验证须提供 Base64 凭证及 SHA256 或 SM3 算法", nil)
	case errors.Is(err, client.ErrNotConfigured):
		write(w, http.StatusServiceUnavailable, "TIMESTAMP_NOT_CONFIGURED", "上级时间戳服务尚未配置", nil)
	case errors.Is(err, client.ErrProtocolPending):
		write(w, http.StatusServiceUnavailable, "TIMESTAMP_INTEGRATION_PENDING", "上级时间戳接口协议尚未对接", nil)
	default:
		write(w, http.StatusBadGateway, "TIMESTAMP_UPSTREAM_ERROR", "上级时间戳服务调用失败", nil)
	}
}

func write(w http.ResponseWriter, status int, code, message string, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(types.Response{Code: code, Message: message, Data: data})
}
