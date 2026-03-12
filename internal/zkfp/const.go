//go:build windows

package zkfp

// DISPID constants from ZKFPEngX ActiveX (CZKFPEngX.h)
const (
	dispidEnrollCount       = 0x6
	dispidCancelEnroll      = 0x1
	dispidBeginCapture      = 0x4
	dispidCancelCapture     = 0x5
	dispidBeginEnroll       = 0xe
	dispidInitEngine        = 0x1a
	dispidEndEngine         = 0x1b
	dispidGetTemplateAsString = 0x27
	dispidGetTemplateAsStringEx = 0xdb
	dispidVerFingerFromStr  = 0x24
	dispidCreateFPCacheDBEx = 0xd4
	dispidFreeFPCacheDBEx   = 0xd5
	dispidAddRegTemplateStrToFPCacheDBEx = 0xda
	dispidIdentificationFromStrInFPCacheDB = 0x42
	dispidFPEngineVersion   = 0x33
)

// ZKFPEngX CLSID: CA69969C-2F27-41D3-954D-A48B941C3BA7
// ProgID is typically registered by ZKFinger SDK setup (e.g. from Biokey.ocx)
const defaultProgID = "ZKFPEngX.ZKFPEngX"
