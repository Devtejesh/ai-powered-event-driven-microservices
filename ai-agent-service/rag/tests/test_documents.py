from pathlib import Path

import pytest

from pdf_rag.documents import (
    extracted_text_path,
    extract_pdf_text,
    sha256_file,
    split_text,
    validate_pdf_filenames,
    write_extracted_text,
)
from pdf_rag.errors import RagError


def test_selection_accepts_only_existing_pdf_basenames(tmp_path: Path) -> None:
    docs = tmp_path / "documents"
    docs.mkdir()
    expected = docs / "named.pdf"
    expected.write_bytes(b"%PDF-test")

    assert validate_pdf_filenames(["named.pdf"], docs) == [expected]


@pytest.mark.parametrize(
    "filename", ["../named.pdf", "/tmp/named.pdf", "folder\\named.pdf", "notes.txt"]
)
def test_selection_rejects_paths_and_non_pdf_names(
    tmp_path: Path, filename: str
) -> None:
    with pytest.raises(RagError, match="No files were processed"):
        validate_pdf_filenames([filename], tmp_path)


def test_selection_validates_entire_batch_before_returning(tmp_path: Path) -> None:
    (tmp_path / "present.pdf").write_bytes(b"%PDF-test")

    with pytest.raises(RagError, match="absent.pdf"):
        validate_pdf_filenames(["present.pdf", "absent.pdf"], tmp_path)


def test_selection_rejects_duplicate_names(tmp_path: Path) -> None:
    (tmp_path / "present.pdf").write_bytes(b"%PDF-test")

    with pytest.raises(RagError, match="more than once"):
        validate_pdf_filenames(["present.pdf", "present.pdf"], tmp_path)


def test_extract_pdf_uses_ocr_only_when_page_has_no_text(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    pypdf = pytest.importorskip("pypdf")
    writer = pypdf.PdfWriter()
    writer.add_blank_page(width=200, height=200)
    pdf = tmp_path / "scanned.pdf"
    with pdf.open("wb") as target:
        writer.write(target)
    called: list[int] = []

    def fake_ocr(page_number: int) -> str:
        called.append(page_number)
        return "OCR text from scanned page"

    assert extract_pdf_text(pdf, ocr_page=fake_ocr) == "OCR text from scanned page\n"
    assert called == [0]


def test_extract_pdf_keeps_text_without_ocr(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    pypdf = pytest.importorskip("pypdf")

    writer = pypdf.PdfWriter()
    writer.add_blank_page(width=200, height=200)
    pdf = tmp_path / "text.pdf"
    with pdf.open("wb") as target:
        writer.write(target)
    monkeypatch.setattr(
        "pypdf.PdfReader",
        lambda _path: type(
            "Reader",
            (),
            {
                "pages": [
                    type(
                        "Page",
                        (),
                        {"extract_text": lambda _self: "Selectable text"},
                    )()
                ]
            },
        )(),
    )
    assert extract_pdf_text(pdf) == "Selectable text\n"


def test_chunks_are_bounded_and_overlap() -> None:
    text = "one two three four five six"
    chunks = split_text(text, chunk_size=12, overlap=5)

    assert chunks == ["one two", "two three", "three four", "four five", "five six"]


@pytest.mark.parametrize(("size", "overlap"), [(0, 0), (10, -1), (10, 10)])
def test_chunk_settings_are_validated(size: int, overlap: int) -> None:
    with pytest.raises(RagError):
        split_text("some text", chunk_size=size, overlap=overlap)


def test_output_path_is_content_versioned_and_writer_preserves_existing_file(
    tmp_path: Path,
) -> None:
    pdf = tmp_path / "selected.pdf"
    pdf.write_bytes(b"first")
    digest = sha256_file(pdf)
    output = extracted_text_path(pdf, tmp_path / "extracted", digest)
    output.parent.mkdir()
    output.write_text("existing output\n", encoding="utf-8")

    with pytest.raises(RagError, match="Refusing to overwrite"):
        write_extracted_text(output, "new output\n")
    assert output.read_text(encoding="utf-8") == "existing output\n"

    write_extracted_text(output, "existing output\n")
    assert output.read_text(encoding="utf-8") == "existing output\n"
