"""Small CLI wrapper for running predictions from a saved model.

This module provides a very small command-line helper to load a
`Predictor` from a `model_dir` and print predictions for either a
single `--text` or a file with one text per line via `--file`.

Example:
    python -m analysis.src.analytics_engine.predict --model_dir ./outputs --text "I love this"
"""

import argparse
from .predictor import Predictor


def parse_args():
    p = argparse.ArgumentParser()
    p.add_argument("--model_dir", required=True, help="directory with saved model/tokenizer")
    p.add_argument("--text", required=False, help="single input text")
    p.add_argument("--file", required=False, help="file with one text per line")
    return p.parse_args()


def predict_texts(model_dir, texts):
    p = Predictor(model_dir)
    return p.predict(texts)


def main():
    args = parse_args()
    texts = []
    if args.text:
        texts = [args.text]
    elif args.file:
        with open(args.file, "r") as f:
            texts = [l.strip() for l in f if l.strip()]
    else:
        print("Provide --text or --file")
        return
    results = predict_texts(args.model_dir, texts)
    for r in results:
        print(r)


if __name__ == "__main__":
    main()
