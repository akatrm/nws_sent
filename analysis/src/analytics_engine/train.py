"""Train a pretrained transformer model for sentiment classification.

This script is a convenience entrypoint that fine-tunes a
Hugging Face transformer for binary/multi-class sentiment tasks. It
supports loading CSV training files (via `datasets` or a local pandas
fallback), tokenizes examples, and runs `transformers.Trainer` with a
standard set of training arguments.

This file is intended as a cli/utility rather than an importable
library; for programmatic training prefer directly constructing a
`Trainer` with the provided preprocessing steps.
"""

import argparse
import logging
from datasets import load_dataset
from transformers import (
    AutoTokenizer,
    AutoModelForSequenceClassification,
    TrainingArguments,
    Trainer,
    DataCollatorWithPadding,
)
import numpy as np

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


def parse_args():
    p = argparse.ArgumentParser(description="Fine-tune a pretrained model for sentiment")
    p.add_argument("--model", default="distilbert-base-uncased", help="pretrained model")
    p.add_argument("--output_dir", default="./outputs", help="where to save the model")
    p.add_argument("--epochs", type=int, default=3)
    p.add_argument("--batch_size", type=int, default=16)
    p.add_argument("--max_length", type=int, default=128)
    p.add_argument("--train_file", default=None, help="optional CSV train file")
    p.add_argument("--validation_file", default=None, help="optional CSV validation file")
    p.add_argument("--local_only", action="store_true", help="If set, load CSVs only via pandas locally (no datasets CSV loader)")
    p.add_argument("--delimiter", default=",", help="CSV delimiter when using local_only or diagnostics")
    p.add_argument("--encoding", default="utf-8", help="File encoding for local CSVs")
    p.add_argument("--seed", type=int, default=42)
    return p.parse_args()


def main():
    args = parse_args()

    # Load dataset: CSV input or fallback to glue/sst2
    if args.train_file:
        data_files = {}
        data_files["train"] = args.train_file
        if args.validation_file:
            data_files["validation"] = args.validation_file

        if args.local_only:
            # Read CSVs locally with pandas and construct a datasets DatasetDict
            try:
                import pandas as pd
                from datasets import Dataset, DatasetDict

                dfs = {}
                for split_name, path in data_files.items():
                    logger.info(f"[local_only] Reading {split_name} from {path} with encoding={args.encoding} delimiter={args.delimiter}")
                    df = pd.read_csv(path, delimiter=args.delimiter, encoding=args.encoding)
                    logger.info(f"[local_only] {path}: shape={df.shape} columns={list(df.columns)}")
                    # Drop unnamed index column if pandas added it from saved DataFrame
                    unnamed = [c for c in df.columns if c.startswith("Unnamed")]
                    if unnamed:
                        df = df.drop(columns=unnamed)
                    dfs[split_name] = Dataset.from_pandas(df.reset_index(drop=True))

                # Ensure we have a DatasetDict with at least 'train'
                if len(dfs) == 1 and "train" in dfs:
                    from datasets import DatasetDict

                    ds = DatasetDict({"train": dfs["train"]})
                else:
                    ds = DatasetDict(dfs)

            except Exception:
                logger.exception("Failed to read CSVs using pandas in local_only mode")
                raise
        else:
            # Try to load CSV via the datasets library. If it fails, print
            # helpful diagnostics using pandas so the root cause (encoding,
            # parsing) is easier to see.
            try:
                ds = load_dataset("csv", data_files=data_files, delimiter=args.delimiter)
            except Exception:
                logger.exception("datasets.load_dataset failed for CSV - falling back to pandas diagnostics")
                try:
                    import pandas as pd
                    # print diagnostics for each provided file
                    for split_name, path in data_files.items():
                        try:
                            logger.info(f"Reading {split_name} file with pandas: {path}")
                            df = pd.read_csv(path, delimiter=args.delimiter, encoding=args.encoding)
                            logger.info(f"Pandas read {path}: shape={df.shape}")
                            logger.info(f"Columns: {list(df.columns)}")
                            logger.info(f"Dtypes:\n{df.dtypes}")
                            logger.info(f"First 5 rows:\n{df.head().to_dict(orient='records')}")
                        except Exception as e2:
                            logger.exception(f"Failed to read {path} with pandas: {e2}")
                except Exception:
                    logger.exception("Pandas diagnostics also failed or pandas is not installed")
                # Re-raise original exception so user sees the DatasetGenerationError stack
                raise

        # Expect columns: "text" and "label" (0/1)
        text_col = "text"
        label_col = "label"
    else:
        ds = load_dataset("glue", "sst2")
        text_col = "sentence"
        label_col = "label"

    tokenizer = AutoTokenizer.from_pretrained(args.model, use_fast=True)

    def preprocess(batch):
        return tokenizer(batch[text_col], truncation=True, padding="max_length", max_length=args.max_length)

    ds = ds.map(preprocess, batched=True)
    ds = ds.rename_column(label_col, "labels")
    ds.set_format(type="torch", columns=["input_ids", "attention_mask", "labels"]) 

    # Convert labels column to a Python list in a safe way. Some Dataset
    # Column objects don't implement `tolist()`, so fall back to `list()`.
    labels_col = ds["train"]["labels"]
    try:
        labels_list = labels_col.tolist()
    except Exception:
        labels_list = list(labels_col)
    num_labels = len(set(labels_list))
    model = AutoModelForSequenceClassification.from_pretrained(args.model, num_labels=num_labels)

    # `datasets.load_metric` has been moved to the `evaluate` library.
    # Use evaluate.load(...) instead for metrics.
    try:
        import evaluate
    except Exception:
        logger.exception("The 'evaluate' package is required for metrics. Please install with 'pip install evaluate'.")
        raise

    metric_acc = evaluate.load("accuracy")
    metric_f1 = evaluate.load("f1")

    def compute_metrics(pred):
        labels = pred.label_ids
        preds = np.argmax(pred.predictions, axis=1)
        acc = metric_acc.compute(predictions=preds, references=labels)
        f1 = metric_f1.compute(predictions=preds, references=labels, average="weighted")
        return {"accuracy": acc["accuracy"], "f1": f1["f1"]}

    training_args = TrainingArguments(
        output_dir=args.output_dir,
        eval_strategy="epoch",
        save_strategy="epoch",
        num_train_epochs=args.epochs,
        per_device_train_batch_size=args.batch_size,
        per_device_eval_batch_size=args.batch_size,
        logging_steps=50,
        load_best_model_at_end=True,
        metric_for_best_model="f1",
        seed=args.seed,
        fp16=False,
    )

    data_collator = DataCollatorWithPadding(tokenizer)

    trainer = Trainer(
        model=model,
        args=training_args,
        train_dataset=ds["train"],
        eval_dataset=ds.get("validation"),
        data_collator=data_collator,
        compute_metrics=compute_metrics,
    )

    trainer.train()
    trainer.save_model(args.output_dir)
    tokenizer.save_pretrained(args.output_dir)


if __name__ == "__main__":
    main()
