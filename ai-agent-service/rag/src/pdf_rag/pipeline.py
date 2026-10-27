from __future__ import annotations

from pathlib import Path

from .config import EXTRACTED_DIR
from .documents import (
    extract_pdf_text,
    extracted_text_path,
    sha256_file,
    split_text,
    validate_pdf_filenames,
    write_extracted_text,
)
from .errors import RagError
from .providers import GeminiProvider
from .store import PostgresStore


def ingest_documents(
    filenames: list[str],
    store: PostgresStore,
    provider: GeminiProvider,
    documents_dir: Path,
    extracted_dir: Path = EXTRACTED_DIR,
    chunk_size: int = 1200,
    overlap: int = 200,
) -> list[str]:
    # Validate the entire explicit selection before connecting to external services
    # or extracting any page.
    pdf_paths = validate_pdf_filenames(filenames, documents_dir)
    results: list[str] = []
    for pdf_path in pdf_paths:
        digest = sha256_file(pdf_path)
        output_path = extracted_text_path(pdf_path, extracted_dir, digest)
        indexed_hash = store.document_hash(pdf_path.name, provider.embedding_model)

        if indexed_hash == digest and output_path.is_file():
            results.append(f"Skipped unchanged document: {pdf_path.name}")
            continue

        text = extract_pdf_text(pdf_path)
        write_extracted_text(output_path, text)
        if indexed_hash == digest:
            results.append(
                f"Restored extracted text; document is already indexed: {pdf_path.name}"
            )
            continue

        chunks = split_text(text, chunk_size=chunk_size, overlap=overlap)
        if not chunks:
            raise RagError(f"No non-empty text chunks were created for {pdf_path.name!r}.")
        embeddings = provider.embed_documents(chunks)
        store.replace_document(
            pdf_path.name, digest, provider.embedding_model, chunks, embeddings
        )
        results.append(
            f"Indexed {pdf_path.name}: {len(chunks)} chunk(s); "
            f"extracted text: {output_path}"
        )
    return results


def answer_instruction(
    instruction: str,
    store: PostgresStore,
    provider: GeminiProvider,
    top_k: int = 5,
) -> str:
    if not instruction.strip():
        raise RagError("Provide a non-empty natural-language instruction.")
    if top_k <= 0:
        raise RagError("The retrieval limit must be greater than zero.")
    query_embedding = provider.embed_query(instruction)
    context = store.search(query_embedding, top_k)
    if not context:
        raise RagError(
            "No document chunks are indexed yet. Run `pdf-rag ingest <filename.pdf>` "
            "with one or more explicit PDF filenames first."
        )
    return provider.answer(instruction, context)
