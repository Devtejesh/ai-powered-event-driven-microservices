from __future__ import annotations

import hashlib
import re
import shutil
from pathlib import Path
from typing import Callable

from .errors import RagError

_SAFE_PDF_NAME = re.compile(r"^[^/\\\x00]+\.pdf$", re.IGNORECASE)


def validate_pdf_filenames(
    filenames: list[str], documents_dir: Path
) -> list[Path]:
    if not filenames:
        raise RagError("Provide at least one PDF filename to ingest.")

    errors: list[str] = []
    paths: list[Path] = []
    seen: set[str] = set()

    for filename in filenames:
        if not _SAFE_PDF_NAME.fullmatch(filename) or filename in {".pdf", "..pdf"}:
            errors.append(
                f"{filename!r} is not a safe PDF basename. "
                "Pass a filename only, without directories or paths."
            )
            continue
        if filename in seen:
            errors.append(f"{filename!r} was listed more than once.")
            continue
        seen.add(filename)
        path = documents_dir / filename
        if not path.is_file():
            errors.append(
                f"PDF {filename!r} was not found in {documents_dir}."
            )
            continue
        if path.resolve().parent != documents_dir.resolve():
            errors.append(f"PDF {filename!r} must be directly inside {documents_dir}.")
            continue
        paths.append(path)

    if errors:
        raise RagError("No files were processed:\n- " + "\n- ".join(errors))
    return paths


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    try:
        with path.open("rb") as source:
            for block in iter(lambda: source.read(1024 * 1024), b""):
                digest.update(block)
    except OSError as exc:
        raise RagError(f"Could not read PDF {path.name!r}: {exc}") from exc
    return digest.hexdigest()


def extracted_text_path(pdf_path: Path, extracted_dir: Path, digest: str) -> Path:
    return extracted_dir / f"{pdf_path.name}.{digest[:16]}.txt"


def extract_pdf_text(
    pdf_path: Path,
    ocr_page: Callable[[int], str] | None = None,
) -> str:
    try:
        from pypdf import PdfReader
    except ImportError as exc:
        raise RagError("PDF extraction dependency missing. Install rag/ dependencies.") from exc

    try:
        reader = PdfReader(str(pdf_path))
        if not reader.pages:
            raise RagError(f"PDF {pdf_path.name!r} contains no pages.")
        page_texts: list[str] = []
        for page_number, page in enumerate(reader.pages):
            text = page.extract_text() or ""
            if not _has_usable_text(text):
                text = ocr_page(page_number) if ocr_page else _ocr_pdf_page(
                    pdf_path, page_number
                )
            page_texts.append(text.strip())
    except RagError:
        raise
    except Exception as exc:
        raise RagError(f"Could not read PDF {pdf_path.name!r}: {exc}") from exc

    extracted = "\n\n".join(text for text in page_texts if text)
    if not extracted:
        raise RagError(
            f"No readable text was extracted from {pdf_path.name!r}. "
            "Check that the PDF is valid and that Tesseract OCR is installed."
        )
    return extracted + "\n"


def _has_usable_text(text: str) -> bool:
    return any(character.isalnum() for character in text)


def _ocr_pdf_page(pdf_path: Path, page_number: int) -> str:
    if not shutil.which("tesseract"):
        raise RagError(
            f"Page {page_number + 1} of {pdf_path.name!r} needs OCR, but the "
            "Tesseract executable was not found. Install Tesseract and ensure "
            "it is on PATH."
        )
    try:
        import fitz
        import pytesseract
    except ImportError as exc:
        raise RagError(
            "OCR dependencies missing. Install the rag/ Python dependencies."
        ) from exc

    try:
        with fitz.open(pdf_path) as document:
            page = document.load_page(page_number)
            image = page.get_pixmap(matrix=fitz.Matrix(2, 2), alpha=False)
            from PIL import Image

            return pytesseract.image_to_string(Image.frombytes(
                "RGB", [image.width, image.height], image.samples
            ))
    except Exception as exc:
        raise RagError(
            f"OCR failed on page {page_number + 1} of {pdf_path.name!r}: {exc}"
        ) from exc


def write_extracted_text(path: Path, text: str) -> None:
    try:
        path.parent.mkdir(parents=True, exist_ok=True)
        try:
            with path.open("x", encoding="utf-8") as output:
                output.write(text)
            return
        except FileExistsError:
            existing_text = path.read_text(encoding="utf-8")
            if existing_text != text:
                raise RagError(
                    f"Refusing to overwrite existing extracted text at {path}. "
                    "Move it aside before reprocessing this PDF."
                )
    except RagError:
        raise
    except OSError as exc:
        raise RagError(f"Could not write extracted text to {path}: {exc}") from exc


def split_text(text: str, chunk_size: int = 1200, overlap: int = 200) -> list[str]:
    if chunk_size <= 0:
        raise RagError("Chunk size must be greater than zero.")
    if overlap < 0 or overlap >= chunk_size:
        raise RagError("Chunk overlap must be non-negative and smaller than chunk size.")

    normalized = re.sub(r"\s+", " ", text).strip()
    if not normalized:
        return []

    words = normalized.split(" ")
    chunks: list[str] = []
    start = 0
    while start < len(words):
        end = start
        length = 0
        while end < len(words):
            candidate_length = length + (1 if end > start else 0) + len(words[end])
            if candidate_length > chunk_size and end > start:
                break
            length = candidate_length
            end += 1

        chunks.append(" ".join(words[start:end]))
        if end >= len(words):
            break

        next_start = end
        overlap_length = 0
        while next_start > start + 1:
            word_length = len(words[next_start - 1])
            candidate_length = overlap_length + (1 if overlap_length else 0) + word_length
            if candidate_length > overlap:
                break
            overlap_length = candidate_length
            next_start -= 1
        start = next_start
    return chunks
