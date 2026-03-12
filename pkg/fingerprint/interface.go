package fingerprint

import "time"

// CaptureResult holds template strings from a single capture.
type CaptureResult struct {
	Template9  string `json:"template9"`
	Template10 string `json:"template10,omitempty"`
}

// VerifyRequest for 1:1 verification.
type VerifyRequest struct {
	RegisteredTemplate  string `json:"registered_template"`
	VerificationTemplate string `json:"verification_template"`
	DoLearning          bool   `json:"do_learning,omitempty"`
}

// VerifyResult is the 1:1 match result.
type VerifyResult struct {
	Match bool  `json:"match"`
	Score int32 `json:"score,omitempty"`
}

// TemplateEntry for 1:N identify (one template set per user).
type TemplateEntry struct {
	ID         int32  `json:"id"`
	Template9  string `json:"template9"`
	Template10 string `json:"template10,omitempty"`
}

// IdentifyRequest for 1:N identification.
type IdentifyRequest struct {
	Templates          []TemplateEntry `json:"templates"`
	VerificationTemplate string        `json:"verification_template"`
}

// IdentifyResult is the 1:N match result.
type IdentifyResult struct {
	MatchedID   int32 `json:"matched_id"`   // -1 if no match
	Score       int32 `json:"score"`
	Processed   int32 `json:"processed"`
}

// Service interface for baca (capture), verifikasi (1:1), identifikasi (1:N).
// Implemented by the zkfp engine wrapper + API layer.
type Service interface {
	// Capture reads one fingerprint from hardware and returns template string(s). Blocks until capture or timeout.
	Capture(timeout time.Duration) (*CaptureResult, error)
	// Enroll runs multi-press enrollment and returns template string(s). Blocks until done or timeout.
	Enroll(presses int, timeout time.Duration) (*CaptureResult, error)
	// Verify performs 1:1 verification. Templates are provided by caller.
	Verify(regTemplate, verTemplate string, doLearning bool) (match bool, score int32, err error)
	// Identify performs 1:N identification. Templates list and verification template from caller.
	Identify(templates []TemplateEntry, verTemplate string) (matchedID int32, score int32, processed int32, err error)
}
