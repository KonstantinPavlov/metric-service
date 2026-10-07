for i in {1..50000}; do
  curl -s -X POST http://localhost:8080/update/ \
    -H "Content-Type: application/json" \
    -d "{\"id\":\"testMetric_$i\",\"type\":\"counter\",\"delta\":$i}" \
    > /dev/null
done