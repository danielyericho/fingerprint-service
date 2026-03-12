package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"fingerprint-service/internal/zkfp"
	"fingerprint-service/pkg/fingerprint"
)

func main() {
	cmd := flag.String("cmd", "", "command: capture, enroll, verify, identify")
	timeout := flag.Int("timeout", 30, "timeout seconds for capture/enroll")
	presses := flag.Int("presses", 3, "enroll presses")
	regTpl := flag.String("reg", "", "registered template (for verify)")
	verTpl := flag.String("ver", "", "verification template (for verify/identify)")
	templatesFile := flag.String("templates", "", "JSON file with templates array for identify (each: id, template9, template10)")
	flag.Parse()

	if *cmd == "" {
		fmt.Fprintf(os.Stderr, "usage: cli -cmd capture|enroll|verify|identify [options]\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	engine, err := zkfp.NewEngine("")
	if err != nil {
		log.Fatalf("engine: %v", err)
	}
	defer engine.Close()
	if _, err = engine.Init(); err != nil {
		log.Fatalf("init: %v", err)
	}
	_ = engine.SetFPEngineVersion("9")

	svc := &cliService{engine: engine}
	to := time.Duration(*timeout) * time.Second

	switch *cmd {
	case "capture":
		result, err := svc.Capture(to)
		if err != nil {
			log.Fatalf("capture: %v", err)
		}
		_ = json.NewEncoder(os.Stdout).Encode(result)
	case "enroll":
		result, err := svc.Enroll(*presses, to)
		if err != nil {
			log.Fatalf("enroll: %v", err)
		}
		_ = json.NewEncoder(os.Stdout).Encode(result)
	case "verify":
		if *regTpl == "" || *verTpl == "" {
			log.Fatal("verify requires -reg and -ver")
		}
		match, _, err := svc.Verify(*regTpl, *verTpl, false)
		if err != nil {
			log.Fatalf("verify: %v", err)
		}
		fmt.Printf("match=%v\n", match)
	case "identify":
		if *verTpl == "" || *templatesFile == "" {
			log.Fatal("identify requires -ver and -templates")
		}
		var list []fingerprint.TemplateEntry
		f, err := os.Open(*templatesFile)
		if err != nil {
			log.Fatalf("templates file: %v", err)
		}
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&list); err != nil {
			log.Fatalf("decode templates: %v", err)
		}
		id, score, proc, err := svc.Identify(list, *verTpl)
		if err != nil {
			log.Fatalf("identify: %v", err)
		}
		fmt.Printf("matched_id=%d score=%d processed=%d\n", id, score, proc)
	default:
		log.Fatalf("unknown cmd: %s", *cmd)
	}
}

type cliService struct {
	engine *zkfp.Engine
}

func (c *cliService) Capture(to time.Duration) (*fingerprint.CaptureResult, error) {
	t9, t10, err := c.engine.CaptureTemplate(to)
	if err != nil {
		return nil, err
	}
	return &fingerprint.CaptureResult{Template9: t9, Template10: t10}, nil
}

func (c *cliService) Enroll(presses int, to time.Duration) (*fingerprint.CaptureResult, error) {
	t9, t10, err := c.engine.EnrollTemplate(presses, to)
	if err != nil {
		return nil, err
	}
	return &fingerprint.CaptureResult{Template9: t9, Template10: t10}, nil
}

func (c *cliService) Verify(reg, ver string, doLearning bool) (bool, int32, error) {
	match, err := c.engine.VerFingerFromStr(reg, ver, doLearning)
	return match, 0, err
}

func (c *cliService) Identify(templates []fingerprint.TemplateEntry, ver string) (int32, int32, int32, error) {
	handle, err := c.engine.CreateFPCacheDBEx()
	if err != nil {
		return -1, 0, 0, err
	}
	defer c.engine.FreeFPCacheDBEx(handle)
	for _, t := range templates {
		t10 := t.Template10
		if t10 == "" {
			t10 = t.Template9
		}
		_, _ = c.engine.AddRegTemplateStrToFPCacheDBEx(handle, t.ID, t.Template9, t10)
	}
	return c.engine.IdentificationFromStrInFPCacheDB(handle, ver)
}
