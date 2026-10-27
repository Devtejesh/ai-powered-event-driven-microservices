from __future__ import annotations

from typing import Any

import pytest
from pgvector import Vector

from pdf_rag.config import VECTOR_DIMENSIONS
from pdf_rag.errors import RagError
from pdf_rag.store import PostgresStore


class FakeCursor:
    def __init__(self, connection: "FakeConnection") -> None:
        self.connection = connection

    def __enter__(self) -> "FakeCursor":
        return self

    def __exit__(self, *_args: object) -> None:
        return None

    def execute(self, query: str, params: Any = None) -> None:
        if self.connection.fail_on and self.connection.fail_on in query:
            raise RuntimeError("simulated database failure")
        self.connection.statements.append((query, params))

    def executemany(self, query: str, params: Any) -> None:
        if self.connection.fail_on and self.connection.fail_on in query:
            raise RuntimeError("simulated database failure")
        self.connection.statements.append((query, list(params)))

    def fetchone(self) -> tuple[str] | None:
        return self.connection.fetchone_result

    def fetchall(self) -> list[tuple[str, str]]:
        return self.connection.fetchall_result


class FakeConnection:
    def __init__(self) -> None:
        self.statements: list[tuple[str, Any]] = []
        self.commits = 0
        self.rollbacks = 0
        self.closed = False
        self.fail_on: str | None = None
        self.fetchone_result: tuple[str] | None = None
        self.fetchall_result: list[tuple[str, str]] = []

    def cursor(self) -> FakeCursor:
        return FakeCursor(self)

    def commit(self) -> None:
        self.commits += 1

    def rollback(self) -> None:
        self.rollbacks += 1

    def close(self) -> None:
        self.closed = True


def test_schema_creates_pgvector_tables_and_commits() -> None:
    connection = FakeConnection()
    store = PostgresStore(connection)

    store.ensure_schema()

    statements = "\n".join(query for query, _params in connection.statements)
    assert "CREATE EXTENSION IF NOT EXISTS vector" in statements
    assert "rag_documents" in statements
    assert "embedding_model" in statements
    assert f"vector({VECTOR_DIMENSIONS})" in statements
    assert "vector_cosine_ops" in statements
    assert connection.commits == 1


def test_document_hash_reads_hash_by_explicit_filename() -> None:
    connection = FakeConnection()
    connection.fetchone_result = ("digest",)
    store = PostgresStore(connection)

    assert store.document_hash("selected.pdf", "embedding-model") == "digest"
    assert connection.statements[0][1] == ("selected.pdf", "embedding-model")


def test_replacement_is_committed_atomically_and_replaces_chunks() -> None:
    connection = FakeConnection()
    store = PostgresStore(connection)
    vectors = [[0.0] * VECTOR_DIMENSIONS]

    store.replace_document(
        "selected.pdf", "digest", "embedding-model", ["chunk"], vectors
    )

    statements = [query for query, _params in connection.statements]
    assert "INSERT INTO rag_documents" in statements[0]
    assert "DELETE FROM rag_chunks" in statements[1]
    assert "INSERT INTO rag_chunks" in statements[2]
    inserted_vector = connection.statements[2][1][0][3]
    assert isinstance(inserted_vector, Vector)
    assert inserted_vector.to_list() == vectors[0]
    assert connection.commits == 1
    assert connection.rollbacks == 0


def test_failed_replacement_rolls_back_and_reports_error() -> None:
    connection = FakeConnection()
    connection.fail_on = "INSERT INTO rag_chunks"
    store = PostgresStore(connection)

    with pytest.raises(RagError, match="Could not store document"):
        store.replace_document(
            "selected.pdf",
            "digest",
            "embedding-model",
            ["chunk"],
            [[0.0] * VECTOR_DIMENSIONS],
        )

    assert connection.commits == 0
    assert connection.rollbacks == 1


def test_search_uses_cosine_distance_and_returns_sources() -> None:
    connection = FakeConnection()
    connection.fetchall_result = [("selected.pdf", "relevant text")]
    store = PostgresStore(connection)

    assert store.search([0.0] * VECTOR_DIMENSIONS, 4) == [
        ("selected.pdf", "relevant text")
    ]
    query, params = connection.statements[0]
    assert "embedding <=> %s" in query
    assert isinstance(params[0], Vector)
    assert params[0].to_list() == [0.0] * VECTOR_DIMENSIONS
    assert params[1] == 4
