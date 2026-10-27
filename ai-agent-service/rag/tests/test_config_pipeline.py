from pathlib import Path
from types import SimpleNamespace

import pytest

from pdf_rag.config import DOCUMENTS_DIR, Settings, VECTOR_DIMENSIONS
from pdf_rag.cli import main
from pdf_rag.errors import RagError
from pdf_rag.pipeline import answer_instruction, ingest_documents
from pdf_rag.providers import GeminiProvider


def _environment(monkeypatch: pytest.MonkeyPatch) -> None:
    values = {
        "GEMINI_API_KEY": "test-key",
        "PSQL_HOST": "db-host",
        "PSQL_PORT": "5432",
        "PSQL_USER": "db-user",
        "PSQL_PASS": "db-password",
        "PSQL_DBNAME": "db-name",
        "PSQL_SSLMODE": "verify-full",
    }
    for name, value in values.items():
        monkeypatch.setenv(name, value)


def test_settings_require_explicit_secrets_and_database_configuration(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    for name in (
        "GEMINI_API_KEY",
        "PSQL_HOST",
        "PSQL_PORT",
        "PSQL_USER",
        "PSQL_PASS",
        "PSQL_DBNAME",
        "PSQL_SSLMODE",
    ):
        monkeypatch.delenv(name, raising=False)

    with pytest.raises(RagError, match="GEMINI_API_KEY"):
        Settings.from_env()


def test_document_directory_points_inside_ai_agent_service() -> None:
    assert DOCUMENTS_DIR.name == "documents"
    assert DOCUMENTS_DIR.parent.name == "knowledge"
    assert DOCUMENTS_DIR.parent.parent.name == "ai-agent-service"


def test_cli_validates_explicit_filenames_before_environment(
    capsys: pytest.CaptureFixture[str],
) -> None:
    assert main(["ingest", "missing.pdf"]) == 2
    assert "missing.pdf" in capsys.readouterr().err


def test_settings_reuse_psql_variables_and_validate_port(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    _environment(monkeypatch)
    settings = Settings.from_env()

    assert settings.postgres_connection_kwargs() == {
        "host": "db-host",
        "port": 5432,
        "user": "db-user",
        "password": "db-password",
        "dbname": "db-name",
        "sslmode": "verify-full",
    }
    monkeypatch.setenv("PSQL_PORT", "not-a-port")
    with pytest.raises(RagError, match="PSQL_PORT"):
        Settings.from_env()


def test_ingestion_skips_unchanged_document_before_extraction(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    pdf = tmp_path / "chosen.pdf"
    pdf.write_bytes(b"PDF bytes")
    digest = __import__("hashlib").sha256(b"PDF bytes").hexdigest()
    store = SimpleNamespace(document_hash=lambda _name, _model: digest)
    provider = SimpleNamespace(
        embedding_model="test-embedding-model",
        embed_documents=lambda _chunks: pytest.fail("unchanged documents are not embedded"),
    )
    output = tmp_path / "extracted" / f"chosen.pdf.{digest[:16]}.txt"
    output.parent.mkdir()
    output.write_text("already extracted", encoding="utf-8")
    monkeypatch.setattr(
        "pdf_rag.pipeline.extract_pdf_text",
        lambda _path: pytest.fail("unchanged documents are not extracted"),
    )

    result = ingest_documents(
        ["chosen.pdf"], store, provider, documents_dir=tmp_path, extracted_dir=output.parent
    )

    assert result == ["Skipped unchanged document: chosen.pdf"]


def test_invalid_batch_is_rejected_before_hash_or_extraction(
    tmp_path: Path,
) -> None:
    (tmp_path / "selected.pdf").write_bytes(b"PDF bytes")
    store = SimpleNamespace(
        document_hash=lambda _name: pytest.fail("invalid batches must fail first")
    )
    provider = SimpleNamespace(
        embed_documents=lambda _chunks: pytest.fail("invalid batches must fail first")
    )

    with pytest.raises(RagError, match="absent.pdf"):
        ingest_documents(
            ["selected.pdf", "absent.pdf"],
            store,
            provider,
            documents_dir=tmp_path,
            extracted_dir=tmp_path / "extracted",
        )


def test_changed_document_is_extracted_embedded_and_stored(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    (tmp_path / "selected.pdf").write_bytes(b"new PDF bytes")
    observed: dict[str, object] = {}

    class Store:
        def document_hash(self, _name: str, _model: str) -> None:
            return None

        def replace_document(self, *args: object) -> None:
            observed["stored"] = args

    class Provider:
        embedding_model = "test-embedding-model"

        def embed_documents(self, chunks: list[str]) -> list[list[float]]:
            observed["chunks"] = chunks
            return [[0.0] * VECTOR_DIMENSIONS for _ in chunks]

    monkeypatch.setattr(
        "pdf_rag.pipeline.extract_pdf_text", lambda _path: "Selected PDF content."
    )
    extracted = tmp_path / "extracted"
    messages = ingest_documents(
        ["selected.pdf"],
        Store(),
        Provider(),
        documents_dir=tmp_path,
        extracted_dir=extracted,
    )

    assert messages[0].startswith("Indexed selected.pdf: 1 chunk(s)")
    assert observed["chunks"] == ["Selected PDF content."]
    assert observed["stored"][0] == "selected.pdf"
    assert (extracted / f"selected.pdf.{observed['stored'][1][:16]}.txt").is_file()


def test_question_answers_use_retrieved_chunks_only() -> None:
    observed: dict[str, object] = {}
    store = SimpleNamespace(search=lambda _vector, limit: [("source.pdf", "context")])

    class Provider:
        embedding_model = "test-embedding-model"

        def embed_query(self, text: str) -> list[float]:
            observed["query"] = text
            return [0.0] * VECTOR_DIMENSIONS

        def answer(self, instruction: str, context: list[tuple[str, str]]) -> str:
            observed["context"] = context
            return f"Answer: {instruction}"

    assert answer_instruction("Summarize this", store, Provider()) == (
        "Answer: Summarize this"
    )
    assert observed["context"] == [("source.pdf", "context")]


def test_answer_reports_empty_index_and_invalid_input() -> None:
    provider = SimpleNamespace(embed_query=lambda _text: [0.0] * VECTOR_DIMENSIONS)
    store = SimpleNamespace(search=lambda _vector, _limit: [])

    with pytest.raises(RagError, match="No document chunks are indexed"):
        answer_instruction("Question", store, provider)
    with pytest.raises(RagError, match="non-empty"):
        answer_instruction(" ", store, provider)


def test_gemini_provider_configures_embeddings_and_answer() -> None:
    class FakeModels:
        def embed_content(self, **kwargs: object) -> object:
            assert kwargs["model"] == "embedding-model"
            assert kwargs["config"].output_dimensionality == VECTOR_DIMENSIONS
            contents = kwargs["contents"]
            return SimpleNamespace(
                embeddings=[
                    SimpleNamespace(values=[0.1] * VECTOR_DIMENSIONS)
                    for _ in contents
                ]
            )

        def generate_content(self, **kwargs: object) -> object:
            assert kwargs["model"] == "answer-model"
            assert "Source: source.pdf" in kwargs["contents"]
            return SimpleNamespace(text="Grounded answer")

    class FakeClient:
        models = FakeModels()

    provider = object.__new__(GeminiProvider)
    provider._types = SimpleNamespace(
        EmbedContentConfig=lambda **kwargs: SimpleNamespace(**kwargs),
        GenerateContentConfig=lambda **kwargs: SimpleNamespace(**kwargs),
    )
    provider._client = FakeClient()
    provider._embedding_model = "embedding-model"
    provider._answer_model = "answer-model"

    assert len(provider.embed_documents(["document"])[0]) == VECTOR_DIMENSIONS
    assert len(provider.embed_query("question")) == VECTOR_DIMENSIONS
    assert provider.answer("question", [("source.pdf", "evidence")]) == "Grounded answer"
