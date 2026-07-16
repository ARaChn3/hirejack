# hirejack

ATS security auditing tool. Tests data integrity, injection vectors, and DoS resilience through poisoned applicant data.

## Audit Modes

### 1. Data Integrity Audit
Submit known-good data → retrieve via ATS → diff for corruption.

**Resume sections tested:**
- Contact: name, email, phone, address, LinkedIn
- Personal: DOB, nationality, pronouns
- Experience: job titles, companies, dates, descriptions
- Education: degrees, institutions, dates, GPA
- Skills: technical, languages, certifications
- Projects: titles, descriptions, URLs
- References: names, contacts

**Corruption patterns detected:**
- Unicode mangling (diacritics, CJK, RTL)
- Field truncation at common limits
- Date format conversion errors
- Special character stripping (`C++` → `C`)
- Phone/URL reformatting corruption
- Whitespace normalization issues

### 2. Injection Testing
Inject payloads to test ATS input validation.

**Vectors:**
- SQLi in text fields
- XXE in embedded XML (DOCX structure)
- DoS via recursive structures
- Unicode bombs

### 3. ML Training Data Priority
Fields with highest ML signal (highest-value audit targets):
- Skills/keywords (screening algorithms)
- Job titles (matching models)  
- Company names (prestige signals)
- Education (degree requirements)
- Experience duration (filters)

## Stack

- Go backend (single binary, embedded frontend)
- React + Vite frontend
- pdfcpu for PDF manipulation

## Build

```bash
cd frontend && pnpm build
cd .. && go build -o hirejack .
./hirejack  # serves on :8080
```

## License

MIT
