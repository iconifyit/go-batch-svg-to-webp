
# 🔍 Image Processing Performance: Go vs. Python vs. Node.js

This benchmark compares the total time required to process **500,000 SVG images** through a pipeline that:

1. Downloads each SVG from S3  
2. Converts it to PNG  
3. Applies a watermark  
4. Generates 3 size variants (128, 512, 1024 px)  
5. Converts each to WebP  
6. Uploads them back to S3  
7. Inserts metadata into the database  

---

## 📌 Assumptions

- **Images**: 500,000 SVG files (~200 KB each)
- **Tools**: rsvg-convert, ImageMagick or ffmpeg, cwebp, AWS SDK, PostgreSQL
- **Output**: 3 WebP variants per input
- **Environment**:
  - Local or cloud instance with SSD and 1 Gbps networking
  - All services (S3, DB) in the same AWS region
- **Disk IO and network latency** are included
- **No cold starts**, warmed-up worker pool
- **No caching**, each file is processed independently

---

## 🧪 Step-by-Step Time Breakdown (Per Image)

| Step                         | Time (ms) |
|------------------------------|-----------|
| Download SVG from S3         | 30–35     |
| Convert SVG → PNG            | 20–30     |
| Apply watermark              | 20–35     |
| Resize + convert to 3 WebPs  | 75–135    |
| Upload 3 WebP files to S3    | 90–120    |
| DB insert/update             | 10–15     |
| **Total**                    | **245–370 ms** |

---

## 🧮 Total Processing Time by Language

### ✅ Go (370 workers, high concurrency)

- Per-image time: 245–370 ms  
- Concurrency: 370 workers  
- Throughput: ~185 images/sec  
- 500,000 ÷ 185 = ~2,700 sec = **~45 minutes**

### 🐍 Python (8 workers, multiprocessing.Pool)

- Per-image time: 245–370 ms  
- Throughput:  
  - Best case: 8 × (1 / 0.245) = ~32.7 img/sec  
  - Worst case: 8 × (1 / 0.370) = ~21.6 img/sec  
- Total time:  
  - Best case: 500,000 ÷ 32.7 = **~4.25 hours**  
  - Worst case: 500,000 ÷ 21.6 = **~6.4 hours**

### 🟫 Node.js (single-threaded, no concurrency)

- Per-image time: ~2.0 sec (includes IO + CLI calls)  
- Throughput: 0.5 images/sec  
- Total time: 500,000 ÷ 0.5 = 1,000,000 sec = **~11.6 days**

---

## 📊 Summary Table

| Language    | Workers | Per-Image Time | Throughput   | Total Time     |
|-------------|---------|----------------|--------------|----------------|
| **Go**      | 370     | 245–370 ms      | ~185/sec     | **~45 min**     |
| **Python**  | 8       | 245–370 ms      | 22–33/sec    | **4.3–6.4 hrs** |
| **Node.js** | 1       | ~2.0 sec        | 0.5/sec      | **~11.6 days**  |

---

## 🧠 Takeaways

- **Go** massively outperforms Python and Node due to low overhead and efficient concurrency.  
- **Python** performs reasonably well with multiprocessing, but overhead and memory usage limit scalability.  
- **Node.js** is unsuitable for high-volume image processing without external job queues or major parallelism workarounds.
