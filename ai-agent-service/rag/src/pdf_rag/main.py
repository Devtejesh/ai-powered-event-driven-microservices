from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from .config import Settings
from .errors import RagError
from .pipeline import answer_instruction
from .providers import GeminiProvider
from .store import PostgresStore


@asynccontextmanager
async def lifespan(app: FastAPI):
    settings = Settings.from_env()
    store = PostgresStore.connect(settings)

    try:
        store.ensure_schema()
        app.state.store = store
        app.state.provider = GeminiProvider(settings)
        yield
    finally:
        store.close()


app = FastAPI(
    title="Go Commerce RAG API",
    description="Ask questions about indexed PDF documents.",
    lifespan=lifespan,
)


class AskRequest(BaseModel):
    question: str = Field(min_length=1, max_length=5000)
    top_k: int = Field(default=5, ge=1, le=20)


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/ask")
def ask(request: AskRequest):
    try:
        answer = answer_instruction(
            request.question,
            app.state.store,
            app.state.provider,
            top_k=request.top_k,
        )
        return {"answer": answer}
    except RagError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc