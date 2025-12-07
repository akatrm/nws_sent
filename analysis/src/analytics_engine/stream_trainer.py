"""Streaming training utilities.

This module implements `StreamTrainer`, a compact asynchronous consumer
that accepts JSON payloads (lists of examples) and performs incremental
training steps using `ModelTrainer`. It is not a full-featured
replacement for batch training pipelines but is useful for demoing
online/stream ingestion behavior.
"""

import asyncio
import logging
from typing import Optional, List

from .model_trainer import ModelTrainer

logger = logging.getLogger(__name__)


class StreamTrainer:
        """A minimal online trainer that consumes text,label CSV lines from an
        async queue and performs incremental training steps.

        Notes:
        - This is intentionally simple and meant for streaming demos. It does
            not replace a full Hugging Face `Trainer` for production.
        - Input lines should be CSV rows: text,label (label integer)
        """

    def __init__(self, 
        model_name: str = "distilbert-base-uncased", 
        model_dir: Optional[str] = None, 
        device: Optional[str] = None, 
        lr: float = 5e-5, 
        batch_size: int = 8, 
        max_length: int = 128):

        self.model_name = model_name
        # use a private backing field for model_dir so we can expose a read-only property
        self._model_dir = model_dir
        self.device = device
        self.lr = lr
        self.batch_size = batch_size
        self.max_length = max_length
        self.queue: asyncio.Queue = asyncio.Queue()
        self._task: Optional[asyncio.Task] = None
        self._running = False

    @property
    def model_dir(self):
        return self._model_dir

    async def start(self):
        if self._running:
            return
        self._running = True
        # delegate to ModelTrainer which will handle device selection. Prefer loading
        # from `model_dir` when provided (local checkpoint) otherwise use `model_name`.
        self.model_trainer = ModelTrainer(model_name=self.model_name, model_dir=self.model_dir, device=self.device, lr=self.lr, max_length=self.max_length)
        self._task = asyncio.create_task(self._consumer_loop())
        logger.info("StreamTrainer started")

    async def stop(self):
        if not self._running:
            return
        # attempt to save the model via ModelTrainer before shutting down
        try:
            self.model_trainer.save()
        finally:
            self._running = False
            # wait for queue to drain a little and cancel task
            if self._task:
                await self._task
                self._task = None
            self.model_trainer = None
            logger.info("StreamTrainer stopped")

    async def enqueue_line(self, json: dict):
        """Enqueue a single JSON payload containing examples: {"examples": [{"text":...,"label":...}, ...]}"""
        data: List[(str, int)] = []
        for l in json.get("examples", []):
            data.append((l["text"], l["label"]))

        await self.queue.put(data)

    async def _consumer_loop(self):
        buffer: List[(str, int)] = []
        while self._running:
            try:
                line = await asyncio.wait_for(self.queue.get(), timeout=1.0)
            except asyncio.TimeoutError:
                line = None

            if line:
                buffer.extend(line)

            # If we have enough for a batch, run a train step
            if len(buffer) >= self.batch_size:
                batch_lines = buffer[: self.batch_size]
                buffer = buffer[self.batch_size :]
                try:
                    self._train_step(batch_lines)
                except Exception:
                    logger.exception("Error during train step")

        # drain remaining buffer once stopped
        if buffer:
            try:
                self._train_step(buffer)
            except Exception:
                logger.exception("Error during final train step")


    def _train_step(self, lines: List[(str, int)]):
        if not lines:
            return

        loss = self.model_trainer.train_batch(lines)
        if loss is not None:
            logger.info(f"Trained batch size={len(lines)} loss={loss:.4f}")
