# PDF-to-RAG workflow

This standalone Python CLI ingests only PDF filenames you explicitly provide from
`ai-agent-service/knowledge/documents/`. It extracts selectable text directly,
uses Tesseract for pages without readable text, writes versioned text files under
`ai-agent-service/knowledge/extracted/`, and indexes chunks in PostgreSQL with
pgvector. The `ask` command retrieves relevant chunks and asks Google Gemini to
answer from those excerpts.

The Python project is separate from the Go modules. It does not configure or
start PostgreSQL, and it does not scan or select PDFs automatically.

## Requirements

- Python 3.10 or later.
- Tesseract OCR installed and available on `PATH` when a PDF has scanned pages.
- A PostgreSQL database where the role can create the RAG tables and use the
  `vector` extension (pgvector must be installed on the PostgreSQL server).
- A Google Gemini API key with access to the configured embedding and answer
  models.
- Connection details and API credentials supplied through environment variables.

The repository's Compose file does not create a PostgreSQL service. Use a
PostgreSQL instance you have configured with pgvector; no connection address or
credential is assumed here.

## Install

From the repository root, create a Python environment inside this directory and
install the project with its test dependency:

```sh
cd ai-agent-service/rag
python3 -m venv .venv
```

Activate it:

```sh
# Linux or macOS
source .venv/bin/activate

# Windows PowerShell
# .venv\Scripts\Activate.ps1
```

Then install:

```sh
python -m pip install -e ".[test]"
```

On Debian/Ubuntu, install the OCR executable separately with your system package
manager (for example, `sudo apt install tesseract-ocr`). For other operating
systems, install Tesseract using the platform's package manager or official
distribution and ensure its executable is on `PATH`.

## Configure environment variables

Set these in the shell that will run the CLI. Do not put real credentials in
source files or commit them:

| Variable | Required | Purpose |
| --- | --- | --- |
| `GEMINI_API_KEY` | Yes | Google Gemini API credential |
| `PSQL_HOST` | Yes | PostgreSQL server host |
| `PSQL_PORT` | Yes | PostgreSQL server port |
| `PSQL_USER` | Yes | PostgreSQL role |
| `PSQL_PASS` | Yes | PostgreSQL password |
| `PSQL_DBNAME` | Yes | Database name |
| `PSQL_SSLMODE` | Yes | PostgreSQL SSL mode for this connection |
| `GEMINI_EMBEDDING_MODEL` | No | Embedding model; default `gemini-embedding-001` |
| `GEMINI_ANSWER_MODEL` | No | Answer model; default `gemini-2.5-flash` |

`PSQL_*` names match those used by the existing Go services. The CLI intentionally
requires every connection value instead of supplying assumed credentials or
database defaults. For example, set each variable using the secret-management
method appropriate for your shell or deployment environment; do not copy
credentials into this README.

At first run, the CLI creates `rag_documents` and `rag_chunks` and requests
`CREATE EXTENSION IF NOT EXISTS vector`. If your database role cannot install the
extension, ask the database administrator to enable pgvector before retrying.
Embeddings use 768 dimensions; changing that model dimension requires a
coordinated schema migration and re-index.

## Ingest selected PDFs

Pass one or more exact basenames (not paths). Every requested name is validated
before any document is processed. If any name is invalid or missing, the entire
batch stops with an error.

```sh
pdf-rag ingest "Systemdesignnotes.pdf"
pdf-rag ingest "Systemdesignnotes.pdf" "FINAL VARIOUS POST.pdf"
```

Only the names passed to this command are considered; the workflow never lists
the directory to discover PDFs. Non-PDF names, path components, missing files,
and duplicate names are rejected.

Each extracted file is named
`<original-pdf-filename>.<first-16-characters-of-pdf-sha256>.txt`. This keeps
older extracted output intact when a PDF changes and makes each output
deterministic for its input version. Re-running ingestion skips an unchanged PDF
when its hash and embedding model are already indexed and its extracted output
exists. If output for an indexed PDF version is missing, it is regenerated
without embedding or re-indexing that document. Changing the embedding model
causes selected documents to be re-embedded on their next ingestion.

When a document changes, its chunks are replaced in a single database
transaction, preventing stale chunk vectors from remaining in search results.

## Ask a question

Provide a natural-language question or instruction as one quoted argument:

```sh
pdf-rag ask "Summarize the main topics in the system design notes."
pdf-rag ask --top-k 8 "Which eligibility criteria are listed?"
```

The answer uses only the retrieved document excerpts and reports when they do
not contain enough information. Source filenames are included in the retrieved
context provided to Gemini.

## Troubleshooting

- **Missing environment variable:** set the named `GEMINI_API_KEY` or `PSQL_*`
  variable in the current process environment and retry.
- **PostgreSQL connection error:** verify the values you supplied, network
  access, and PostgreSQL availability. The CLI does not create the database.
- **pgvector extension error:** have an administrator install/enable pgvector
  and grant the necessary permission.
- **OCR unavailable:** install Tesseract and ensure `tesseract` is on `PATH`.
  Text-based pages do not require OCR.
- **No indexed chunks:** run `pdf-rag ingest` with one or more exact PDF names.
- **Existing extracted output conflict:** versioned filenames are not
  overwritten. If a matching output path already contains different text, the
  workflow stops and preserves that file; move it aside manually before retrying.

## Tests

From `ai-agent-service/rag/` with the environment activated:

```sh
pytest
```

Tests use temporary files and fake provider/database seams; they do not require
Gemini credentials or a PostgreSQL server.
