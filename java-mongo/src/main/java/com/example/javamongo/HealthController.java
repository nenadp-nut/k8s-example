package com.example.javamongo;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.data.mongodb.core.MongoTemplate;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.net.InetAddress;
import java.util.HashMap;
import java.util.Map;

@RestController
public class HealthController {

    @Value("${app.name:java-mongo-service}")
    private String appName;

    private final MongoTemplate mongoTemplate;

    public HealthController(MongoTemplate mongoTemplate) {
        this.mongoTemplate = mongoTemplate;
    }

    @GetMapping("/health")
    public ResponseEntity<Map<String, Object>> health() {
        Map<String, Object> body = new HashMap<>();
        body.put("status", "ok");
        body.put("service", appName);
        try {
            body.put("hostname", InetAddress.getLocalHost().getHostName());
        } catch (Exception e) {
            body.put("hostname", "unknown");
        }
        try {
            mongoTemplate.getDb().runCommand(org.bson.Document.parse("{ ping: 1 }"));
            body.put("mongodb", "connected");
        } catch (Exception e) {
            body.put("mongodb", e.getMessage());
        }
        return ResponseEntity.ok(body);
    }
}
