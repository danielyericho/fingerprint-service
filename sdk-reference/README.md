# Referensi SDK ZKFinger (ZKTeco)

Folder ini berisi salinan **referensi** dari ZKFinger SDK untuk keperluan development:

- **Readme.txt** — Informasi SDK, changelog, sistem operasi yang didukung
- **Licence.txt** — Lisensi
- **Samples/VC/** — Kode sample C++ (ActiveX):
  - **CZKFPEngX.h** — Deklarasi API ActiveX (DISPID, method)
  - **CZKFPEngX.cpp** — Implementasi wrapper MFC
  - **DemoDlg.cpp** / **DemoDlg.h** — Contoh penggunaan: InitEngine, BeginEnroll, GetTemplateAsString, VerFingerFromStr, IdentificationInFPCacheDB, dll.

**Catatan:** Driver dan komponen COM (mis. Biokey.ocx) **tidak** disertakan di repo. Install SDK/driver dari paket asli (folder Fingerprint atau setup.exe) di mesin Windows yang akan menjalankan fingerprint-service.
