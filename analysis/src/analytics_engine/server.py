from fastapi import FastAPI, Request, Form
import logging
from typing import Optional
import os

from .stream_trainer import StreamTrainer
from .predictor import Predictor
from fastapi import HTTPException
from pydantic import BaseModel
from typing import List
from pathlib import Path
import json

def load_server_config() -> dict:
    """Load repository-level config.json if present and return as dict."""
    repo_root = os.path.dirname(os.path.dirname(__file__))
    cfg_path = os.path.join(repo_root, "config.json")
    p = Path(cfg_path)
    if not p.exists():
        return {}
    try:
        return json.loads(p.read_text(encoding="utf-8"))
    except Exception:
        return {}

logger = logging.getLogger(__name__)
app = FastAPI()

class PredictRequest(BaseModel):
    text: Optional[str] = None
    texts: Optional[List[str]] = None


class ServerManager:
    """Encapsulate server state: manages a single StreamTrainer and cached Predictors.

    - Lazily creates a StreamTrainer when start_trainer is called
    - Caches Predictor instances per model_dir to avoid repeated loads
    - Provides thin wrappers used by FastAPI routes
    """

    def __init__(self):
        self._stream_trainer: Optional[StreamTrainer] = None
        # single Predictor instance (only one supported at a time)
        self._predictor: Optional[Predictor] = None
        # default model_dir used when none provided
        self.model_dir =  os.path.dirname(os.path.dirname(os.path.abspath(__file__))) + "/outputs"
        # currently loaded predictor model dir
        self._predictor_model_dir: Optional[str] = None

    async def start_trainer(self):
        if self._stream_trainer is not None:
            return {"status": "already_running"}
        if self._stream_trainer is None:
            self._stream_trainer = self.init_stream_trainer()
        await self._stream_trainer.start()
        return {"status": "started"}

    def init_stream_trainer(self, model: str = "distilbert-base-uncased", batch_size: int = 8, model_dir: Optional[str] = None):
        cfg = load_server_config()
        md = model_dir or cfg.get("model_dir") or self.model_dir
        bs = batch_size or cfg.get("batch_size", 8)
        lr = cfg.get("lr", 5e-5)
        max_length = cfg.get("max_length", 128)
        return  StreamTrainer(model_name=model, model_dir=md, batch_size=bs, device=None, lr=lr, max_length=max_length)
        
    async def stop_trainer(self):
        if self._stream_trainer is None:
            return {"status": "not_running"}
        await self._stream_trainer.stop()
        self._stream_trainer = None
        return {"status": "stopped"}

    async def ingest_stream(self, request: Request):
        if self._stream_trainer is None:
            return {"error": "trainer_not_started"}

        received_data = []
        async for chunk in request.stream():
            if not chunk:
                continue
            text = chunk.decode("utf-8")
            received_data.append(text)
        complete_data = "".join(received_data)
        json_data = json.loads(complete_data)

        self._stream_trainer.enqueue_line(json_data)

        return {"status": "accepted"}

    def is_running(self) -> bool:
        return self._stream_trainer is not None

    def predict(self, req: PredictRequest):
        # If training is currently running, don't serve predictions
        if self.is_running():
            raise HTTPException(status_code=409, detail="Training operation in progress; predictions unavailable")

        # Otherwise, use the single pre-initialized Predictor
        if self._predictor is None:
            raise HTTPException(status_code=400, detail="No predictor initialized. Call /predictor/load to initialize a predictor before requesting predictions.")

        texts = []
        if req.text:
            texts = [req.text]
        elif req.texts:
            texts = req.texts
        else:
            raise HTTPException(status_code=400, detail="Provide 'text' or 'texts' in the request")

        results = self._predictor.predict(texts)
        return {"results": results}

    def start_predict(self, model_dir: Optional[str] = None):
        """Initialize a single Predictor instance from model_dir (or manager default).

        Returns a status dict.
        """
        if self._predictor is not None:
            return {"status": "already_loaded", "model_dir": self._predictor_model_dir}
        md = model_dir or self.model_dir
        self._predictor = Predictor(md)
        self._predictor_model_dir = md
        return {"status": "loaded", "model_dir": md}

    def stop_predict(self):
        """Unload the currently cached Predictor. If model_dir provided, validate it matches."""
        if self._predictor is None:
            return {"status": "not_loaded"}
        # drop reference to allow GC
        prev = self._predictor_model_dir
        self._predictor = None
        self._predictor_model_dir = None
        return {"status": "unloaded", "model_dir": prev}

    def get_predictor_model_dir(self) -> Optional[str]:
        return self._predictor_model_dir

    # convenience aliases requested by user
    def start_predictor(self, model_dir: Optional[str] = None):
        return self.start_predict(model_dir=model_dir)

    def stop_predictor(self, model_dir: Optional[str] = None):
        return self.stop_predict(model_dir=model_dir)


# single global server manager used by route handlers
_server_manager = ServerManager()


@app.post("/stream/start")
async def stream_start():
    return await _server_manager.start_trainer()


@app.post("/stream/stop")
async def stream_stop():
    return await _server_manager.stop_trainer()


@app.get("/stream/status")
async def stream_status():
    return {"running": _server_manager.is_running()}


@app.post("/stream/train")
async def stream_train(request: Request):
    return await _server_manager.ingest_stream(request)


@app.post("/predict")
async def predict(req: PredictRequest):
    return _server_manager.predict(req)


@app.post('/predictor/start')
async def predictor_load():
    """Initialize and cache a single Predictor from model_dir."""
    return _server_manager.start_predict()


@app.post('/predictor/stop')
async def predictor_unload():
    """Unload the currently cached Predictor. If model_dir provided, validate it matches."""
    return _server_manager.stop_predict()
