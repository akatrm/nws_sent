import logging
from typing import List, Optional
from pathlib import Path

import torch
import torch.nn as nn
from torch.optim import AdamW
from transformers import AutoTokenizer, AutoModelForSequenceClassification

logger = logging.getLogger(__name__)


class ModelTrainer:
    """Encapsulates model/tokenizer initialization and a simple training step API.

    Responsibilities:
    - load tokenizer and model
    - provide train_batch(lines: List[str]) to train on CSV lines (text,label)
    - save model and tokenizer
    """

    def __init__(self, model_name: str = "distilbert-base-uncased", model_dir: Optional[str] = None, device: Optional[str] = None, lr: float = 5e-5, max_length: int = 128):
        """Create a ModelTrainer.

        If `model_dir` is provided it will be used to load a saved model/tokenizer
        instead of retrieving a pretrained model by name. This allows initializing
        from checkpoints on disk.
        """
        self.model_name = model_name
        self.model_dir = model_dir
        self.device = device or ("cuda" if torch.cuda.is_available() else "cpu")
        self.lr = lr
        self.max_length = max_length

        # prefer a provided model_dir only if it exists on disk; otherwise fall back
        # to the model_name (which will be resolved by Hugging Face hub)
        if model_dir:
            p = Path(model_dir)
            if p.exists() and p.is_dir():
                src = model_dir
            else:
                src = model_name
        else:
            src = model_name
        self.tokenizer = AutoTokenizer.from_pretrained(src, use_fast=True)
        self.model = AutoModelForSequenceClassification.from_pretrained(src)
        self.model.to(self.device)

        self.optimizer = AdamW(self.model.parameters(), lr=self.lr)
        self.loss_fn = nn.CrossEntropyLoss()

    def _parse_lines(self, lines: List[(str, str)]):
        texts = []
        labels = []
        for l in lines:
            l = l.strip()
            if not l:
                continue
            if "," in l:
                text, label = l.rsplit(",", 1)
                text = text.strip().strip('"')
                try:
                    label = int(label.strip())
                except Exception:
                    continue
                texts.append(text)
                labels.append(label)
            else:
                continue
        return texts, labels

    def train_batch(self, lines: List[str]):
        """Perform a training step on the provided CSV lines.

        Returns the loss value or None if no valid examples.
        """
        texts, labels = self._parse_lines(lines)
        if not texts:
            return None

        enc = self.tokenizer(texts, truncation=True, padding=True, max_length=self.max_length, return_tensors="pt")
        input_ids = enc["input_ids"].to(self.device)
        attention_mask = enc["attention_mask"].to(self.device)
        labels_t = torch.tensor(labels, dtype=torch.long, device=self.device)

        self.model.train()
        outputs = self.model(input_ids=input_ids, attention_mask=attention_mask)
        logits = outputs.logits
        loss = self.loss_fn(logits, labels_t)

        self.optimizer.zero_grad()
        loss.backward()
        self.optimizer.step()

        logger.debug(f"ModelTrainer trained batch size={len(labels)} loss={loss.item():.4f}")
        return float(loss.item())

    def save(self):
        """Save model and tokenizer to out_dir.

        If out_dir is not provided, `self.model_dir` will be used if set. Raises
        ValueError if no target directory is available.
        """
        self.model.save_pretrained(self.model_dir)
        self.tokenizer.save_pretrained(self.model_dir)
        logger.info(f"Saved model and tokenizer to {self.model_dir}")
