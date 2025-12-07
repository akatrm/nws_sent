"""Inference utilities for the analytics engine.

This module exposes a small `Predictor` class that wraps a
Hugging Face tokenizer and sequence-classification model for batched
inference. It is designed to be lightweight and to keep model loading
concise for use from a long-running server process.

Public API:
- Predictor(model_dir, device=None): load a tokenizer+model from
  `model_dir` and call `predict(texts)` to obtain label predictions and
  probabilities.
"""

from typing import List, Optional
import torch
import logging
from transformers import AutoTokenizer, AutoModelForSequenceClassification

logger = logging.getLogger(__name__)


class Predictor:
    """Wrapper around tokenizer+model for inference.

    Usage:
        p = Predictor(model_dir='outputs')
        results = p.predict(['text1','text2'])
    """

    def __init__(self, model_dir: str, device: Optional[str] = None):
        self.model_dir = model_dir
        self.device = device or ("cuda" if torch.cuda.is_available() else "cpu")
        self.tokenizer = AutoTokenizer.from_pretrained(model_dir)
        self.model = AutoModelForSequenceClassification.from_pretrained(model_dir)
        self.model.to(self.device)
        self.model.eval()

    def predict(self, texts: List[str]):
        if not texts:
            return []
        enc = self.tokenizer(texts, truncation=True, padding=True, return_tensors="pt")
        input_ids = enc["input_ids"].to(self.device)
        attention_mask = enc["attention_mask"].to(self.device)
        with torch.no_grad():
            logits = self.model(input_ids=input_ids, attention_mask=attention_mask).logits
            probs = torch.softmax(logits, dim=-1).cpu().numpy()
            preds = probs.argmax(axis=1).tolist()

        results = []
        for t, p, pred in zip(texts, probs.tolist(), preds):
            results.append({"text": t, "pred": int(pred), "probs": p})

        return results
