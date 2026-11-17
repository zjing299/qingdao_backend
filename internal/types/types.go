package types

// ApplyRequest contains a hex encoded digest computed by the frontend.
type ApplyRequest struct {
	RID string `json:"rid"`
}

type ApplyResult struct {
	RID       string `json:"rid"`
	Algorithm string `json:"algorithm"`
	SN        string `json:"sn"`
	TS        int64  `json:"ts"`  // Unix milliseconds, supplied by the upstream TSA.
	TSR       string `json:"tsr"` // Base64 encoded timestamp response.
}

type VerifyRequest struct {
	TSR       string `json:"tsr"`
	RID       string `json:"rid"`
	Algorithm string `json:"algorithm"` // SHA256 or SM3.
}

type VerifyResult struct {
	Valid     bool   `json:"valid"`
	RID       string `json:"rid"`
	Algorithm string `json:"algorithm"`
	SN        string `json:"sn,omitempty"`
	TS        int64  `json:"ts,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

type Response struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}
