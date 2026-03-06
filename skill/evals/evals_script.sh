for model in large2 kimik25 minimaxm25 groq-llama4 sonnet; do
  echo "=========================================="
  echo "Running evals for: $model"
  echo "=========================================="
  python skill/evals/eval_runner.py \
    --skill skill/splitwise/SKILL.md \
    --evals skill/evals/evals.json \
    --model $model \
    --run-stage2-always \
    --show-responses \
    --output skill/evals/evals_outputs/results-$model.json
done