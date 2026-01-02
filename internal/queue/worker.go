package queue

import (
	"context"
	"deploy-platform/internal/build"
	"deploy-platform/internal/database"
	"deploy-platform/internal/models"
	"log"
	"sync"
)

// WorkerPool manages multiple build workers
type WorkerPool struct {
	queue    BuildQueue
	buildSvc *build.Service
	workers  int
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(queue BuildQueue, buildSvc *build.Service, numWorkers int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		queue:    queue,
		buildSvc: buildSvc,
		workers:  numWorkers,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start starts all workers
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
	log.Printf("✅ Started %d build workers", wp.workers)
}

// Stop stops all workers
func (wp *WorkerPool) Stop() {
	wp.cancel()
	wp.wg.Wait()
	log.Println("🛑 All workers stopped")
}

func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	log.Printf("Worker %d started", id)

	for {
		select {
		case <-wp.ctx.Done():
			log.Printf("Worker %d stopping", id)
			return
		default:
			deploymentID, err := wp.queue.Dequeue(wp.ctx)
			if err != nil {
				if err == context.Canceled {
					return
				}
				log.Printf("Worker %d: Error dequeuing: %v", id, err)
				continue
			}

			log.Printf("Worker %d: Processing deployment %d", id, deploymentID)
			if err := wp.buildSvc.BuildDeployment(wp.ctx, deploymentID); err != nil {
				log.Printf("Worker %d: Build failed for deployment %d: %v", id, deploymentID, err)
				// Update deployment status
				database.DB.Model(&models.Deployment{}).Where("id = ?", deploymentID).Update("status", "failed")
			} else {
				log.Printf("Worker %d: Build completed for deployment %d", id, deploymentID)
			}
		}
	}
}
