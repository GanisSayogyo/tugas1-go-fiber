
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const endpoint = "http://localhost:3000/api/v1/supervisors/1/assign"

type result struct {
	studentID int
	status    int
	body      string
	err       error
}

func main() {
	token := os.Getenv("ADMIN_TOKEN")
	if token == "" {
		fmt.Println("ADMIN_TOKEN belum diatur")
		os.Exit(1)
	}

	rawIDs := os.Getenv("STUDENT_IDS")
	parts := strings.Split(rawIDs, ",")

	if len(parts) != 30 {
		fmt.Println("STUDENT_IDS harus berisi tepat 30 ID mahasiswa")
		os.Exit(1)
	}

	ids := make([]int, 0, 30)
	for _, part := range parts {
		id, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || id <= 0 {
			fmt.Println("ID mahasiswa tidak valid:", part)
			os.Exit(1)
		}
		ids = append(ids, id)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	start := make(chan struct{})
	results := make(chan result, len(ids))

	var wg sync.WaitGroup
	var success atomic.Int32
	var conflict atomic.Int32
	var other atomic.Int32

	for _, studentID := range ids {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()
			<-start

			payload, _ := json.Marshal(map[string]int{
				"student_id": id,
			})

			req, err := http.NewRequest(
				http.MethodPost,
				endpoint,
				bytes.NewReader(payload),
			)
			if err != nil {
				results <- result{studentID: id, err: err}
				return
			}

			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				results <- result{studentID: id, err: err}
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)

			switch resp.StatusCode {
			case http.StatusCreated:
				success.Add(1)
			case http.StatusConflict:
				conflict.Add(1)
			default:
				other.Add(1)
			}

			results <- result{
				studentID: id,
				status:    resp.StatusCode,
				body:      string(body),
			}
		}(studentID)
	}

	startTime := time.Now()
	close(start)
	wg.Wait()
	close(results)

	fmt.Println("\n=== CONCURRENCY TEST ===")
	for r := range results {
		if r.err != nil {
			fmt.Printf("Student %d ERROR: %v\n", r.studentID, r.err)
			continue
		}

		fmt.Printf("Student %d -> HTTP %d\n", r.studentID, r.status)
		if r.status != http.StatusCreated &&
			r.status != http.StatusConflict {
			fmt.Println("Response:", r.body)
		}
	}

	fmt.Println("\n=== SUMMARY ===")
	fmt.Println("Total request :", len(ids))
	fmt.Println("Success       :", success.Load())
	fmt.Println("Conflict      :", conflict.Load())
	fmt.Println("Other         :", other.Load())
	fmt.Println("Duration      :", time.Since(startTime))
}
