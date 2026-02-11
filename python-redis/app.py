"""
Python service using Redis cache.
Exposes HTTP API for cache get/set and is called by the Go service.
"""
import os
import socket
from flask import Flask, request, jsonify
import redis

app = Flask(__name__)
APP_NAME = os.environ.get("APP_NAME", "python-redis-service")
CACHE_TTL = int(os.environ.get("CACHE_TTL_SECONDS", "300"))

# Redis connection from env (injected via ConfigMap/Secret in K8s)
redis_host = os.environ.get("REDIS_HOST", "redis-service")
redis_port = int(os.environ.get("REDIS_PORT", "6379"))
redis_client = redis.Redis(host=redis_host, port=redis_port, decode_responses=True)


@app.route("/health")
def health():
    try:
        redis_client.ping()
        redis_status = "connected"
    except Exception as e:
        redis_status = str(e)
    return jsonify({
        "status": "ok",
        "service": APP_NAME,
        "hostname": socket.gethostname(),
        "redis": redis_status,
    })


@app.route("/cache/<key>", methods=["GET"])
def get_cache(key):
    value = redis_client.get(key)
    if value is None:
        return jsonify({"key": key, "value": None, "from": APP_NAME}), 404
    return jsonify({"key": key, "value": value, "from": APP_NAME})


@app.route("/cache/<key>", methods=["PUT", "POST"])
def set_cache(key):
    data = request.get_json(force=True, silent=True) or {}
    value = data.get("value", request.get_data(as_text=True) or "")
    ttl = data.get("ttl", CACHE_TTL)
    redis_client.setex(key, ttl, value)
    return jsonify({"key": key, "value": value, "ttl": ttl, "from": APP_NAME})


@app.route("/cache/<key>", methods=["DELETE"])
def delete_cache(key):
    redis_client.delete(key)
    return jsonify({"key": key, "deleted": True, "from": APP_NAME})


if __name__ == "__main__":
    port = int(os.environ.get("PORT", "8080"))
    app.run(host="0.0.0.0", port=port)
