package api

import (
	"fingerprint-service/internal/zkfp"
	"fingerprint-service/pkg/fingerprint"
	"time"
)

type serviceImpl struct {
	engine *zkfp.Engine
}

func (s *serviceImpl) Capture(timeout time.Duration) (*fingerprint.CaptureResult, error) {
	t9, t10, err := s.engine.CaptureTemplate(timeout)
	if err != nil {
		return nil, err
	}
	return &fingerprint.CaptureResult{Template9: t9, Template10: t10}, nil
}

func (s *serviceImpl) Enroll(presses int, timeout time.Duration) (*fingerprint.CaptureResult, error) {
	t9, t10, err := s.engine.EnrollTemplate(presses, timeout)
	if err != nil {
		return nil, err
	}
	return &fingerprint.CaptureResult{Template9: t9, Template10: t10}, nil
}

func (s *serviceImpl) Verify(regTemplate, verTemplate string, doLearning bool) (match bool, score int32, err error) {
	match, err = s.engine.VerFingerFromStr(regTemplate, verTemplate, doLearning)
	if err != nil {
		return false, 0, err
	}
	return match, 0, nil
}

func (s *serviceImpl) Identify(templates []fingerprint.TemplateEntry, verTemplate string) (matchedID int32, score int32, processed int32, err error) {
	handle, err := s.engine.CreateFPCacheDBEx()
	if err != nil {
		return -1, 0, 0, err
	}
	defer s.engine.FreeFPCacheDBEx(handle)
	for _, t := range templates {
		t10 := t.Template10
		if t10 == "" {
			t10 = t.Template9
		}
		_, err = s.engine.AddRegTemplateStrToFPCacheDBEx(handle, t.ID, t.Template9, t10)
		if err != nil {
			return -1, 0, 0, err
		}
	}
	matchedID, score, processed, err = s.engine.IdentificationFromStrInFPCacheDB(handle, verTemplate)
	if err != nil {
		return -1, 0, 0, err
	}
	return matchedID, score, processed, nil
}
