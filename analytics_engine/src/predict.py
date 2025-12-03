import argparse
from .predictor import Predictor
import torch


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
    for t, (pred, prob) in zip(texts, results):
        print(f"TEXT: {t}")
        print(f"PRED: {pred}  PROBS: {prob}")
        print("-" * 40)


if __name__ == "__main__":
    main()
