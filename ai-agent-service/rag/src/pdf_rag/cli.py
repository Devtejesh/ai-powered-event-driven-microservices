from __future__ import annotations

import argparse
import sys

from .config import DOCUMENTS_DIR, Settings
from .errors import RagError
from .pipeline import answer_instruction, ingest_documents
from .providers import GeminiProvider
from .store import PostgresStore
from .documents import validate_pdf_filenames


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="pdf-rag",
        description="Ingest explicitly named PDFs or answer from indexed documents.",
    )
    commands = parser.add_subparsers(dest="command", required=True)

    ingest = commands.add_parser(
        "ingest", help="Extract and index only the PDF filenames you provide."
    )
    ingest.add_argument(
        "filenames",
        nargs="+",
        metavar="PDF_FILENAME",
        help="PDF basename(s) located directly in ai-agent-service/knowledge/documents/",
    )
    ask = commands.add_parser(
        "ask", help="Answer a natural-language instruction from retrieved chunks."
    )
    ask.add_argument("instruction", help="Question or instruction to answer.")
    ask.add_argument("--top-k", type=int, default=5, help="Retrieved chunks (default: 5).")
    return parser


def main(argv: list[str] | None = None) -> int:
    args = build_parser().parse_args(argv)
    try:
        if args.command == "ingest":
            validate_pdf_filenames(args.filenames, DOCUMENTS_DIR)
        settings = Settings.from_env()
        store = PostgresStore.connect(settings)
        try:
            store.ensure_schema()
            provider = GeminiProvider(settings)
            if args.command == "ingest":
                for message in ingest_documents(
                    args.filenames,
                    store,
                    provider,
                    documents_dir=DOCUMENTS_DIR,
                ):
                    print(message)
            else:
                print(answer_instruction(args.instruction, store, provider, args.top_k))
        finally:
            store.close()
        return 0
    except RagError as exc:
        print(f"Error: {exc}", file=sys.stderr)
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
