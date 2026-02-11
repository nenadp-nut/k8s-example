package com.example.javamongo.document;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/docs")
public class DocumentController {

    @Value("${app.name:java-mongo-service}")
    private String appName;

    private final DocumentRepository repository;

    public DocumentController(DocumentRepository repository) {
        this.repository = repository;
    }

    @GetMapping
    public Map<String, Object> list() {
        List<Document> docs = repository.findTop20ByOrderByIdDesc();
        Map<String, Object> result = new HashMap<>();
        result.put("documents", docs);
        result.put("from", appName);
        return result;
    }

    @PostMapping
    public ResponseEntity<Map<String, Object>> create(@RequestBody Map<String, Object> body) {
        Document doc = new Document();
        doc.setData(body);
        doc = repository.save(doc);
        Map<String, Object> result = new HashMap<>();
        result.put("id", doc.getId());
        result.put("data", doc.getData());
        result.put("from", appName);
        return ResponseEntity.status(201).body(result);
    }

    @GetMapping("/{id}")
    public ResponseEntity<?> get(@PathVariable String id) {
        return repository.findById(id)
                .map(doc -> ResponseEntity.ok((Object) doc))
                .orElse(ResponseEntity.notFound().build());
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Map<String, Object>> delete(@PathVariable String id) {
        if (!repository.existsById(id)) {
            return ResponseEntity.notFound().build();
        }
        repository.deleteById(id);
        Map<String, Object> result = new HashMap<>();
        result.put("id", id);
        result.put("deleted", true);
        result.put("from", appName);
        return ResponseEntity.ok(result);
    }
}
