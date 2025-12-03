# Sentiment Analysis (Transformer fine-tune)

This small project fine-tunes a pre-trained Transformer for sentiment analysis using Hugging Face Transformers and Datasets.

Files:
- `requirements.txt` - Python dependencies
- `src/train.py` - Training script (fine-tune a model)
- `src/predict.py` - Inference script
- `data/sample.csv` - Small example dataset

Quick start (macOS / zsh):

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt

# Train on sample data
python src/train.py --model distilbert-base-uncased --train_file data/sample.csv --validation_file data/sample.csv --output_dir outputs --epochs 1 --batch_size 8

# Predict
python src/predict.py --model_dir outputs --text "I love this product!"
```

CSV format: header `text,label` where label is `0` (negative) or `1` (positive).

Next steps / ideas:
- Use a larger dataset (SST-2, IMDB)
- Add evaluation & hyperparameter tuning
- Export to ONNX or TorchScript for faster inference
- Add FastAPI wrapper (a starter included in requirements)

API (FastAPI) usage
-------------------

The project includes a FastAPI server at `src/server.py` exposing streaming training and prediction endpoints.

Run the server (zsh):

```bash
# from project root
uvicorn src.server:app --reload --host 0.0.0.0 --port 8000
```

zsh-friendly curl examples

1) Start the stream trainer

```bash
curl -X POST "http://127.0.0.1:8000/stream/start" \
	-H "Accept: application/json"
```

2) Stop the stream trainer

```bash
curl -X POST "http://127.0.0.1:8000/stream/stop" \
	-H "Accept: application/json"
```

3) Stream training data (newline-delimited CSV lines). This endpoint accepts raw streamed bodies.

```bash
curl -N -X POST "http://127.0.0.1:8000/stream/train" \
	-H "Content-Type: text/plain" \
	--data-binary $'text,label\n"I love this product!",1\n"This is terrible",0\n'
```

4) Check stream status

```bash
curl -X GET "http://127.0.0.1:8000/stream/status" \
	-H "Accept: application/json"
```

5) Initialize/load the predictor

```bash
curl -X POST "http://127.0.0.1:8000/predictor/start" \
	-H "Accept: application/json"
```

6) Unload the predictor

```bash
curl -X POST "http://127.0.0.1:8000/predictor/stop" \
	-H "Accept: application/json"
```

7) Make a prediction (single text)

```bash
curl -X POST "http://127.0.0.1:8000/predict" \
	-H "Content-Type: application/json" \
	-d '{"text":"I rarely use this product!"}'
```

8) Make a prediction (multiple texts)

```bash
curl -X POST "http://127.0.0.1:8000/predict" \
	-H "Content-Type: application/json" \
	-d '{"texts":["I love it!","Not what I expected."]}'
```

Notes
- If the stream trainer is running, `/predict` returns HTTP 409 (training in progress). Stop the trainer to allow predictions.
- The `/stream/train` endpoint expects raw newline-delimited CSV lines; use `--data-binary` to POST multiple lines with `curl`.
