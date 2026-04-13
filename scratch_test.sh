#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"

echo "Testing Level 1 (Beginner) Path..."
curl -s -X POST "$BASE_URL/engine/play" \
  -H "Content-Type: application/json" \
  -d '{
    "fen": "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
    "level": 1,
    "movetime": 1000
  }' | jq .

echo -e "\nTesting Level 6 (Pro) Path..."
curl -s -X POST "$BASE_URL/engine/play" \
  -H "Content-Type: application/json" \
  -d '{
    "fen": "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
    "level": 6,
    "movetime": 2000
  }' | jq .

echo -e "\nTesting Manual ELO Path..."
curl -s -X POST "$BASE_URL/engine/play" \
  -H "Content-Type: application/json" \
  -d '{
    "fen": "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
    "elo": 2000,
    "movetime": 1000
  }' | jq .
