from __future__ import annotations

import os
from dataclasses import dataclass
from pathlib import Path

from .errors import RagError


PROJECT_ROOT = Path(__file__).resolve().parents[3]
DOCUMENTS_DIR = PROJECT_ROOT / "knowledge" / "documents"
EXTRACTED_DIR = PROJECT_ROOT / "knowledge" / "extracted"
VECTOR_DIMENSIONS = 768


@dataclass(frozen=True)
class Settings:
    gemini_api_key: str
    embedding_model: str
    answer_model: str
    postgres_host: str
    postgres_port: int
    postgres_user: str
    postgres_password: str
    postgres_database: str
    postgres_sslmode: str

    @classmethod
    def from_env(cls) -> "Settings":
        required = {
            "GEMINI_API_KEY": os.getenv("GEMINI_API_KEY"),
            "PSQL_HOST": os.getenv("PSQL_HOST"),
            "PSQL_PORT": os.getenv("PSQL_PORT"),
            "PSQL_USER": os.getenv("PSQL_USER"),
            "PSQL_PASS": os.getenv("PSQL_PASS"),
            "PSQL_DBNAME": os.getenv("PSQL_DBNAME"),
            "PSQL_SSLMODE": os.getenv("PSQL_SSLMODE"),
        }
        missing = [name for name, value in required.items() if not value]
        if missing:
            raise RagError(
                "Missing required environment variable(s): "
                + ", ".join(missing)
                + ". Set them in your shell before running pdf-rag; "
                "no credentials or database defaults are supplied."
            )

        try:
            port = int(required["PSQL_PORT"] or "")
        except ValueError as exc:
            raise RagError("PSQL_PORT must be an integer.") from exc
        if not 1 <= port <= 65535:
            raise RagError("PSQL_PORT must be between 1 and 65535.")

        return cls(
            gemini_api_key=required["GEMINI_API_KEY"] or "",
            embedding_model=os.getenv(
                "GEMINI_EMBEDDING_MODEL", "gemini-embedding-001"
            ),
            answer_model=os.getenv("GEMINI_ANSWER_MODEL", "gemini-2.5-flash"),
            postgres_host=required["PSQL_HOST"] or "",
            postgres_port=port,
            postgres_user=required["PSQL_USER"] or "",
            postgres_password=required["PSQL_PASS"] or "",
            postgres_database=required["PSQL_DBNAME"] or "",
            postgres_sslmode=required["PSQL_SSLMODE"] or "",
        )

    def postgres_connection_kwargs(self) -> dict[str, str | int]:
        return {
            "host": self.postgres_host,
            "port": self.postgres_port,
            "user": self.postgres_user,
            "password": self.postgres_password,
            "dbname": self.postgres_database,
            "sslmode": self.postgres_sslmode,
        }
