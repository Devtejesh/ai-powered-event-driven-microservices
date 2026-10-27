from __future__ import annotations

from collections.abc import Sequence
from typing import Any

from pgvector import Vector

from .config import Settings, VECTOR_DIMENSIONS
from .errors import RagError


class PostgresStore:
    def __init__(self, connection: Any) -> None:
        self.connection = connection

    @classmethod
    def connect(cls, settings: Settings) -> "PostgresStore":
        try:
            import psycopg
            from pgvector.psycopg import register_vector
        except ImportError as exc:
            raise RagError(
                "PostgreSQL dependencies missing. Install the rag/ Python dependencies."
            ) from exc

        try:
            connection = psycopg.connect(**settings.postgres_connection_kwargs())
            register_vector(connection)
            return cls(connection)
        except Exception as exc:
            raise RagError(
                "Could not connect to PostgreSQL using PSQL_* environment "
                f"settings: {exc}"
            ) from exc

    def close(self) -> None:
        self.connection.close()

    def ensure_schema(self) -> None:
        try:
            with self.connection.cursor() as cursor:
                cursor.execute("CREATE EXTENSION IF NOT EXISTS vector")
                cursor.execute(
                    """
                    CREATE TABLE IF NOT EXISTS rag_documents (
                        source_filename TEXT PRIMARY KEY,
                        sha256 TEXT NOT NULL,
                        embedding_model TEXT NOT NULL,
                        indexed_at TIMESTAMPTZ NOT NULL DEFAULT now()
                    )
                    """
                )
                cursor.execute(
                    """
                    ALTER TABLE rag_documents
                    ADD COLUMN IF NOT EXISTS embedding_model TEXT
                    """
                )
                cursor.execute(
                    f"""
                    CREATE TABLE IF NOT EXISTS rag_chunks (
                        source_filename TEXT NOT NULL
                            REFERENCES rag_documents(source_filename)
                            ON DELETE CASCADE,
                        chunk_index INTEGER NOT NULL,
                        content TEXT NOT NULL,
                        embedding vector({VECTOR_DIMENSIONS}) NOT NULL,
                        PRIMARY KEY (source_filename, chunk_index)
                    )
                    """
                )
                cursor.execute(
                    """
                    CREATE INDEX IF NOT EXISTS rag_chunks_embedding_hnsw
                    ON rag_chunks USING hnsw (embedding vector_cosine_ops)
                    """
                )
            self.connection.commit()
        except Exception as exc:
            self.connection.rollback()
            raise RagError(
                "Could not initialize the RAG tables. Confirm the PostgreSQL "
                "database is reachable and the current role may install/use "
                f"the pgvector extension: {exc}"
            ) from exc

    def document_hash(self, filename: str, embedding_model: str) -> str | None:
        try:
            with self.connection.cursor() as cursor:
                cursor.execute(
                    """
                    SELECT sha256 FROM rag_documents
                    WHERE source_filename = %s AND embedding_model = %s
                    """,
                    (filename, embedding_model),
                )
                row = cursor.fetchone()
            return row[0] if row else None
        except Exception as exc:
            raise RagError(f"Could not check indexed document {filename!r}: {exc}") from exc

    def replace_document(
        self,
        filename: str,
        digest: str,
        embedding_model: str,
        chunks: Sequence[str],
        embeddings: Sequence[Sequence[float]],
    ) -> None:
        if len(chunks) != len(embeddings):
            raise RagError("Chunk and embedding counts do not match.")
        try:
            with self.connection.cursor() as cursor:
                cursor.execute(
                    """
                    INSERT INTO rag_documents
                        (source_filename, sha256, embedding_model, indexed_at)
                    VALUES (%s, %s, %s, now())
                    ON CONFLICT (source_filename) DO UPDATE
                    SET sha256 = EXCLUDED.sha256,
                        embedding_model = EXCLUDED.embedding_model,
                        indexed_at = now()
                    """,
                    (filename, digest, embedding_model),
                )
                cursor.execute(
                    "DELETE FROM rag_chunks WHERE source_filename = %s",
                    (filename,),
                )
                cursor.executemany(
                    """
                    INSERT INTO rag_chunks
                        (source_filename, chunk_index, content, embedding)
                    VALUES (%s, %s, %s, %s)
                    """,
                    [
                        (filename, index, chunk, Vector(embedding))
                        for index, (chunk, embedding) in enumerate(
                            zip(chunks, embeddings, strict=True)
                        )
                    ],
                )
            self.connection.commit()
        except Exception as exc:
            self.connection.rollback()
            raise RagError(
                f"Could not store document {filename!r} in PostgreSQL: {exc}"
            ) from exc

    def search(
        self, embedding: Sequence[float], limit: int
    ) -> list[tuple[str, str]]:
        try:
            with self.connection.cursor() as cursor:
                cursor.execute(
                    """
                    SELECT source_filename, content
                    FROM rag_chunks
                    ORDER BY embedding <=> %s
                    LIMIT %s
                    """,
                    (Vector(embedding), limit),
                )
                return [(row[0], row[1]) for row in cursor.fetchall()]
        except Exception as exc:
            raise RagError(f"Could not retrieve relevant document chunks: {exc}") from exc
